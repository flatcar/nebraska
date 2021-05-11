package sessions

import (
	"net/http"
)

// Codec decodes cookie ID from the name and value pair, and encodes
// cookie value from name and id pair. This type is used by the store.
type Codec interface {
	// Decode should return an ID from the passed name and cookie
	// value.
	Decode(name, value string) (string, error)
	// Encode should return a cookie value from the passed name
	// and ID.
	Encode(name, id string) (string, error)
}

// Cache stores the refcounted session information. This type
// shouldn't be used by libraries - it's only for Store and for
// implementations of this interface.
type Cache interface {
	// GetSessionUse should increase some use count on data the
	// passed session refers to.
	GetSessionUse(session SessionExt)
	// GetSessionUseByID should use the passed builder to create a
	// session object with a use count on the data incremented by
	// one.
	GetSessionUseByID(builder SessionBuilder, id, name string) *Session
	// PutSessionUse should decrement the use count on data and
	// destroy the data if the count drops to zero and the session
	// was marked for destruction.
	PutSessionUse(session SessionExt)
	// MarkOrDestroySessionByID should mark the data for
	// destruction. If data's use count is zero, it should be
	// destroyed.
	MarkOrDestroySessionByID(id string)
	// MarkSession marks the session for destruction. Data
	// shouldn't be destroyed since the existence of the passed
	// session object means that the data is still used.
	MarkSession(session SessionExt)
	// SaveSession should persist the changes made in the session
	// object to the cache.
	SaveSession(session SessionExt) (bool, error)
}

// Store contains cache and codec for session management.
type Store struct {
	cache Cache
	codec Codec
}

// NewStore creates a new sessions store.
func NewStore(cache Cache, codec Codec) *Store {
	return &Store{
		cache: cache,
		codec: codec,
	}
}

// MarkOrDestroySessionByID marks the session with a passed ID for
// destruction. If the session is not currently in use by any request,
// the session should be destroyed immediately.
func (s *Store) MarkOrDestroySessionByID(id string) {
	s.cache.MarkOrDestroySessionByID(id)
}

// GetSessionUse tries to either retrieve the session from either
// context or request, or create a new one. The returned session has
// its use count increased, so is safe from being destroyed until
// PutSessionUse is called.
func (s *Store) GetSessionUse(request *http.Request, name string) *Session {
	if session := SessionFromContext(request.Context()); session != nil {
		s.cache.GetSessionUse(sessionExt{
			session: session,
		})
		return session
	}
	builder := sessionBuilder{
		codec: s.codec,
	}
	if id, err := s.getSessionIDFromRequest(request, name); err == nil {
		if session := s.cache.GetSessionUseByID(builder, id, name); session != nil {
			return session
		}
	}
	session := builder.NewSession(name, s.cache)
	s.cache.GetSessionUse(session)
	return session.Session()
}

// PutSessionUse decreases the use count of the session and
// potentially destroys it, if it was marked.
func (s *Store) PutSessionUse(session *Session) {
	s.cache.PutSessionUse(sessionExt{
		session: session,
	})
}

func (s *Store) getSessionIDFromRequest(request *http.Request, name string) (string, error) {
	cookie, err := request.Cookie(name)
	if err != nil {
		return "", err
	}
	return s.codec.Decode(name, cookie.Value)
}
