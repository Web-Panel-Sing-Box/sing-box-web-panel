package stats_test

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protowire"

	"sing-box-web-panel/internal/services/stats"
)

// rawCodec is a passthrough gRPC codec for the fake server: it ships pre-encoded
// protobuf bytes verbatim, letting the test control the exact wire response.
type rawCodec struct{}

func (rawCodec) Name() string { return "proto" }

func (rawCodec) Marshal(v any) ([]byte, error) { return *(v.(*[]byte)), nil }

func (rawCodec) Unmarshal(data []byte, v any) error {
	*(v.(*[]byte)) = append([]byte(nil), data...)
	return nil
}

func encodeStat(name string, value int64) []byte {
	var b []byte
	b = protowire.AppendTag(b, 1, protowire.BytesType)
	b = protowire.AppendString(b, name)
	b = protowire.AppendTag(b, 2, protowire.VarintType)
	b = protowire.AppendVarint(b, uint64(value))
	return b
}

// encodeQueryStatsResponse builds the wire bytes for a QueryStatsResponse.
func encodeQueryStatsResponse(stats map[string]int64) []byte {
	var out []byte
	for name, value := range stats {
		msg := encodeStat(name, value)
		out = protowire.AppendTag(out, 1, protowire.BytesType)
		out = protowire.AppendBytes(out, msg)
	}
	return out
}

// TestV2RayUserDeltasCodec validates the hand-rolled StatsService codec against
// an in-process gRPC server returning a canned QueryStatsResponse.
func TestV2RayUserDeltasCodec(t *testing.T) {
	response := encodeQueryStatsResponse(map[string]int64{
		"user>>>alice>>>traffic>>>uplink":   1500,
		"user>>>alice>>>traffic>>>downlink": 3000,
		"user>>>bob>>>traffic>>>uplink":     42,
	})

	desc := grpc.ServiceDesc{
		ServiceName: "v2ray.core.app.stats.command.StatsService",
		HandlerType: (*any)(nil),
		Methods: []grpc.MethodDesc{{
			MethodName: "QueryStats",
			Handler: func(_ any, _ context.Context, dec func(any) error, _ grpc.UnaryServerInterceptor) (any, error) {
				var req []byte
				if err := dec(&req); err != nil {
					return nil, err
				}
				resp := response
				return &resp, nil
			},
		}},
	}

	srv := grpc.NewServer(grpc.ForceServerCodec(rawCodec{}))
	srv.RegisterService(&desc, nil)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	go srv.Serve(lis)
	defer srv.Stop()

	src := stats.NewV2RaySource(lis.Addr().String())
	defer src.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	deltas, err := src.UserDeltas(ctx)
	if err != nil {
		t.Fatalf("UserDeltas: %v", err)
	}

	got := make(map[string][2]int64)
	for _, d := range deltas {
		got[d.Name] = [2]int64{d.Up, d.Down}
	}
	if v := got["alice"]; v != [2]int64{1500, 3000} {
		t.Errorf("alice: got up=%d down=%d, want up=1500 down=3000", v[0], v[1])
	}
	if v := got["bob"]; v != [2]int64{42, 0} {
		t.Errorf("bob: got up=%d down=%d, want up=42 down=0", v[0], v[1])
	}
}
