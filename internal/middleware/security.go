package middleware

import "github.com/gin-gonic/gin"

// SecurityHeaders adds basic browser hardening headers to every response.
// HSTS is added by Caddy (only meaningful over HTTPS).
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		// Camera is needed by the manasik QR scanner on our own pages only.
		h.Set("Permissions-Policy", "camera=(self), microphone=(), geolocation=()")
		c.Next()
	}
}
