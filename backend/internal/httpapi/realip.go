package httpapi

import (
	"net"
	"net/http"
	"strings"
)

// TrustedProxyRealIP rewrites r.RemoteAddr from the X-Forwarded-For header,
// but only when the immediate connection (r.RemoteAddr) comes from a
// loopback address — i.e. a local reverse proxy such as the nginx instance
// Forge manages in front of this service (see docs/DEPLOY.md). This avoids
// the IP-spoofing hole in chi's deprecated middleware.RealIP, which trusts
// X-Forwarded-For unconditionally and takes the left-most (client-supplied)
// value: an internet client can set that header to anything, and downstream
// code (rate limiting, logging) would trust it.
//
// It also takes the right-most X-Forwarded-For entry, since that is the one
// appended by our own trusted proxy — not one an upstream client could have
// forged. If there is no reverse proxy in front of this service (e.g. local
// Docker Compose, where the frontend calls the API directly), this
// middleware is a no-op and r.RemoteAddr — the real socket peer — is used
// as-is.
func TrustedProxyRealIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ip := trustedForwardedFor(r); ip != "" {
			r.RemoteAddr = ip
		}
		next.ServeHTTP(w, r)
	})
}

func trustedForwardedFor(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	peer := net.ParseIP(host)
	if peer == nil || !peer.IsLoopback() {
		return ""
	}

	xff := r.Header.Get("X-Forwarded-For")
	if xff == "" {
		return ""
	}

	parts := strings.Split(xff, ",")
	last := strings.TrimSpace(parts[len(parts)-1])
	if last == "" || net.ParseIP(last) == nil {
		return ""
	}

	return last
}
