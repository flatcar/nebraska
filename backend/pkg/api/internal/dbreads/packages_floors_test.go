package dbreads

import (
	"strings"
	"testing"

	"github.com/doug-martin/goqu/v9"
	"github.com/stretchr/testify/require"

	"github.com/flatcar/nebraska/backend/pkg/api/types"
)

func TestSemverToIntArray(t *testing.T) {
	tests := []struct {
		name    string
		column  string
		want    string
		wantErr bool
	}{
		{
			name:   "column",
			column: "version",
			want:   "string_to_array((regexp_split_to_array(version, '[+-]'))[1], '.')::int[]",
		},
		{
			name:   "qualified column",
			column: "p.version",
			want:   "string_to_array((regexp_split_to_array(p.version, '[+-]'))[1], '.')::int[]",
		},
		{
			name:   "placeholder",
			column: "?",
			want:   "string_to_array((regexp_split_to_array(?, '[+-]'))[1], '.')::int[]",
		},
		{
			name:    "reject semicolon",
			column:  "p.version; drop table package",
			wantErr: true,
		},
		{
			name:    "reject nested qualification",
			column:  "public.p.version",
			wantErr: true,
		},
		{
			name:    "reject quoted identifier",
			column:  `p."version"`,
			wantErr: true,
		},
		{
			name:    "reject numeric suffix",
			column:  "version1",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := semverToIntArray(tt.column)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestVersionCompareExpr(t *testing.T) {
	tests := []struct {
		name     string
		column   string
		operator string
		value    string
		wantErr  bool
	}{
		{name: "greater than", column: "p.version", operator: ">", value: "1.2.3"},
		{name: "greater than or equal", column: "p.version", operator: ">=", value: "1.2.3"},
		{name: "less than", column: "p.version", operator: "<", value: "1.2.3"},
		{name: "less than or equal", column: "p.version", operator: "<=", value: "1.2.3"},
		{name: "equal", column: "p.version", operator: "=", value: "1.2.3"},
		{name: "not equal", column: "p.version", operator: "!=", value: "1.2.3"},
		{name: "reject invalid operator", column: "p.version", operator: "OR 1=1", value: "1.2.3", wantErr: true},
		{name: "reject invalid column", column: "p.version; drop table package", operator: ">", value: "1.2.3", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := versionCompareExpr(tt.column, tt.operator, tt.value)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			sql, args, err := goqu.From("package").Where(expr).Prepared(true).ToSQL()
			require.NoError(t, err)

			require.Contains(t, sql, tt.operator)
			require.Contains(t, sql, "string_to_array((regexp_split_to_array(p.version, '[+-]'))[1], '.')::int[]")
			require.Contains(t, sql, "string_to_array((regexp_split_to_array(?, '[+-]'))[1], '.')::int[]")
			require.NotContains(t, sql, tt.value)
			require.Equal(t, []interface{}{tt.value}, args)
		})
	}
}

func TestVersionCompareExprBindsMaliciousVersionValue(t *testing.T) {
	maliciousVersion := "1.2.3'); drop table package; --"

	expr, err := versionCompareExpr("p.version", ">", maliciousVersion)
	require.NoError(t, err)

	sql, args, err := goqu.From("package").Where(expr).Prepared(true).ToSQL()
	require.NoError(t, err)

	require.False(t, strings.Contains(sql, maliciousVersion))
	require.Equal(t, []interface{}{maliciousVersion}, args)
}

func TestGetRequiredChannelFloorsGuardClauses(t *testing.T) {
	q := &Queries{}

	floors, err := q.GetRequiredChannelFloors(nil, "1.0.0")
	require.ErrorIs(t, err, types.ErrNoPackageFound)
	require.Nil(t, floors)

	floors, err = q.GetRequiredChannelFloors(&types.Channel{}, "1.0.0")
	require.ErrorIs(t, err, types.ErrNoPackageFound)
	require.Nil(t, floors)

	floors, err = q.GetRequiredChannelFloors(&types.Channel{Package: &types.Package{Version: "2.0.0"}}, "")
	require.EqualError(t, err, "instance version cannot be empty")
	require.Nil(t, floors)
}
