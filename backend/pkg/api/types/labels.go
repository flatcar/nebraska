package types

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// ErrInvalidLabels indicates that a package's labels do not satisfy the
// limits checked by Labels.Validate.
var ErrInvalidLabels = errors.New("nebraska: invalid package labels")

const (
	// MaxLabels is the maximum number of labels on a package.
	MaxLabels = 16
	// MaxLabelKeyLength is the maximum length of a label key, in characters.
	MaxLabelKeyLength = 255
	// MaxLabelValueLength is the maximum length of a label value, in characters.
	MaxLabelValueLength = 1024
)

var labelKeyRegexp = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

// Labels are key/value strings attached to a package. Nebraska
// stores them and hands them to clients in the update response without
// interpreting them.
type Labels map[string]string

// Validate checks the number of labels, the characters and length of each
// key, and the length of each value.
func (l Labels) Validate() error {
	if len(l) > MaxLabels {
		return fmt.Errorf("%w: %d labels, the maximum is %d", ErrInvalidLabels, len(l), MaxLabels)
	}
	for k, v := range l {
		if len(k) > MaxLabelKeyLength || !labelKeyRegexp.MatchString(k) {
			return fmt.Errorf("%w: key %q must be 1 to %d characters from A-Z, a-z, 0-9, '.', '_' and '-'", ErrInvalidLabels, k, MaxLabelKeyLength)
		}
		// Values are sent as XML attributes, and the encoder replaces
		// characters XML cannot carry instead of failing.
		if !utf8.ValidString(v) || strings.IndexFunc(v, isNotXMLChar) >= 0 {
			return fmt.Errorf("%w: value of %q must be valid UTF-8 using only characters allowed in XML", ErrInvalidLabels, k)
		}
		if utf8.RuneCountInString(v) > MaxLabelValueLength {
			return fmt.Errorf("%w: value of %q is longer than %d characters", ErrInvalidLabels, k, MaxLabelValueLength)
		}
	}
	return nil
}

// isNotXMLChar reports whether r falls outside the Char production of
// XML 1.0, https://www.w3.org/TR/xml/#charsets.
func isNotXMLChar(r rune) bool {
	switch {
	case r == '\t' || r == '\n' || r == '\r':
		return false
	case r < 0x20 || r == 0xFFFE || r == 0xFFFF:
		return true
	}
	return false
}

// Value implements the driver.Valuer interface.
func (l Labels) Value() (driver.Value, error) {
	if l == nil {
		return "{}", nil
	}
	b, err := json.Marshal(map[string]string(l))
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// Scan implements the sql.Scanner interface.
func (l *Labels) Scan(src interface{}) error {
	var b []byte
	switch src := src.(type) {
	case []byte:
		b = src
	case string:
		b = []byte(src)
	case nil:
		*l = nil
		return nil
	default:
		return fmt.Errorf("cannot convert %T to Labels", src)
	}
	m := map[string]string{}
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	*l = m
	return nil
}
