package middleware

import (
	"net/http"
	"net/url"
	"strings"
)

var loopbackHosts = map[string]bool{
	"localhost":         true,
	"127.0.0.1":         true,
	"::1":               true,
	"[::1]":             true,
	"0.0.0.0":           true,
}

func isSameOrigin(originHeader, requestHost string) bool {
	o, err := url.Parse(originHeader)
	if err != nil {
		return false
	}

	originHost := o.Host
	originClean := strings.TrimPrefix(strings.TrimPrefix(originHost, "["), "]") // handle [::1]
	requestClean := strings.TrimPrefix(strings.TrimPrefix(requestHost, "["), "]")

	if originClean == requestClean {
		return true
	}

	if loopbackHosts[strings.Split(originClean, ":")[0]] &&
		loopbackHosts[strings.Split(requestClean, ":")[0]] {
		return true
	}

	return false
}

func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Vary", "Origin")
			}

			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Set("Access-Control-Max-Age", "86400")
				w.WriteHeader(http.StatusNoContent)
				return
			}

			if origin != "" {
				if isSameOrigin(origin, r.Host) {
					next.ServeHTTP(w, r)
					return
				}

				allowed := false
				for _, o := range allowedOrigins {
					if o == "*" || o == origin || strings.HasSuffix(origin, o) {
						allowed = true
						break
					}
				}
				if !allowed {
					http.Error(w, `{"error":"origin not allowed"}`, http.StatusForbidden)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
