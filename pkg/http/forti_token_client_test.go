package http

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/mlueckert/fortiweb_exporter/internal/config"
)

func TestEncodeToken(t *testing.T) {
	tok, err := EncodeToken("admin", "a", "root")
	if err != nil {
		t.Fatal(err)
	}
	b, err := base64.StdEncoding.DecodeString(string(tok))
	if err != nil {
		t.Fatal(err)
	}
	if want := `{"username":"admin","password":"a","vdom":"root"}`; string(b) != want {
		t.Errorf("got %s, want %s", b, want)
	}

	tok, _ = EncodeToken("admin", "a", "")
	b, _ = base64.StdEncoding.DecodeString(string(tok))
	if want := `{"username":"admin","password":"a","vdom":"root"}`; string(b) != want {
		t.Errorf("default vdom: got %s, want %s", b, want)
	}
}

func TestResolveToken(t *testing.T) {
	if tok, _ := ResolveToken(config.TargetAuth{Token: "abc", Username: "x", Password: "y"}); tok != "abc" {
		t.Errorf("explicit token not used, got %q", tok)
	}
	if _, err := ResolveToken(config.TargetAuth{Username: "x"}); err == nil {
		t.Error("expected error for missing password")
	}
	if _, err := ResolveToken(config.TargetAuth{}); err == nil {
		t.Error("expected error for empty auth")
	}
}

func TestAuthorizationHeader(t *testing.T) {
	var got string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("Authorization")
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	u, _ := url.Parse(srv.URL)
	c, err := NewFortiClient(context.Background(), *u, srv.Client(), config.TargetAuth{Username: "admin", Password: "a", Vdom: "root"})
	if err != nil {
		t.Fatal(err)
	}
	var obj map[string]interface{}
	if err := c.Get("/api/v2.0/system/status.systemresource", "", &obj); err != nil {
		t.Fatal(err)
	}
	want, _ := EncodeToken("admin", "a", "root")
	if got != string(want) {
		t.Errorf("Authorization header = %q, want %q", got, want)
	}
}

func TestStatusCodes(t *testing.T) {
	for _, tc := range []struct {
		code    int
		wantErr bool
	}{
		{200, false},
		{220, false},
		{401, true},
		{500, true},
	} {
		srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(tc.code)
			w.Write([]byte(`{"results":{"cfg_sync_state":"In sync"}}`))
		}))
		u, _ := url.Parse(srv.URL)
		c, err := NewFortiClient(context.Background(), *u, srv.Client(), config.TargetAuth{Token: "t"})
		if err != nil {
			t.Fatal(err)
		}
		var obj map[string]interface{}
		err = c.Get("/api/v2.0/system/status.hastatus", "", &obj)
		if (err != nil) != tc.wantErr {
			t.Errorf("code %d: got err %v, wantErr %v", tc.code, err, tc.wantErr)
		}
		srv.Close()
	}
}
