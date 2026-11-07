package service

import (
	"encoding/json"
	"testing"
)

func TestLiftLegacyXdnsMasks(t *testing.T) {
	tests := []struct {
		name       string
		stream     string
		wantLifted int
		wantStream string
	}{
		{
			name:       "legacy domain and resolver strings become structured objects",
			stream:     `{"finalmask":{"udp":[{"type":"xdns","settings":{"domains":["example.com:txt","example.org:a"],"resolvers":["example.com:txt+udp://1.1.1.1:53"]}}]}}`,
			wantLifted: 1,
			wantStream: `{"finalmask":{"udp":[{"type":"xdns","settings":{"domains":[{"names":["example.com"],"types":[16]},{"names":["example.org"],"types":[1]}],"resolvers":[{"addrs":["1.1.1.1:53"]}]}}]}}`,
		},
		{
			name:       "domain without a type suffix carries no types",
			stream:     `{"finalmask":{"udp":[{"type":"xdns","settings":{"domains":["example.com"]}}]}}`,
			wantLifted: 1,
			wantStream: `{"finalmask":{"udp":[{"type":"xdns","settings":{"domains":[{"names":["example.com"]}]}}]}}`,
		},
		{
			name:       "aaaa suffix maps to type 28",
			stream:     `{"finalmask":{"udp":[{"type":"xdns","settings":{"domains":["example.net:aaaa"]}}]}}`,
			wantLifted: 1,
			wantStream: `{"finalmask":{"udp":[{"type":"xdns","settings":{"domains":[{"names":["example.net"],"types":[28]}]}}]}}`,
		},
		{
			name:       "unknown type suffix is dropped rather than rejected",
			stream:     `{"finalmask":{"udp":[{"type":"xdns","settings":{"domains":["example.io:srv"]}}]}}`,
			wantLifted: 1,
			wantStream: `{"finalmask":{"udp":[{"type":"xdns","settings":{"domains":[{"names":["example.io"]}]}}]}}`,
		},
		{
			name:       "structured mask is left alone",
			stream:     `{"finalmask":{"udp":[{"type":"xdns","settings":{"domains":[{"names":["example.com"],"types":[16]}],"resolvers":[{"addrs":["1.1.1.1:53"]}]}}]}}`,
			wantLifted: 0,
			wantStream: `{"finalmask":{"udp":[{"type":"xdns","settings":{"domains":[{"names":["example.com"],"types":[16]}],"resolvers":[{"addrs":["1.1.1.1:53"]}]}}]}}`,
		},
		{
			name:       "non-xdns masks are untouched",
			stream:     `{"finalmask":{"udp":[{"type":"salamander","settings":{"password":"swordfish"}}]}}`,
			wantLifted: 0,
			wantStream: `{"finalmask":{"udp":[{"type":"salamander","settings":{"password":"swordfish"}}]}}`,
		},
		{
			name:       "stream without finalmask is untouched",
			stream:     `{"network":"tcp","security":"tls"}`,
			wantLifted: 0,
			wantStream: `{"network":"tcp","security":"tls"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stream map[string]any
			if err := json.Unmarshal([]byte(tt.stream), &stream); err != nil {
				t.Fatalf("unmarshal stream: %v", err)
			}

			lifted := liftLegacyXdnsMasks(stream)
			if lifted != tt.wantLifted {
				t.Errorf("liftLegacyXdnsMasks lifted = %d, want %d", lifted, tt.wantLifted)
			}

			got, err := json.Marshal(stream)
			if err != nil {
				t.Fatalf("marshal stream: %v", err)
			}
			var want map[string]any
			if err := json.Unmarshal([]byte(tt.wantStream), &want); err != nil {
				t.Fatalf("unmarshal wantStream: %v", err)
			}
			wantJSON, err := json.Marshal(want)
			if err != nil {
				t.Fatalf("marshal wantStream: %v", err)
			}
			if string(got) != string(wantJSON) {
				t.Errorf("stream after lift = %s, want %s", got, wantJSON)
			}

			if second := liftLegacyXdnsMasks(stream); second != 0 {
				t.Errorf("second lift = %d, want 0 (healing must be idempotent)", second)
			}
		})
	}
}

func TestLiftedLegacyXdnsMaskBuildsInXray(t *testing.T) {
	stream := map[string]any{
		"network": "tcp",
		"finalmask": map[string]any{
			"udp": []any{
				map[string]any{
					"type": "xdns",
					"settings": map[string]any{
						"domains":   []any{"example.com:txt", "example.org:a"},
						"resolvers": []any{"example.com:txt+udp://1.1.1.1:53"},
					},
				},
			},
		},
	}
	if lifted := liftLegacyXdnsMasks(stream); lifted != 1 {
		t.Fatalf("lifted = %d, want 1", lifted)
	}
	inbound := map[string]any{
		"tag": "xdns-lift", "listen": "127.0.0.1", "port": 8443, "protocol": "vless",
		"settings": map[string]any{
			"clients":    []any{map[string]any{"id": "b831381d-6324-4d53-ad4f-8cda48b30811", "email": "xdns"}},
			"decryption": "none",
		},
		"streamSettings": stream,
	}
	assertXrayAccepts(t, "the lifted xdns mask", buildGoldenInbound(t, inbound))
}
