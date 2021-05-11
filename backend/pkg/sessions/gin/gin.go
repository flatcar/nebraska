package gin

import (
	"github.com/gin-gonic/gin"

	"github.com/kinvolk/nebraska/backend/pkg/sessions"
)

func SessionsMiddleware(s *sessions.Store, name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := s.GetSessionUse(c.Request, name)
		oldRequest := c.Request
		ctx := sessions.ContextWithSession(c.Request.Context(), session)
		c.Request = c.Request.WithContext(ctx)
		defer func() {
			c.Request = oldRequest
			s.PutSessionUse(session)
		}()
		c.Next()
	}
}

func GetSession(c *gin.Context) *sessions.Session {
	return sessions.SessionFromContext(c.Request.Context())
}

func SaveSession(c *gin.Context, session *sessions.Session) error {
	return session.Save(c.Writer)
}
