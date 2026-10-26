package web

import (
	"crypto/sha256"
	"encoding/base64"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSecureHeaders(t *testing.T) {
	h := testServer(t, &countingAdvisor{}).Handler()
	for _, p := range []string{"/", "/api/overview", "/js/app.js", "/no/such/asset.js"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, p, nil))
		hd := rec.Header()
		csp := hd.Get("Content-Security-Policy")
		if !strings.Contains(csp, "script-src 'self' 'sha256-") || !strings.Contains(csp, "frame-ancestors 'none'") {
			t.Errorf("%s: CSP %q", p, csp)
		}
		if hd.Get("X-Content-Type-Options") != "nosniff" || hd.Get("X-Frame-Options") != "DENY" {
			t.Errorf("%s: headers %v", p, hd)
		}
		if hd.Get("Strict-Transport-Security") != "" {
			t.Errorf("%s: HSTS over plain HTTP", p)
		}
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Header().Get("Strict-Transport-Security") == "" {
		t.Error("no HSTS behind an HTTPS proxy")
	}
}

// The theme bootstrap in index.html must stay allowed: a CSP that misses its
// hash leaves the page on the wrong theme until the module script loads.
func TestCSPAllowsIndexInlineScript(t *testing.T) {
	sub, _ := fs.Sub(static, "static")
	b, err := fs.ReadFile(sub, "index.html")
	if err != nil {
		t.Fatal(err)
	}
	m := inlineScript.FindSubmatch(b)
	if m == nil {
		t.Skip("index.html has no inline script")
	}
	sum := sha256.Sum256(m[1])
	want := "'sha256-" + base64.StdEncoding.EncodeToString(sum[:]) + "'"
	if csp := contentSecurityPolicy(sub); !strings.Contains(csp, want) {
		t.Errorf("CSP %q lacks %s", csp, want)
	}
}
