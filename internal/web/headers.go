package web

import (
	"crypto/sha256"
	"encoding/base64"
	"io/fs"
	"net/http"
	"regexp"
	"strings"
)

// inlineScript matches an attribute-less <script> block: the pages' theme
// bootstraps, which run before any stylesheet to avoid a flash of the wrong
// theme and so cannot move to a file.
var inlineScript = regexp.MustCompile(`(?s)<script>(.*?)</script>`)

// contentSecurityPolicy allows the UI's own scripts and the inline scripts in
// root's pages by hash, so a URL or markup that slips into a template still
// cannot run script; javascript: links are refused too. Styles keep
// 'unsafe-inline' for the charts' positioning style attributes. Photos are
// hotlinked from the shops, so img-src takes any https origin.
func contentSecurityPolicy(root fs.FS) string {
	scripts := []string{"'self'"}
	fs.WalkDir(root, ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(name, ".html") {
			return nil
		}
		b, err := fs.ReadFile(root, name)
		if err != nil {
			return nil
		}
		for _, m := range inlineScript.FindAllSubmatch(b, -1) {
			sum := sha256.Sum256(m[1])
			scripts = append(scripts, "'sha256-"+base64.StdEncoding.EncodeToString(sum[:])+"'")
		}
		return nil
	})
	return strings.Join([]string{
		"default-src 'self'",
		"script-src " + strings.Join(scripts, " "),
		"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com",
		"font-src https://fonts.gstatic.com",
		"img-src 'self' https: data:",
		"connect-src 'self'",
		"object-src 'none'",
		"base-uri 'none'",
		"form-action 'self'",
		"frame-ancestors 'none'",
	}, "; ")
}

// secureHeaders sets the security headers on every response. HSTS goes out
// only over HTTPS (Railway's proxy reports it in X-Forwarded-Proto) and
// without includeSubDomains, since the parent domain hosts other demos.
func secureHeaders(csp string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy", csp)
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			h.Set("Strict-Transport-Security", "max-age=31536000")
		}
		next.ServeHTTP(w, r)
	})
}
