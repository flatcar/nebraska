package gob

import (
	"encoding"
	"encoding/binary"
	stdgob "encoding/gob"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kinvolk/nebraska/backend/pkg/sessions"
)

func TestMain(m *testing.M) {
	if os.Getenv("NEBRASKA_SKIP_TESTS") != "" {
		return
	}

	os.Exit(m.Run())
}

func TestNilTo(t *testing.T) {
	var to sessions.ValuesType = nil
	data := testData()
	c := New()
	assert.NoError(t, c.Copy(&to, data))
	assert.Equal(t, to, data)
}

func TestNotNilTo(t *testing.T) {
	to := sessions.ValuesType{}
	data := testData()
	c := New()
	assert.NoError(t, c.Copy(&to, data))
	assert.Equal(t, to, data)
}

type customData struct {
	Foo int
	Bar int
}

func TestCustomData(t *testing.T) {
	stdgob.Register(customData{})
	data := testData()
	data["custom"] = customData{
		Foo: 123,
		Bar: 456,
	}
	to := sessions.ValuesType{}
	c := New()
	assert.NoError(t, c.Copy(&to, data))
	assert.Equal(t, to, data)
}

type customDataWithPrivate struct {
	foo int
	Bar int
}

func (d customDataWithPrivate) toBytes() ([]byte, error) {
	b := make([]byte, 16)
	binary.BigEndian.PutUint64(b[0:], uint64(d.foo))
	binary.BigEndian.PutUint64(b[8:], uint64(d.Bar))
	return b, nil
}

func (d *customDataWithPrivate) fromBytes(b []byte) error {
	if len(b) != 16 {
		return fmt.Errorf("invalid byte length")
	}
	d.foo = int(binary.BigEndian.Uint64(b[0:]))
	d.Bar = int(binary.BigEndian.Uint64(b[8:]))
	return nil
}

type customDataGob struct {
	customDataWithPrivate
}

var _ stdgob.GobEncoder = customDataGob{}
var _ stdgob.GobDecoder = &customDataGob{}

func (d customDataGob) GobEncode() ([]byte, error) {
	return d.toBytes()
}

func (d *customDataGob) GobDecode(b []byte) error {
	return d.fromBytes(b)
}

func TestGobEncoderDecoder(t *testing.T) {
	stdgob.Register(customDataGob{})
	data := testData()
	data["custom"] = customDataGob{
		customDataWithPrivate: customDataWithPrivate{
			foo: 123,
			Bar: 456,
		},
	}
	to := sessions.ValuesType{}
	c := New()
	assert.NoError(t, c.Copy(&to, data))
	assert.Equal(t, to, data)
}

type customDataBinary struct {
	customDataWithPrivate
}

var _ encoding.BinaryMarshaler = customDataBinary{}
var _ encoding.BinaryUnmarshaler = &customDataBinary{}

func (d customDataBinary) MarshalBinary() ([]byte, error) {
	return d.toBytes()
}

func (d *customDataBinary) UnmarshalBinary(b []byte) error {
	return d.fromBytes(b)
}

func TestBinaryMarshallerUnmarshaller(t *testing.T) {
	stdgob.Register(customDataBinary{})
	data := testData()
	data["custom"] = customDataBinary{
		customDataWithPrivate: customDataWithPrivate{
			foo: 123,
			Bar: 456,
		},
	}
	to := sessions.ValuesType{}
	c := New()
	assert.NoError(t, c.Copy(&to, data))
	assert.Equal(t, to, data)
}

func testData() sessions.ValuesType {
	return sessions.ValuesType{
		"foo":    "bar",
		"answer": 42,
	}
}
