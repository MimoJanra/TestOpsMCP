package mcp

import (
	"encoding/json"
	"testing"
)

func TestIsNotification(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want bool
	}{
		{"omitted", `{"jsonrpc":"2.0","method":"ping"}`, true},
		{"null", `{"jsonrpc":"2.0","id":null,"method":"ping"}`, true},
		{"zero int", `{"jsonrpc":"2.0","id":0,"method":"ping"}`, false},
		{"string id", `{"jsonrpc":"2.0","id":"abc","method":"ping"}`, false},
		{"int id", `{"jsonrpc":"2.0","id":42,"method":"ping"}`, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var req JSONRPCRequest
			if err := json.Unmarshal([]byte(tc.raw), &req); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if got := req.IsNotification(); got != tc.want {
				t.Errorf("IsNotification()=%v want %v (raw=%s)", got, tc.want, tc.raw)
			}
		})
	}
}

func TestProtocolVersionConstant(t *testing.T) {
	if ProtocolVersion == "" {
		t.Fatal("ProtocolVersion must not be empty")
	}
}
