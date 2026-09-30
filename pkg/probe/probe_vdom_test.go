package probe

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/google/go-jsonnet"
	"github.com/mlueckert/fortiweb_exporter/internal/config"
)

func TestProbeUsesRootForSystemAndEachVDOMForPolicies(t *testing.T) {
	fixtures := map[string]string{
		"api/v2.0/system/status.systemresource":  "testdata/system_status_systemresource.jsonnet",
		"api/v2.0/system/status.systemstatus":    "testdata/system_status_systemstatus.jsonnet",
		"api/v2.0/system/status.hastatus":        "testdata/system_status_hastatus.jsonnet",
		"api/v2.0/system/status.systemoperation": "testdata/system_status_systemoperation.jsonnet",
		"api/v2.0/system/vip":                    "testdata/system_vip.jsonnet",
		"api/v2.0/policy/policystatus":           "testdata/policy_policyStatus.jsonnet",
	}
	responses := make(map[string]string, len(fixtures))
	for path, fixture := range fixtures {
		body, err := jsonnet.MakeVM().EvaluateFile(fixture)
		if err != nil {
			t.Fatalf("evaluate fixture %q: %v", fixture, err)
		}
		responses[path] = body
	}

	var (
		mu             sync.Mutex
		requestedVdoms = make(map[string][]string)
	)
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		var credentials struct {
			Username string `json:"username"`
			Password string `json:"password"`
			Vdom     string `json:"vdom"`
		}
		token, err := base64.StdEncoding.DecodeString(r.Header.Get("Authorization"))
		if err != nil {
			t.Errorf("decode Authorization header: %v", err)
		} else if err := json.Unmarshal(token, &credentials); err != nil {
			t.Errorf("decode Authorization credentials: %v", err)
		} else if credentials.Username != "admin" || credentials.Password != "test-password" {
			t.Errorf("unexpected credentials for %s", path)
		}

		mu.Lock()
		requestedVdoms[path] = append(requestedVdoms[path], credentials.Vdom)
		mu.Unlock()

		body, ok := responses[path]
		if !ok {
			http.Error(w, "unexpected API path", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	pc := &ProbeCollector{}
	success, err := pc.Probe(
		context.Background(),
		map[string]string{"target": srv.URL, "profile": "default"},
		srv.Client(),
		config.FortiWebExporterConfig{
			AuthKeys: config.AuthKeys{
				config.Target("default"): {
					Username: "admin",
					Password: "test-password", // betterleaks:allow (test value)
				},
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if !success {
		t.Fatal("probe returned failure")
	}

	mu.Lock()
	defer mu.Unlock()
	for _, path := range []string{
		"api/v2.0/system/status.systemresource",
		"api/v2.0/system/status.systemstatus",
		"api/v2.0/system/status.hastatus",
		"api/v2.0/system/status.systemoperation",
		"api/v2.0/system/vip",
	} {
		if got := requestedVdoms[path]; len(got) != 1 || got[0] != "root" {
			t.Errorf("%s requested with VDOMs %v, want [root]", path, got)
		}
	}

	wantPolicyVdoms := []string{"root", "tenant-a", "shared", "tenant-b"}
	if got := requestedVdoms["api/v2.0/policy/policystatus"]; !equalStrings(got, wantPolicyVdoms) {
		t.Errorf("policy requests used VDOMs %v, want %v", got, wantPolicyVdoms)
	}
}

func equalStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
