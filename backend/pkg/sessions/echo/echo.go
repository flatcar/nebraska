package echo

import (
	"github.com/labstack/echo/v4"

	"github.com/flatcar/nebraska/backend/pkg/sessions"
)

func SessionsMiddleware(s *sessions.Store, name string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			session := s.GetSessionUse(c.Request(), name)
			c.Set("session", session)
			defer func() {
				s.PutSessionUse(session)
			}()
			return next(c)
		}
	}
}

func GetSession(c echo.Context) *sessions.Session {
	session, ok := c.Get("session").(*sessions.Session)
	if !ok {
		// Should never happen, because session middleware is always set
		return nil
	}
	return session
}

func SaveSession(c echo.Context, session *sessions.Session) error {
	// Scheme also honours X-Forwarded-Proto, so cookies stay HTTPS-only
	// behind a TLS-terminating proxy. A spoofed header can only add the
	// Secure attribute, never remove it.
	return session.Save(c.Response().Writer, c.Scheme() == "https")
}
