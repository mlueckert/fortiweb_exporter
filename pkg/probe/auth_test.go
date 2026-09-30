package probe

import (
	"testing"

	"github.com/mlueckert/fortiweb_exporter/internal/config"
)

func TestResolveAuth(t *testing.T) {
	profile := config.TargetAuth{Username: "p-user", Password: "p-pass"} // betterleaks:allow (test value)

	got := resolveAuth(map[string]string{}, profile)
	if got.Username != "p-user" || got.Password != "p-pass" {
		t.Errorf("profile values not used: %+v", got)
	}

	got = resolveAuth(map[string]string{"username": "u", "password": "p"}, profile)
	if got.Username != "u" || got.Password != "p" {
		t.Errorf("params did not override profile: %+v", got)
	}

	got = resolveAuth(map[string]string{"password": "p"}, profile)
	if got.Username != "p-user" || got.Password != "p" {
		t.Errorf("password param did not override profile: %+v", got)
	}
}
