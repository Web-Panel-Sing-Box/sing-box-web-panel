package stats

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protowire"

	"sing-box-web-panel/internal/domain"
)

// V2RaySource reads cumulative per-user counters from the sing-box V2Ray API
// (gRPC StatsService). Querying with reset=true returns the delta since the
// previous call. Requires a sing-box binary built with `with_v2ray_api` and an
// `experimental.v2ray_api.stats` block in the config.
//
// To keep dependencies minimal there is no generated protobuf code: requests
// and responses are encoded/decoded by hand with the protobuf wire codec.
type V2RaySource struct {
	addr string

	mu   sync.Mutex
	conn *grpc.ClientConn
}

func NewV2RaySource(apiAddress string) *V2RaySource {
	return &V2RaySource{addr: apiAddress}
}

// sing-box renames the gRPC service at init() to the standard V2Ray name for
// compatibility with existing stats clients, so the proto package name
// (experimental.v2rayapi) is NOT used on the wire.
const queryStatsMethod = "/v2ray.core.app.stats.command.StatsService/QueryStats"

func (s *V2RaySource) dial() (*grpc.ClientConn, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conn != nil {
		return s.conn, nil
	}
	conn, err := grpc.NewClient(s.addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial v2ray api: %w", err)
	}
	s.conn = conn
	return conn, nil
}

// UserDeltas returns per-user up/down byte deltas since the previous call.
func (s *V2RaySource) UserDeltas(ctx context.Context) ([]domain.UserTraffic, error) {
	conn, err := s.dial()
	if err != nil {
		return nil, err
	}

	req := &queryStatsRequest{Pattern: "user>>>", Reset: true}
	resp := &queryStatsResponse{}
	if err := conn.Invoke(ctx, queryStatsMethod, req, resp, grpc.ForceCodec(wireCodec{})); err != nil {
		return nil, fmt.Errorf("query stats: %w", err)
	}

	byName := make(map[string]*domain.UserTraffic)
	for _, st := range resp.Stats {
		name, dir, ok := parseUserStat(st.Name)
		if !ok {
			continue
		}
		ut := byName[name]
		if ut == nil {
			ut = &domain.UserTraffic{Name: name}
			byName[name] = ut
		}
		switch dir {
		case "uplink":
			ut.Up += st.Value
		case "downlink":
			ut.Down += st.Value
		}
	}

	out := make([]domain.UserTraffic, 0, len(byName))
	for _, ut := range byName {
		out = append(out, *ut)
	}
	return out, nil
}

func (s *V2RaySource) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conn != nil {
		err := s.conn.Close()
		s.conn = nil
		return err
	}
	return nil
}

// parseUserStat decodes "user>>>NAME>>>traffic>>>uplink|downlink".
func parseUserStat(name string) (user, direction string, ok bool) {
	rest, found := strings.CutPrefix(name, "user>>>")
	if !found {
		return "", "", false
	}
	parts := strings.Split(rest, ">>>")
	if len(parts) < 3 {
		return "", "", false
	}
	return parts[0], parts[len(parts)-1], true
}

// --- hand-rolled protobuf wire codec for the two message types ---

type queryStatsRequest struct {
	Pattern string
	Reset   bool
}

type stat struct {
	Name  string
	Value int64
}

type queryStatsResponse struct {
	Stats []stat
}

// wireCodec implements grpc's encoding.Codec for the StatsService messages only.
type wireCodec struct{}

func (wireCodec) Name() string { return "proto" }

func (wireCodec) Marshal(v any) ([]byte, error) {
	req, ok := v.(*queryStatsRequest)
	if !ok {
		return nil, fmt.Errorf("wireCodec: unsupported marshal type %T", v)
	}
	var b []byte
	if req.Pattern != "" {
		b = protowire.AppendTag(b, 1, protowire.BytesType)
		b = protowire.AppendString(b, req.Pattern)
	}
	if req.Reset {
		b = protowire.AppendTag(b, 2, protowire.VarintType)
		b = protowire.AppendVarint(b, 1)
	}
	return b, nil
}

func (wireCodec) Unmarshal(data []byte, v any) error {
	resp, ok := v.(*queryStatsResponse)
	if !ok {
		return fmt.Errorf("wireCodec: unsupported unmarshal type %T", v)
	}
	b := data
	for len(b) > 0 {
		num, typ, n := protowire.ConsumeTag(b)
		if n < 0 {
			return protowire.ParseError(n)
		}
		b = b[n:]
		if num == 1 && typ == protowire.BytesType {
			msg, mn := protowire.ConsumeBytes(b)
			if mn < 0 {
				return protowire.ParseError(mn)
			}
			b = b[mn:]
			st, err := parseStatMessage(msg)
			if err != nil {
				return err
			}
			resp.Stats = append(resp.Stats, st)
			continue
		}
		fn := protowire.ConsumeFieldValue(num, typ, b)
		if fn < 0 {
			return protowire.ParseError(fn)
		}
		b = b[fn:]
	}
	return nil
}

func parseStatMessage(b []byte) (stat, error) {
	var st stat
	for len(b) > 0 {
		num, typ, n := protowire.ConsumeTag(b)
		if n < 0 {
			return st, protowire.ParseError(n)
		}
		b = b[n:]
		switch {
		case num == 1 && typ == protowire.BytesType:
			name, ln := protowire.ConsumeString(b)
			if ln < 0 {
				return st, protowire.ParseError(ln)
			}
			st.Name = name
			b = b[ln:]
		case num == 2 && typ == protowire.VarintType:
			val, ln := protowire.ConsumeVarint(b)
			if ln < 0 {
				return st, protowire.ParseError(ln)
			}
			st.Value = int64(val)
			b = b[ln:]
		default:
			fn := protowire.ConsumeFieldValue(num, typ, b)
			if fn < 0 {
				return st, protowire.ParseError(fn)
			}
			b = b[fn:]
		}
	}
	return st, nil
}
