package dbreads

import (
	"testing"
)

// BenchmarkDurationParamToPostgresTimings benchmarks the conversion of duration
// parameters to PostgreSQL timing strings. This helps measure the overhead of
// duration parsing in query preparation.
func BenchmarkDurationParamToPostgresTimings(b *testing.B) {
	benchmarks := []struct {
		name  string
		input durationParam
	}{
		{"1h", "1h"},
		{"1d", "1d"},
		{"7d", "7d"},
		{"30d", "30d"},
		{"invalid", "invalid"},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _, _ = durationParamToPostgresTimings(bm.input)
			}
		})
	}
}

// BenchmarkDurationCodeToPostgresTimings benchmarks the conversion of duration
// codes to PostgreSQL timing strings. This measures the performance of internal
// duration code conversion.
func BenchmarkDurationCodeToPostgresTimings(b *testing.B) {
	codes := []struct {
		name string
		code durationCode
	}{
		{"one_hour", oneHour},
		{"one_day", oneDay},
		{"seven_days", sevenDays},
		{"thirty_days", thirtyDays},
	}

	for _, bm := range codes {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _, _ = durationCodeToPostgresTimings(bm.code)
			}
		})
	}
}

// BenchmarkIsNightlyVersion benchmarks version string detection for nightly builds.
// This is called frequently during instance filtering and version breakdown queries.
func BenchmarkIsNightlyVersion(b *testing.B) {
	versions := []struct {
		name    string
		version string
	}{
		{"stable", "3815.2.0"},
		{"beta", "3850.0.0-beta.2"},
		{"nightly", "3900.0.0-nightly-20230415-1234"},
		{"alpha_nightly", "3850.0.0-alpha.1-nightly-2023"},
		{"long_stable", "3815.2.0+20230415-1234-git-abcd1234"},
		{"empty", ""},
	}

	for _, bm := range versions {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = isNightlyVersion(bm.version)
			}
		})
	}
}

// BenchmarkValidatePaginationParams benchmarks pagination parameter validation
// and default value assignment. This is called on every paginated API request.
func BenchmarkValidatePaginationParams(b *testing.B) {
	cases := []struct {
		name    string
		page    uint64
		perPage uint64
	}{
		{"valid_params", 1, 10},
		{"zero_page", 0, 10},
		{"zero_perPage", 1, 0},
		{"both_zero", 0, 0},
		{"large_values", 1000, 100},
	}

	for _, bm := range cases {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = validatePaginationParams(bm.page, bm.perPage)
			}
		})
	}
}

// BenchmarkSqlPaginate benchmarks the SQL LIMIT/OFFSET calculation for pagination.
// This is executed on every paginated query to compute database pagination parameters.
func BenchmarkSqlPaginate(b *testing.B) {
	cases := []struct {
		name    string
		page    uint64
		perPage uint64
	}{
		{"first_page", 1, 10},
		{"tenth_page", 10, 10},
		{"large_page", 100, 50},
		{"small_perPage", 5, 5},
		{"default_pagination", 1, 10},
	}

	for _, bm := range cases {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = sqlPaginate(bm.page, bm.perPage)
			}
		})
	}
}

// BenchmarkIgnoreFakeInstanceCondition benchmarks the SQL condition generation
// for filtering fake/test instances. This condition is included in most instance
// queries to exclude test data from production metrics.
func BenchmarkIgnoreFakeInstanceCondition(b *testing.B) {
	fields := []struct {
		name  string
		field string
	}{
		{"simple_field", "instance_id"},
		{"aliased_field", "ia.instance_id"},
		{"table_qualified", "instance_application.instance_id"},
	}

	for _, bm := range fields {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = ignoreFakeInstanceCondition(bm.field)
			}
		})
	}
}

// BenchmarkGroupDurationCacheKey benchmarks cache key generation for group version
// count queries. Cache key generation happens on every GetGroupVersionCountTimeline call.
func BenchmarkGroupDurationCacheKey(b *testing.B) {
	groupIDs := []string{
		"e96281a6-d1af-4bde-9a0a-97b76e56dc57",
		"short-id",
		"very-long-group-identifier-with-many-characters-for-testing",
	}
	durations := []string{"1h", "1d", "7d", "30d"}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		groupID := groupIDs[i%len(groupIDs)]
		duration := durations[i%len(durations)]
		_ = groupDurationCacheKey{
			GroupID:  groupID,
			Duration: duration,
		}
	}
}

// BenchmarkDurationParamToCodeLookup benchmarks the map lookup performance for
// converting duration parameters to duration codes. This helps measure the
// overhead of the durationParamToCode map access.
func BenchmarkDurationParamToCodeLookup(b *testing.B) {
	params := []durationParam{"1h", "1d", "7d", "30d", "invalid"}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		param := params[i%len(params)]
		_, _ = durationParamToCode[param]
	}
}
