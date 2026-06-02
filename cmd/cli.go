package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"sing-box-web-panel/internal/config"
	libauth "sing-box-web-panel/internal/lib/auth"
	sqliterepo "sing-box-web-panel/internal/repo/sqlite"
	svcapitoken "sing-box-web-panel/internal/services/apitoken"
	svcnode "sing-box-web-panel/internal/services/node"
	"sing-box-web-panel/internal/services/singbox"

	"gopkg.in/yaml.v3"
)

func runCLI(args []string) error {
	if len(args) == 0 {
		return nil
	}
	switch args[0] {
	case "setting":
		return cliSetting(args[1:])
	case "admin":
		return cliAdmin(args[1:])
	case "api-token":
		return cliAPIToken(args[1:])
	case "node":
		return cliNode(args[1:])
	case "core":
		return cliCore(args[1:])
	case "cert":
		return cliCert(args[1:])
	case "-v", "--version", "version":
		fmt.Println("shilka development")
		return nil
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func cliSetting(args []string) error {
	fs := flag.NewFlagSet("setting", flag.ContinueOnError)
	show := fs.Bool("show", false, "show settings")
	port := fs.Int("port", 0, "set panel port")
	listen := fs.String("listen", "", "set panel listen IP")
	publicURL := fs.String("public-url", "", "set subscription public URL")
	tlsMode := fs.String("tls-mode", "", "set panel TLS mode")
	certFile := fs.String("cert-file", "", "set panel certificate path")
	keyFile := fs.String("key-file", "", "set panel key path")
	domain := fs.String("domain", "", "set ACME/self-signed domain")
	resetTOTP := fs.Bool("reset-totp", false, "disable admin TOTP")
	if err := fs.Parse(args); err != nil {
		return err
	}
	cfg := config.MustLoad()
	if *show {
		return printJSON(map[string]any{
			"http_address": cfg.HTTP.Address,
			"tls_mode":     cfg.TLS.Mode,
			"cert_file":    cfg.TLS.CertFile,
			"key_file":     cfg.TLS.KeyFile,
			"public_url":   cfg.Sub.PublicURL,
			"db_path":      cfg.Database.Path,
		})
	}
	updates := map[string]any{}
	if *port > 0 || *listen != "" {
		host, oldPort, err := net.SplitHostPort(cfg.HTTP.Address)
		if err != nil {
			host, oldPort = "127.0.0.1", "8080"
		}
		if *listen != "" {
			host = *listen
		}
		if *port > 0 {
			oldPort = strconv.Itoa(*port)
		}
		updates["http.address"] = net.JoinHostPort(host, oldPort)
	}
	if *publicURL != "" {
		updates["subscription.public_url"] = strings.TrimRight(*publicURL, "/")
	}
	if *tlsMode != "" {
		updates["tls.mode"] = *tlsMode
	}
	if *certFile != "" {
		updates["tls.cert_file"] = *certFile
	}
	if *keyFile != "" {
		updates["tls.key_file"] = *keyFile
	}
	if *domain != "" {
		updates["tls.acme_domains"] = []any{*domain}
		updates["tls.self_signed_hosts"] = []any{*domain}
	}
	if len(updates) > 0 {
		if err := updateYAMLConfig(configPath(), updates); err != nil {
			return err
		}
		fmt.Println("settings updated")
	}
	if *resetTOTP {
		return resetTOTPForAdmin(cfg)
	}
	return nil
}

func cliAdmin(args []string) error {
	if len(args) == 0 || args[0] != "reset-password" {
		return fmt.Errorf("usage: shilka admin reset-password [-username admin] [-password value]")
	}
	fs := flag.NewFlagSet("admin reset-password", flag.ContinueOnError)
	username := fs.String("username", "admin", "admin username")
	password := fs.String("password", "", "new password")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if *password == "" {
		return fmt.Errorf("-password is required")
	}
	cfg := config.MustLoad()
	db, err := openCLIStorage(cfg)
	if err != nil {
		return err
	}
	defer db.Close()
	repo := sqliterepo.NewAdminRepo(db)
	admin, err := repo.GetByUsername(context.Background(), *username)
	if err != nil {
		return err
	}
	hasher := libauth.NewArgon2Hasher(cfg.Auth.Argon2MemoryKB, cfg.Auth.Argon2Iterations, cfg.Auth.Argon2Parallelism)
	hash, err := hasher.Hash(*password)
	if err != nil {
		return err
	}
	admin.PasswordHash = hash
	if err := repo.Update(context.Background(), admin); err != nil {
		return err
	}
	fmt.Println("admin password updated")
	return nil
}

func cliAPIToken(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: shilka api-token create|list|revoke|enable|disable")
	}
	cfg := config.MustLoad()
	db, err := openCLIStorage(cfg)
	if err != nil {
		return err
	}
	defer db.Close()
	svc := svcapitoken.NewService(sqliterepo.NewAPITokenRepo(db))
	ctx := context.Background()
	switch args[0] {
	case "create":
		fs := flag.NewFlagSet("api-token create", flag.ContinueOnError)
		name := fs.String("name", "node", "token name")
		scopes := fs.String("scopes", "node", "comma-separated scopes")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		created, err := svc.Create(ctx, *name, *scopes)
		if err != nil {
			return err
		}
		return printJSON(map[string]any{"id": created.Token.ID, "name": created.Token.Name, "token": created.Raw})
	case "list":
		tokens, err := svc.List(ctx)
		if err != nil {
			return err
		}
		return printJSON(tokens)
	case "revoke":
		id, err := cliID(args[1:])
		if err != nil {
			return err
		}
		return svc.Delete(ctx, id)
	case "enable", "disable":
		id, err := cliID(args[1:])
		if err != nil {
			return err
		}
		return svc.SetEnabled(ctx, id, args[0] == "enable")
	default:
		return fmt.Errorf("unknown api-token command %q", args[0])
	}
}

func cliNode(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: shilka node add|list|probe|sync|delete")
	}
	cfg := config.MustLoad()
	db, err := openCLIStorage(cfg)
	if err != nil {
		return err
	}
	defer db.Close()
	nodeRepo := sqliterepo.NewNodeRepo(db)
	svc := svcnode.NewService(nodeRepo, sqliterepo.NewInboundRepo(db), sqliterepo.NewClientRepo(db), svcnode.NewHTTPClient())
	ctx := context.Background()
	switch args[0] {
	case "add":
		fs := flag.NewFlagSet("node add", flag.ContinueOnError)
		name := fs.String("name", "", "node name")
		address := fs.String("address", "", "node address")
		port := fs.Int("port", 443, "node port")
		scheme := fs.String("scheme", "https", "http or https")
		basePath := fs.String("base-path", "", "panel base path")
		token := fs.String("token", "", "remote API token")
		allowPrivate := fs.Bool("allow-private", false, "allow private node address")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		n, err := svc.Create(ctx, svcnode.Input{Name: *name, Scheme: *scheme, Address: *address, Port: *port, BasePath: *basePath, APITokenSecret: *token, Enabled: true, AllowPrivateAddress: *allowPrivate})
		if err != nil {
			return err
		}
		return printJSON(n)
	case "list":
		nodes, err := svc.List(ctx)
		if err != nil {
			return err
		}
		return printJSON(nodes)
	case "probe", "sync", "delete":
		id, err := cliID(args[1:])
		if err != nil {
			return err
		}
		if args[0] == "delete" {
			return svc.Delete(ctx, id)
		}
		if args[0] == "probe" {
			n, err := svc.Probe(ctx, id)
			if err != nil {
				return err
			}
			return printJSON(n)
		}
		res, err := svc.Sync(ctx, id)
		if err != nil {
			return err
		}
		return printJSON(res)
	default:
		return fmt.Errorf("unknown node command %q", args[0])
	}
}

func cliCore(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: shilka core start|stop|restart|reload|status|config-check")
	}
	cfg := config.MustLoad()
	pm := singbox.NewProcessManager(singbox.ProcessConfig{
		Mode:         cfg.SingBox.ProcessMode,
		Binary:       cfg.SingBox.BinaryPath,
		ConfigPath:   cfg.SingBox.ConfigPath,
		WorkingDir:   cfg.SingBox.WorkingDir,
		ServiceName:  cfg.SingBox.ServiceName,
		RestartDelay: cfg.SingBox.RestartDelay,
	}, io.Discard, slog.New(slog.NewTextHandler(io.Discard, nil)))
	ctx := context.Background()
	switch args[0] {
	case "start":
		return pm.Start(ctx)
	case "stop":
		return pm.Stop(ctx)
	case "restart":
		return pm.Restart(ctx)
	case "reload":
		return pm.Reload(ctx)
	case "status":
		st, err := pm.Status(ctx)
		if err != nil {
			return err
		}
		return printJSON(st)
	case "config-check":
		return singbox.NewChecker(cfg.SingBox.BinaryPath, cfg.SingBox.CheckTimeout).Check(ctx, cfg.SingBox.ConfigPath)
	default:
		return fmt.Errorf("unknown core command %q", args[0])
	}
}

func cliCert(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: shilka cert set-files|reset|show")
	}
	switch args[0] {
	case "show":
		cfg := config.MustLoad()
		return printJSON(cfg.TLS)
	case "reset":
		if err := updateYAMLConfig(configPath(), map[string]any{"tls.mode": "off", "tls.cert_file": "", "tls.key_file": ""}); err != nil {
			return err
		}
		fmt.Println("cert settings reset")
		return nil
	case "set-files":
		fs := flag.NewFlagSet("cert set-files", flag.ContinueOnError)
		cert := fs.String("cert", "", "certificate path")
		key := fs.String("key", "", "key path")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *cert == "" || *key == "" {
			return fmt.Errorf("-cert and -key are required")
		}
		return updateYAMLConfig(configPath(), map[string]any{"tls.mode": "file", "tls.cert_file": *cert, "tls.key_file": *key})
	case "issue-domain", "issue-ip":
		return fmt.Errorf("%s is handled by the installer acme/self-signed flow in this release", args[0])
	default:
		return fmt.Errorf("unknown cert command %q", args[0])
	}
}

func resetTOTPForAdmin(cfg *config.Config) error {
	db, err := openCLIStorage(cfg)
	if err != nil {
		return err
	}
	defer db.Close()
	repo := sqliterepo.NewAdminRepo(db)
	admin, err := repo.GetByUsername(context.Background(), cfg.Auth.AdminUser)
	if err != nil {
		return err
	}
	admin.TOTPSecret = ""
	admin.IsTOTPEnabled = false
	admin.TOTPConfirmedAt = nil
	if err := repo.Update(context.Background(), admin); err != nil {
		return err
	}
	fmt.Println("totp disabled")
	return nil
}

func openCLIStorage(cfg *config.Config) (*sql.DB, error) {
	return sqliterepo.New(cfg.Database, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func configPath() string {
	if p := os.Getenv("SHILKA_CONFIG_PATH"); p != "" {
		return p
	}
	return "config/dev.yaml"
}

func updateYAMLConfig(path string, updates map[string]any) error {
	data := map[string]any{}
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(raw, &data); err != nil {
		return err
	}
	for key, value := range updates {
		setNested(data, strings.Split(key, "."), value)
	}
	out, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	backup := path + ".bak"
	if err := os.WriteFile(backup, raw, 0o600); err != nil {
		return err
	}
	tmp := filepath.Join(filepath.Dir(path), "."+filepath.Base(path)+".tmp")
	if err := os.WriteFile(tmp, out, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func setNested(m map[string]any, path []string, value any) {
	if len(path) == 1 {
		m[path[0]] = value
		return
	}
	child, ok := m[path[0]].(map[string]any)
	if !ok {
		child = map[string]any{}
		m[path[0]] = child
	}
	setNested(child, path[1:], value)
}

func cliID(args []string) (int64, error) {
	if len(args) == 0 {
		return 0, fmt.Errorf("id is required")
	}
	return strconv.ParseInt(args[0], 10, 64)
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
