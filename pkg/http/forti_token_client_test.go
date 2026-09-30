package http

import (
	"context"
	"encoding/base64"
	"encoding/json"
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
}

func TestResolveToken(t *testing.T) {
	tok, err := ResolveToken(config.TargetAuth{Username: "x", Password: "y"})
	if err != nil {
		t.Fatal(err)
	}
	want, err := EncodeToken("x", "y", "root")
	if err != nil {
		t.Fatal(err)
	}
	if tok != want {
		t.Errorf("got %q, want root-scoped token %q", tok, want)
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
	c, err := NewFortiClient(context.Background(), *u, srv.Client(), config.TargetAuth{Username: "admin", Password: "a"})
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

func TestWithVdomAuthorization(t *testing.T) {
	var got struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Vdom     string `json:"vdom"`
	}
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := base64.StdEncoding.DecodeString(r.Header.Get("Authorization"))
		if err != nil {
			t.Errorf("decode Authorization header: %v", err)
		}
		if err := json.Unmarshal(b, &got); err != nil {
			t.Errorf("unmarshal Authorization header: %v", err)
		}
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	u, _ := url.Parse(srv.URL)
	client, err := NewFortiClient(context.Background(), *u, srv.Client(), config.TargetAuth{Username: "admin", Password: "a"})
	if err != nil {
		t.Fatal(err)
	}
	client, err = client.WithVdom("000006")
	if err != nil {
		t.Fatal(err)
	}

	var obj map[string]interface{}
	if err := client.Get("api/v2.0/policy/policystatus", "", &obj); err != nil {
		t.Fatal(err)
	}
	if got.Username != "admin" || got.Password != "a" || got.Vdom != "000006" {
		t.Errorf("Authorization credentials = %+v, want same credentials scoped to VDOM 000006", got)
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
		c, err := NewFortiClient(context.Background(), *u, srv.Client(), config.TargetAuth{Username: "admin", Password: "a"})
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
