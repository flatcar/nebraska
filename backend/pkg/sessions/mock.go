package sessions

import (
	"fmt"
	"strconv"
)

type cacheNode struct {
	mark   bool
	values ValuesType
	uses   int
}

// MockCache is an implementation of Cache interface used for mocking.
type MockCache struct {
	nodes    map[string]*cacheNode
	idPrefix string
	nextID   int
}

var _ Cache = &MockCache{}

// New creates a new MockCache.
func NewMockCache() *MockCache {
	return &MockCache{
		nodes:    make(map[string]*cacheNode),
		idPrefix: "id",
		nextID:   0,
	}
}

// UseCountFor returns a use count of the session with a passed ID, or
// -1 the session is not in cache.
func (c *MockCache) UseCountFor(id string) int {
	node, ok := c.nodes[id]
	if !ok {
		return -1
	}
	return node.uses
}

// GetSessionUse is a part of Cache interface.
func (c *MockCache) GetSessionUse(session SessionExt) {
	if node, ok := c.nodes[session.ID()]; ok {
		node.uses++
	}
}

// GetSessionUseByID is a part of Cache interface.
func (c *MockCache) GetSessionUseByID(builder SessionBuilder, id, name string) *Session {
	values := ValuesType{}
	if node, ok := c.nodes[id]; ok {
		if node.mark {
			return nil
		}
		node.uses++
		copyValues(values, node.values)
	}
	return builder.NewExistingSession(name, id, values, c).Session()
}

// PutSessionUse is a part of Cache interface.
func (c *MockCache) PutSessionUse(session SessionExt) {
	if node, ok := c.nodes[session.ID()]; ok {
		node.uses--
		if node.uses == 0 && node.mark {
			delete(c.nodes, session.ID())
		}
	}
}

// MarkOrDestroySessionByID is a part of Cache interface.
func (c *MockCache) MarkOrDestroySessionByID(id string) {
	if node, ok := c.nodes[id]; ok {
		if node.uses > 0 {
			node.mark = true
		} else {
			delete(c.nodes, id)
		}
	}
}

// MarkSession is a part of Cache interface.
func (c *MockCache) MarkSession(session SessionExt) {
	if node, ok := c.nodes[session.ID()]; ok {
		node.mark = true
	}
}

// SaveSession is a part of Cache interface.
func (c *MockCache) SaveSession(session SessionExt) (bool, error) {
	node, ok := c.nodes[session.ID()]
	if ok {
		if node.mark {
			return true, nil
		}
	} else {
		id := c.getNextID()
		session.SetID(id)
		node = &cacheNode{
			mark:   false,
			values: make(ValuesType, len(session.GetValues())),
			uses:   1,
		}
		c.nodes[id] = node
	}
	copyValues(node.values, session.GetValues())
	return false, nil
}

func (c *MockCache) getNextID() string {
	c.nextID++
	return c.idPrefix + strconv.Itoa(c.nextID)
}

func copyValues(to, from ValuesType) {
	for k, v := range from {
		to[k] = v
	}
}

// MockCodec is an implementation of Codec used for mocking.
type MockCodec struct {
	idsToValues map[string]string
	valuesToIDs map[string]string
}

var _ Codec = &MockCodec{}

// NewMockCodec creates a new MockCodec.
func NewMockCodec() *MockCodec {
	return &MockCodec{
		idsToValues: make(map[string]string),
		valuesToIDs: make(map[string]string),
	}
}

// AddIDValueMapping adds a mapping from ID to cookie value and
// back. Not adding mapping will result in an error when encoding or
// decoding.
func (c *MockCodec) AddIDValueMapping(idValuePairs ...string) {
	if len(idValuePairs)%2 == 1 {
		panic(fmt.Sprintf("got an odd number of arguments (%d), while expected pairs of IDs and values", len(idValuePairs)))
	}
	for idx := 0; idx < len(idValuePairs); idx += 2 {
		id := idValuePairs[idx]
		value := idValuePairs[idx+1]
		c.idsToValues[id] = value
		c.valuesToIDs[value] = id
	}
}

// Decode is a part of Codec interface.
func (c *MockCodec) Decode(name, value string) (string, error) {
	if id, ok := c.valuesToIDs[value]; ok {
		return id, nil
	}
	return "", fmt.Errorf("no mapped id for name %q and value %q", name, value)
}

// Encode is a part of Codec interface.
func (c *MockCodec) Encode(name, id string) (string, error) {
	if value, ok := c.idsToValues[id]; ok {
		return value, nil
	}
	return "", fmt.Errorf("no mapped value for name %q and id %q", name, id)
}
