package securecookie

import (
	gsc "github.com/gorilla/securecookie"

	"github.com/flatcar/nebraska/backend/pkg/sessions"
)

type codec struct {
	codecs []gsc.Codec
}

var _ sessions.Codec = &codec{}

func New(keyPairs ...[]byte) sessions.Codec {
	return &codec{
		codecs: gsc.CodecsFromPairs(keyPairs...),
	}
}

func (c *codec) Decode(name, value string) (id string, err error) {
	if err = gsc.DecodeMulti(name, value, &id, c.codecs...); err != nil {
		id = ""
	}
	return
}

func (c *codec) Encode(name, id string) (value string, err error) {
	value, err = gsc.EncodeMulti(name, id, c.codecs...)
	return
}
