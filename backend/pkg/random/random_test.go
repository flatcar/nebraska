package random

import (
	"errors"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	if os.Getenv("NEBRASKA_SKIP_TESTS") != "" {
		return
	}

	randRead = testRandRead

	os.Exit(m.Run())
}

// testRandRead fills the passed slice with repeated sequences of
// digits from 9 to 0.
func testRandRead(data []byte) (int, error) {
	for idx := 0; idx < len(data); idx++ {
		data[idx] = byte('9' - idx%10)
	}
	return len(data), nil
}

func TestRandomData(t *testing.T) {
	for _, tt := range []struct {
		n      int
		output []byte
	}{
		{
			n:      -1,
			output: nil,
		},
		{
			n:      0,
			output: nil,
		},
		{
			n:      1,
			output: []byte{'9'},
		},
		{
			n:      2,
			output: []byte{'9', '8'},
		},
		{
			n:      12,
			output: []byte{'9', '8', '7', '6', '5', '4', '3', '2', '1', '0', '9', '8'},
		},
	} {
		got := Data(tt.n)
		if !byteSliceEqual(tt.output, got) {
			t.Errorf("For n = %d, expected %#v, got %#v", tt.n, tt.output, got)
		}

		gotE, err := DataE(tt.n)
		if err != nil {
			t.Errorf("DataE returned unexpected error for n = %d: %v", tt.n, err)
		}
		if !byteSliceEqual(tt.output, gotE) {
			t.Errorf("DataE for n = %d, expected %#v, got %#v", tt.n, tt.output, gotE)
		}
	}
}

func TestRandomString(t *testing.T) {
	for _, tt := range []struct {
		n      int
		output string
	}{
		{
			n:      -1,
			output: "",
		},
		{
			n:      0,
			output: "",
		},
		{
			n:      1,
			output: "H",
		},
		{
			n:      2,
			output: "HE",
		},
		{
			n:      8,
			output: "HE4DONRV",
		},
		{
			n:      9,
			output: "HE4DONRVG",
		},
		{
			n:      12,
			output: "HE4DONRVGQZT",
		},
		{
			n:      16,
			output: "HE4DONRVGQZTEMJQ",
		},
		{
			n:      17,
			output: "HE4DONRVGQZTEMJQH",
		},
	} {
		got := String(tt.n)
		if tt.output != got {
			t.Errorf("For n = %d, expected %s, got %s", tt.n, tt.output, got)
		}

		gotE, err := StringE(tt.n)
		if err != nil {
			t.Errorf("StringE returned unexpected error for n = %d: %v", tt.n, err)
		}
		if tt.output != gotE {
			t.Errorf("StringE for n = %d, expected %s, got %s", tt.n, tt.output, gotE)
		}
	}
}

func TestRandomDataEError(t *testing.T) {
	origRandRead := randRead
	defer func() { randRead = origRandRead }()

	errMock := errors.New("mock randRead failure")
	randRead = func(b []byte) (int, error) {
		return 0, errMock
	}

	_, err := DataE(10)
	if !errors.Is(err, errMock) {
		t.Errorf("expected error %v, got %v", errMock, err)
	}

	_, err = StringE(10)
	if !errors.Is(err, errMock) {
		t.Errorf("expected error %v, got %v", errMock, err)
	}
}

func byteSliceEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for idx := 0; idx < len(a); idx++ {
		if a[idx] != b[idx] {
			return false
		}
	}
	return true
}
