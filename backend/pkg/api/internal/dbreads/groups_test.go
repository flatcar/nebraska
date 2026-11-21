package dbreads

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flatcar/nebraska/backend/pkg/api/internal/types"
)

// TestDurationParamToPostgresTimings tests conversion of duration parameters
// to PostgreSQL interval strings.
func TestDurationParamToPostgresTimings(t *testing.T) {
	tests := []struct {
		name           string
		input          durationParam
		wantDuration   postgresDuration
		wantInterval   postgresInterval
		wantErr        bool
		errDescription string
	}{
		{
			name:         "one hour duration",
			input:        "1h",
			wantDuration: "1hour",
			wantInterval: "15 minute",
			wantErr:      false,
		},
		{
			name:         "one day duration",
			input:        "1d",
			wantDuration: "1days",
			wantInterval: "1 hour",
			wantErr:      false,
		},
		{
			name:         "seven days duration",
			input:        "7d",
			wantDuration: "7 days",
			wantInterval: "1 days",
			wantErr:      false,
		},
		{
			name:         "thirty days duration",
			input:        "30d",
			wantDuration: "30 days",
			wantInterval: "3 days",
			wantErr:      false,
		},
		{
			name:           "invalid duration parameter",
			input:          "invalid",
			wantDuration:   "",
			wantInterval:   "",
			wantErr:        true,
			errDescription: "should error on invalid duration param",
		},
		{
			name:           "empty duration parameter",
			input:          "",
			wantDuration:   "",
			wantInterval:   "",
			wantErr:        true,
			errDescription: "should error on empty duration param",
		},
		{
			name:           "numeric duration parameter",
			input:          "123",
			wantDuration:   "",
			wantInterval:   "",
			wantErr:        true,
			errDescription: "should error on invalid numeric format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDuration, gotInterval, err := durationParamToPostgresTimings(tt.input)

			if tt.wantErr {
				assert.Error(t, err, tt.errDescription)
				assert.Empty(t, gotDuration)
				assert.Empty(t, gotInterval)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantDuration, gotDuration)
				assert.Equal(t, tt.wantInterval, gotInterval)
			}
		})
	}
}

// TestDurationCodeToPostgresTimings tests conversion of duration codes
// to PostgreSQL timing strings.
func TestDurationCodeToPostgresTimings(t *testing.T) {
	tests := []struct {
		name         string
		code         durationCode
		wantDuration postgresDuration
		wantInterval postgresInterval
		wantErr      bool
	}{
		{
			name:         "one hour code",
			code:         oneHour,
			wantDuration: "1hour",
			wantInterval: "15 minute",
			wantErr:      false,
		},
		{
			name:         "one day code",
			code:         oneDay,
			wantDuration: "1days",
			wantInterval: "1 hour",
			wantErr:      false,
		},
		{
			name:         "seven days code",
			code:         sevenDays,
			wantDuration: "7 days",
			wantInterval: "1 days",
			wantErr:      false,
		},
		{
			name:         "thirty days code",
			code:         thirtyDays,
			wantDuration: "30 days",
			wantInterval: "3 days",
			wantErr:      false,
		},
		{
			name:         "invalid duration code",
			code:         durationCode(999),
			wantDuration: "",
			wantInterval: "",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDuration, gotInterval, err := durationCodeToPostgresTimings(tt.code)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, gotDuration)
				assert.Empty(t, gotInterval)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantDuration, gotDuration)
				assert.Equal(t, tt.wantInterval, gotInterval)
			}
		})
	}
}

// TestIsNightlyVersion tests nightly version detection.
func TestIsNightlyVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    bool
	}{
		{
			name:    "stable version",
			version: "3815.2.0",
			want:    false,
		},
		{
			name:    "beta version",
			version: "3913.0.0-beta",
			want:    false,
		},
		{
			name:    "nightly version",
			version: "4000.0.0-nightly-20260730",
			want:    true,
		},
		{
			name:    "nightly in middle",
			version: "1.0.0-nightly",
			want:    true,
		},
		{
			name:    "nightly with metadata",
			version: "2.0.0-nightly+build123",
			want:    true,
		},
		{
			name:    "empty version",
			version: "",
			want:    false,
		},
		{
			name:    "version with nightly substring",
			version: "3.0.nightly.0",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isNightlyVersion(tt.version)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestGroupDurationCacheKey tests the cache key struct used for version count caching.
func TestGroupDurationCacheKey(t *testing.T) {
	key1 := groupDurationCacheKey{
		GroupID:  "group-123",
		Duration: "7d",
	}
	key2 := groupDurationCacheKey{
		GroupID:  "group-123",
		Duration: "7d",
	}
	key3 := groupDurationCacheKey{
		GroupID:  "group-456",
		Duration: "7d",
	}
	key4 := groupDurationCacheKey{
		GroupID:  "group-123",
		Duration: "1d",
	}

	// Test key equality for map usage
	m := make(map[groupDurationCacheKey]string)
	m[key1] = "value1"

	assert.Equal(t, "value1", m[key2], "identical keys should match")
	assert.NotContains(t, m, key3, "different group ID should not match")
	assert.NotContains(t, m, key4, "different duration should not match")
}

// TestGroupVersionCountCacheExpiry tests cache expiration logic.
func TestGroupVersionCountCacheExpiry(t *testing.T) {
	now := time.Now().UTC()
	tests := []struct {
		name       string
		storedAt   time.Time
		lifespan   time.Duration
		wantExpiry bool
	}{
		{
			name:       "recently stored cache",
			storedAt:   now.Add(-30 * time.Second),
			lifespan:   time.Minute,
			wantExpiry: false,
		},
		{
			name:       "expired cache",
			storedAt:   now.Add(-2 * time.Minute),
			lifespan:   time.Minute,
			wantExpiry: true,
		},
		{
			name:       "cache stored just past lifespan",
			storedAt:   now.Add(-61 * time.Second),
			lifespan:   time.Minute,
			wantExpiry: true,
		},
		{
			name:       "freshly stored cache",
			storedAt:   now,
			lifespan:   time.Minute,
			wantExpiry: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := groupVersionCountCache{
				data:     make(map[time.Time]types.VersionCountMap),
				storedAt: tt.storedAt,
			}

			isExpired := time.Since(cache.storedAt) > tt.lifespan
			assert.Equal(t, tt.wantExpiry, isExpired)
		})
	}
}

// TestIgnoreFakeInstanceCondition tests the SQL condition for filtering fake instances.
func TestIgnoreFakeInstanceCondition(t *testing.T) {
	tests := []struct {
		name          string
		fieldName     string
		wantContains  []string
		wantCondition string
	}{
		{
			name:      "instance_id field",
			fieldName: "instance_id",
			wantContains: []string{
				"instance_id",
				"IS NULL",
				"NOT LIKE",
				"{________-____-____-____-____________}",
			},
			wantCondition: "(instance_id IS NULL OR instance_id NOT LIKE '{________-____-____-____-____________}')",
		},
		{
			name:      "i.instance_id field with alias",
			fieldName: "i.instance_id",
			wantContains: []string{
				"i.instance_id",
				"IS NULL",
				"NOT LIKE",
			},
			wantCondition: "(i.instance_id IS NULL OR i.instance_id NOT LIKE '{________-____-____-____-____________}')",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ignoreFakeInstanceCondition(tt.fieldName)

			assert.Equal(t, tt.wantCondition, got)

			for _, substr := range tt.wantContains {
				assert.Contains(t, got, substr,
					"condition should contain %q", substr)
			}
		})
	}
}

// TestValidatePaginationParams tests pagination parameter validation.
func TestValidatePaginationParams(t *testing.T) {
	tests := []struct {
		name        string
		page        uint64
		perPage     uint64
		wantPage    uint64
		wantPerPage uint64
	}{
		{
			name:        "valid parameters",
			page:        5,
			perPage:     20,
			wantPage:    5,
			wantPerPage: 20,
		},
		{
			name:        "zero page defaults to 1",
			page:        0,
			perPage:     10,
			wantPage:    1,
			wantPerPage: 10,
		},
		{
			name:        "zero perPage defaults to 10",
			page:        1,
			perPage:     0,
			wantPage:    1,
			wantPerPage: 10,
		},
		{
			name:        "both zero use defaults",
			page:        0,
			perPage:     0,
			wantPage:    1,
			wantPerPage: 10,
		},
		{
			name:        "large page number",
			page:        1000,
			perPage:     50,
			wantPage:    1000,
			wantPerPage: 50,
		},
		{
			name:        "max perPage limit",
			page:        1,
			perPage:     1000,
			wantPage:    1,
			wantPerPage: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPage, gotPerPage := validatePaginationParams(tt.page, tt.perPage)
			assert.Equal(t, tt.wantPage, gotPage)
			assert.Equal(t, tt.wantPerPage, gotPerPage)
		})
	}
}

// TestSqlPaginate tests SQL pagination offset calculation.
func TestSqlPaginate(t *testing.T) {
	tests := []struct {
		name       string
		page       uint64
		perPage    uint64
		wantLimit  uint
		wantOffset uint
	}{
		{
			name:       "first page",
			page:       1,
			perPage:    10,
			wantLimit:  10,
			wantOffset: 0,
		},
		{
			name:       "second page",
			page:       2,
			perPage:    10,
			wantLimit:  10,
			wantOffset: 10,
		},
		{
			name:       "third page with 20 per page",
			page:       3,
			perPage:    20,
			wantLimit:  20,
			wantOffset: 40,
		},
		{
			name:       "page 10 with 50 per page",
			page:       10,
			perPage:    50,
			wantLimit:  50,
			wantOffset: 450,
		},
		{
			name:       "single item per page",
			page:       5,
			perPage:    1,
			wantLimit:  1,
			wantOffset: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLimit, gotOffset := sqlPaginate(tt.page, tt.perPage)
			assert.Equal(t, tt.wantLimit, gotLimit, "limit mismatch")
			assert.Equal(t, tt.wantOffset, gotOffset, "offset mismatch")
		})
	}
}
