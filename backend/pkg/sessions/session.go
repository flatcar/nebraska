package sessions

import (
	"net/http"
)

// ValuesType is a type of values container used in the session.
type ValuesType map[interface{}]interface{}

// SessionExt is an interface purely for implementations of the Cache
// interface, so they can change some data in the session that typical
// user of the session shouldn't be allowed to do.
type SessionExt interface {
	// Name returns a name of the session.
	Name() string
	// ID returns an ID of the session.
	ID() string
	// SetID updates an ID of the session.
	SetID(id string)
	// GetValues returns the values map.
	GetValues() ValuesType
	// Session returns the underlying session object.
	Session() *Session
}

// SessionBuilder is an interface purely for implementation of the
// Cache interface, so they can create new session objects.
type SessionBuilder interface {
	// NewSession creates a new session object with a given name, and
	// associates it with the passed cache. The created cache has
	// no ID.
	NewSession(name string, cache Cache) SessionExt
	// NewExistingSession created a session object with passed
	// name, id and values map, and associates it with the passed
	// cache.
	NewExistingSession(name, id string, values ValuesType, cache Cache) SessionExt
}

// Session is a snapshot of the session data stored in the cache.
type Session struct {
	name   string
	id     string
	values ValuesType
	cache  Cache
	codec  Codec
	marked bool
}

// Has returns true if the session object contains a value under the
// passed key.
func (s *Session) Has(key interface{}) bool {
	_, ok := s.values[key]
	return ok
}

// Get returns a value mapped to the key, or nil.
func (s *Session) Get(key interface{}) interface{} {
	return s.values[key]
}

// Set inserts or updates the value under the key.
func (s *Session) Set(key, value interface{}) {
	s.values[key] = value
}

// Drop drops the value under the key.
func (s *Session) Drop(key interface{}) {
	delete(s.values, key)
}

// ID returns an ID of the session.
func (s *Session) ID() string {
	return s.id
}

// Name returns a name of the session.
func (s *Session) Name() string {
	return s.name
}

// Save persists the information stored in the session object to the
// cache associated with the object, and updates the cookie header in
// the response.
//
// TODO: Split the writing the cookie into a separate function, like
// WriteCookie.
//
// TODO: Add support for options, so we don't hardcode the max age
// value here. The support should probably be reflected in the Cache
// interface itself and in Session.
func (s *Session) Save(writer http.ResponseWriter) error {
	if !s.marked {
		var err error
		s.marked, err = s.cache.SaveSession(sessionExt{
			session: s,
		})
		if err != nil {
			return err
		}
	}
	maxAge := 86400 * 30
	value := ""
	if s.marked {
		maxAge = -1
	} else {
		encoded, err := s.codec.Encode(s.Name(), s.ID())
		if err != nil {
			return err
		}
		value = encoded
	}
	cookie := &http.Cookie{
		Name:     s.Name(),
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
	}
	http.SetCookie(writer, cookie)
	return nil
}

// Mark marks the session for destruction.
func (s *Session) Mark() {
	s.marked = true
	s.cache.MarkSession(sessionExt{
		session: s,
	})
}

type sessionExt struct {
	session *Session
}

var _ SessionExt = sessionExt{}

func (s sessionExt) ID() string {
	return s.session.ID()
}

func (s sessionExt) Name() string {
	return s.session.Name()
}

func (s sessionExt) SetID(id string) {
	s.session.id = id
}

func (s sessionExt) GetValues() ValuesType {
	return s.session.values
}

func (s sessionExt) Session() *Session {
	return s.session
}

type sessionBuilder struct {
	codec Codec
}

var _ SessionBuilder = sessionBuilder{}

func (b sessionBuilder) NewSession(name string, cache Cache) SessionExt {
	return sessionExt{
		session: &Session{
			name:   name,
			id:     "",
			values: make(ValuesType),
			cache:  cache,
			codec:  b.codec,
			marked: false,
		},
	}
}

func (b sessionBuilder) NewExistingSession(name, id string, values ValuesType, cache Cache) SessionExt {
	return sessionExt{
		session: &Session{
			name:   name,
			id:     id,
			values: values,
			cache:  cache,
			codec:  b.codec,
			marked: false,
		},
	}
}
