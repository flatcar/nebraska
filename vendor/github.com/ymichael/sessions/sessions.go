package sessions

import (
	"errors"
	"net/http"
	"time"

	"github.com/zenazn/goji/web"
)

var SessionIdNotFound = errors.New("Session Id Not Found.")

// Configuration Options for Sessions.
type SessionOptions struct {
	Name          string
	Secret        string
	ObjEnvKey     string
	SidEnvKey     string
	Store         Store
	CookieOptions *CookieOptions
}

// Helper to create SessionOptions object with sensible defaults.
func NewSessionOptions(secret string, store Store) *SessionOptions {
	return &SessionOptions{
		Name:          "gojisid",
		ObjEnvKey:     "sessionObject",
		SidEnvKey:     "sessionId",
		CookieOptions: &CookieOptions{"/", 0, true, false},
		Secret:        secret,
		Store:         store,
	}
}

// Initialise session object.
func (s SessionOptions) RetrieveOrCreateSession(c *web.C, r *http.Request) {
	var sessionId string
	var sessionObj map[string]interface{}
	// If cookie is set, retrieve session and set on context.
	// Otherwise, create new session.
	sessionId, err := s.GetValueFromCookie(r)
	if err != nil {
		s.CreateNewSession(c)
		return
	}
	sessionObj, err = s.Store.Get(sessionId)
	if err != nil {
		sessionObj = make(map[string]interface{})
	}
	c.Env[s.SidEnvKey] = sessionId
	c.Env[s.ObjEnvKey] = sessionObj
}

// Create new session id and session obj.
func (s SessionOptions) CreateNewSession(c *web.C) {
	c.Env[s.SidEnvKey] = GenerateRandomString(24)
	c.Env[s.ObjEnvKey] = make(map[string]interface{})
}

// Get session id from request cookie. Returns SessionIdNotFound error if not found.
func (s SessionOptions) GetValueFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(s.Name)
	if err != nil {
		return "", SessionIdNotFound
	}
	val, err := UnsignMessage(cookie.Value, s.Secret)
	if err != nil {
		return "", SessionIdNotFound
	}
	return val, nil
}

// Set session cookie on response.
func (s SessionOptions) SetCookie(c *web.C, w http.ResponseWriter) {
	val := SignMessage(s.GetSessionId(c), s.Secret)
	cookie := NewCookie(s.Name, val, s.CookieOptions)
	http.SetCookie(w, cookie)
}

// Removes cookie by setting value to "", maxAge to -1 and to expired.
func (s SessionOptions) RemoveCookie(c *web.C, w http.ResponseWriter) {
	s.UpdateCookie(c, w, "", -1)
}

func (s SessionOptions) UpdateCookie(c *web.C, w http.ResponseWriter, value string, maxAge int) {
	// TODO: Refactor Set,Remove,Update Cookie to use one codepath.
	// Sign cookie value.
	value = SignMessage(value, s.Secret)

	// Find cookie in w.Header()
	h := w.Header()
	for idx, line := range h["Set-Cookie"] {
		cookie, err := CookieFromString(line)
		if err != nil {
			continue
		}
		// If found, set cookie maxAge to -1
		if cookie.Name == s.Name {
			cookie.Value = value
			cookie.MaxAge = maxAge
			if maxAge > 0 {
				d := time.Duration(maxAge) * time.Second
				cookie.Expires = time.Now().Add(d)
			} else if maxAge < 0 {
				cookie.Expires = time.Unix(1, 0)
			}
			h["Set-Cookie"][idx] = cookie.String()
			return
		}
	}
	// Otherwise, create new cookie with maxAge set to -1
	cookie := NewCookie(s.Name, value, s.CookieOptions)
	cookie.MaxAge = maxAge
	if maxAge > 0 {
		d := time.Duration(maxAge) * time.Second
		cookie.Expires = time.Now().Add(d)
	} else if maxAge < 0 {
		cookie.Expires = time.Unix(1, 0)
	}
	http.SetCookie(w, cookie)
}

// Get session object from context.
func (s SessionOptions) GetSessionObject(c *web.C) map[string]interface{} {
	return c.Env[s.ObjEnvKey].(map[string]interface{})
}

// Get session id from context.
func (s SessionOptions) GetSessionId(c *web.C) string {
	return c.Env[s.SidEnvKey].(string)
}

// Persist Session to Store.
func (s SessionOptions) SaveSession(c *web.C) {
	var key, obj interface{}
	var ok bool
	if key, ok = c.Env[s.SidEnvKey]; !ok {
		return
	}
	if obj, ok = c.Env[s.ObjEnvKey]; !ok {
		return
	}
	s.Store.Save(key.(string), obj.(map[string]interface{}))
}

// Destroy session. Will be regenerated next request.
func (s SessionOptions) DestroySession(c *web.C, w http.ResponseWriter) {
	s.Store.Destroy(s.GetSessionId(c))
	s.RemoveCookie(c, w)
	delete(c.Env, s.SidEnvKey)
	delete(c.Env, s.ObjEnvKey)
}

// Regenerate session. Destroys current session and assigns a completely new
// session id.
func (s SessionOptions) RegenerateSession(c *web.C, w http.ResponseWriter) {
	s.DestroySession(c, w)
	s.CreateNewSession(c)
	s.UpdateCookie(c, w, s.GetSessionId(c), s.CookieOptions.MaxAge)
}

// Returns session middleware.
func (s SessionOptions) Middleware() web.MiddlewareType {
	middlewareFn := func(c *web.C, h http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if c.Env == nil {
				c.Env = make(map[interface{}]interface{})
			}
			s.RetrieveOrCreateSession(c, r)
			s.SetCookie(c, w)
			h.ServeHTTP(w, r)
			s.SaveSession(c)
		}
		return http.HandlerFunc(fn)
	}
	return middlewareFn
}
