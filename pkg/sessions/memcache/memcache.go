package memcache

import (
	"sync"

	"github.com/kinvolk/nebraska/pkg/random"
	"github.com/kinvolk/nebraska/pkg/sessions"
)

// ValuesCopier is used for copying values from session to cache
// during saving, and from cache back to session.
type ValuesCopier interface {
	Copy(to *sessions.ValuesType, from sessions.ValuesType) error
}

type sessionInfo struct {
	values  sessions.ValuesType
	uses    uint64
	destroy bool
}

type memCache struct {
	sessionsLock sync.Mutex
	sessions     map[string]*sessionInfo
	copier       ValuesCopier
	randomString func(int) string
}

var _ sessions.Cache = &memCache{}

// New returns a memory-based implementation of sessions.Cache.
func New(copier ValuesCopier) sessions.Cache {
	return newCache(copier)
}

func newCache(copier ValuesCopier) *memCache {
	return &memCache{
		sessions:     make(map[string]*sessionInfo),
		copier:       copier,
		randomString: random.String,
	}
}

// GetSessionUse is a part of sessions.Cache interface.
func (m *memCache) GetSessionUse(session sessions.SessionExt) {
	m.mutateSessionInfo(session.ID(), func(info *sessionInfo) {
		info.uses++
	})
}

// GetSessionUseByID is a part of sessions.Cache interface.
func (m *memCache) GetSessionUseByID(builder sessions.SessionBuilder, id, name string) *sessions.Session {
	var session *sessions.Session = nil
	m.mutateSessionInfo(id, func(info *sessionInfo) {
		if info.destroy {
			return
		}
		info.uses++
		session = m.newExistingSession(builder, id, name, info.values)
	})
	return session
}

func (m *memCache) newExistingSession(builder sessions.SessionBuilder, id, name string, values sessions.ValuesType) *sessions.Session {
	var valuesCopy sessions.ValuesType
	if err := m.copier.Copy(&valuesCopy, values); err != nil {
		// This shouldn't happen - copying of values succeeded
		// in copying values from session object to cache
		// before, so it would be strange for it to fail when
		// copying it back from cache to the session object.
		panic("Corrupted session values")
	}
	return builder.NewExistingSession(name, id, valuesCopy, m).Session()
}

// PutSessionUse is a part of sessions.Cache interface.
func (m *memCache) PutSessionUse(session sessions.SessionExt) {
	m.mutateSessionInfo(session.ID(), func(info *sessionInfo) {
		if info.uses == 0 {
			panic("mismatched GetSessionUse and PutSessionUse")
		}
		info.uses--
		if info.destroy && info.uses == 0 {
			delete(m.sessions, session.ID())
		}
	})
}

// MarkOrDestroySessionByID is a part of sessions.Cache interface.
func (m *memCache) MarkOrDestroySessionByID(id string) {
	m.mutateSessionInfo(id, func(info *sessionInfo) {
		if info.uses > 0 {
			info.destroy = true
		} else {
			delete(m.sessions, id)
		}
	})
}

// MarkSession is a part of sessions.Cache interface.
func (m *memCache) MarkSession(session sessions.SessionExt) {
	m.mutateSessionInfo(session.ID(), func(info *sessionInfo) {
		info.destroy = true
	})
}

type mutateInfoCallback func(info *sessionInfo)

func (m *memCache) mutateSessionInfo(id string, callback mutateInfoCallback) {
	if id == "" {
		return
	}
	m.sessionsLock.Lock()
	defer m.sessionsLock.Unlock()
	if info, ok := m.sessions[id]; ok {
		callback(info)
	}
}

// SaveSession is a part of sessions.Cache interface.
func (m *memCache) SaveSession(session sessions.SessionExt) (bool, error) {
	m.sessionsLock.Lock()
	defer m.sessionsLock.Unlock()
	if info, ok := m.sessions[session.ID()]; ok {
		if info.destroy {
			return true, nil
		}
		if err := m.storeValues(session, info); err != nil {
			return false, err
		}
	} else {
		// first save ever
		session.SetID(m.randomString(64))
		info := &sessionInfo{
			values:  nil,
			uses:    1,
			destroy: false,
		}
		if err := m.storeValues(session, info); err != nil {
			return false, err
		}
		m.sessions[session.ID()] = info
	}
	return false, nil
}

func (m *memCache) storeValues(session sessions.SessionExt, info *sessionInfo) error {
	var values sessions.ValuesType
	if err := m.copier.Copy(&values, session.GetValues()); err != nil {
		return err
	}
	info.values = values
	return nil
}
