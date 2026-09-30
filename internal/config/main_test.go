package config

import "testing"

func TestParseAuthKeys(t *testing.T) {
	for _, tc := range []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"credentials", "default:\n  username: admin\n  password: example\n", false},
		{"pre-created token", "default:\n  token: encoded\n", true},
		{"explicit vdom", "default:\n  vdom: other\n", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			keys, err := parseAuthKeys([]byte(tc.input))
			if (err != nil) != tc.wantErr {
				t.Fatalf("parseAuthKeys() error = %v, wantErr %v", err, tc.wantErr)
			}
			if !tc.wantErr && (keys["default"].Username != "admin" || keys["default"].Password != "example") {
				t.Errorf("parsed credentials = %+v, want username and password", keys["default"])
			}
		})
	}
}
