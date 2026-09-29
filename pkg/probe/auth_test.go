package probe

import (
	"testing"

	"github.com/mlueckert/fortiweb_exporter/internal/config"
)

func TestResolveAuth(t *testing.T) {
	profile := config.TargetAuth{Username: "p-user", Password: "p-pass", Vdom: "p-vdom"} // betterleaks:allow (test value)

	got := resolveAuth(map[string]string{}, profile)
	if got.Username != "p-user" || got.Password != "p-pass" || got.Vdom != "p-vdom" {
		t.Errorf("profile values not used: %+v", got)
	}

	got = resolveAuth(map[string]string{"username": "u", "password": "p", "vdom": "v"}, profile)
	if got.Username != "u" || got.Password != "p" || got.Vdom != "v" {
		t.Errorf("params did not override profile: %+v", got)
	}

	got = resolveAuth(map[string]string{"password": "p"}, config.TargetAuth{Token: "t", Username: "p-user"})
	if got.Token != "" || got.Username != "p-user" || got.Password != "p" {
		t.Errorf("credential params should replace profile token: %+v", got)
	}

	got = resolveAuth(map[string]string{"token": "t", "username": "u"}, profile)
	if got.Token != "t" {
		t.Errorf("token param not used: %+v", got)
	}
}
