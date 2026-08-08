package random

import (
	"errors"
	"testing"
)

// TestDataPanicsOnReadError documents the current behaviour of Data(): if the
// underlying rand.Read call returns an error, Data() panics instead of
// propagating the error to the caller.
//
// This test exists to make the behaviour explicit and to serve as a regression
// anchor while a proper fix (returning an error instead of panicking) is
// implemented. See the companion issue for the proposed DataE() API.
func TestDataPanicsOnReadError(t *testing.T) {
	orig := randRead
	defer func() { randRead = orig }()

	// Inject a failing rand.Read
	randRead = func(b []byte) (int, error) {
		return 0, errors.New("simulated entropy source failure")
	}

	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected Data() to panic on rand.Read error, but it did not")
		}
	}()

	// This call should panic with the injected error
	Data(16)
}

// TestDataErrorVariantProposal shows what the API would look like if Data()
// returned an error instead of panicking. This is the proposed DataE() function
// that callers could use when they want to handle entropy failures gracefully
// rather than crashing the process.
func TestDataErrorVariantProposal(t *testing.T) {
	orig := randRead
	defer func() { randRead = orig }()

	// Test with normal read - should succeed
	randRead = testRandRead
	data, err := dataE(16)
	if err != nil {
		t.Errorf("dataE() returned unexpected error: %v", err)
	}
	if len(data) != 16 {
		t.Errorf("dataE() returned %d bytes, want 16", len(data))
	}

	// Test with failing read - should return error, not panic
	randRead = func(b []byte) (int, error) {
		return 0, errors.New("simulated entropy source failure")
	}
	_, err = dataE(16)
	if err == nil {
		t.Error("dataE() should return error when rand.Read fails")
	}
}

// dataE is the proposed error-returning variant of Data().
// It returns (nil, err) when rand.Read fails instead of panicking.
// This demonstrates the fix without breaking existing callers of Data().
func dataE(n int) ([]byte, error) {
	if n < 1 {
		return nil, nil
	}
	data := make([]byte, n)
	if _, err := randRead(data); err != nil {
		return nil, err
	}
	return data, nil
}
