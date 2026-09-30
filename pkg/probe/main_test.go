package probe

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProbeHandlerRejectsUnsupportedAuthParams(t *testing.T) {
	for _, query := range []string{
		"?target=https://example.com&token=encoded",
		"?target=https://example.com&token=",
		"?target=https://example.com&vdom=other",
	} {
		t.Run(query, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/probe"+query, nil)
			w := httptest.NewRecorder()
			ProbeHandler(w, r)
			if w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), "not supported") {
				t.Errorf("response = %d %q, want unsupported parameter error", w.Code, w.Body.String())
			}
		})
	}
}
