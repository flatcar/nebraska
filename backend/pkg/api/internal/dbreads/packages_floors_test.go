package dbreads

import (
	"errors"
	"testing"

	"github.com/doug-martin/goqu/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flatcar/nebraska/backend/pkg/api/internal/types"
)

func TestSemverToIntArray(t *testing.T) {
	validCases := []struct {
		name   string
		column string
		want   string
	}{
		{
			name:   "qualified column",
			column: "p.version",
			want:   "string_to_array((regexp_split_to_array(p.version, '[+-]'))[1], '.')::int[]",
		},
		{
			name:   "simple column",
			column: "version",
			want:   "string_to_array((regexp_split_to_array(version, '[+-]'))[1], '.')::int[]",
		},
		{
			name:   "short column",
			column: "id",
			want:   "string_to_array((regexp_split_to_array(id, '[+-]'))[1], '.')::int[]",
		},
		{
			name:   "placeholder for bound value",
			column: "?",
			want:   "string_to_array((regexp_split_to_array(?, '[+-]'))[1], '.')::int[]",
		},
	}

	for _, tt := range validCases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := semverToIntArray(tt.column)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}

	invalidCases := []struct {
		name   string
		column string
	}{
		{name: "sql injection semicolon", column: "version; DROP TABLE users;--"},
		{name: "sql injection equality", column: "1=1"},
		{name: "sql injection closing paren", column: "version)"},
		{name: "sql injection quote escape", column: "' OR '1'='1"},
		// Known regex strictness limitation: lowercase-only identifiers reject
		// otherwise-safe names that contain uppercase letters or digits.
		{name: "uppercase letter rejected by regex", column: "Version"},
		{name: "digit in name rejected by regex", column: "col1"},
	}

	for _, tt := range invalidCases {
		t.Run(tt.name, func(t *testing.T) {
			_, err := semverToIntArray(tt.column)
			require.Error(t, err)
		})
	}
}

func TestVersionCompareExpr(t *testing.T) {
	validOperators := []string{">", ">=", "<", "<=", "=", "!="}

	for _, op := range validOperators {
		t.Run("valid operator "+op, func(t *testing.T) {
			expr, err := versionCompareExpr("p.version", op, "1.0.0")
			require.NoError(t, err)
			require.NotNil(t, expr)
		})
	}

	invalidOperators := []struct {
		name     string
		operator string
	}{
		{name: "LIKE", operator: "LIKE"},
		{name: "semicolon injection", operator: "; DROP TABLE"},
		{name: "double greater-than", operator: ">>"},
		{name: "empty string", operator: ""},
	}

	for _, tt := range invalidOperators {
		t.Run("invalid operator "+tt.name, func(t *testing.T) {
			expr, err := versionCompareExpr("p.version", tt.operator, "1.0.0")
			require.Error(t, err)
			assert.Nil(t, expr)
		})
	}

	t.Run("invalid column propagates error", func(t *testing.T) {
		expr, err := versionCompareExpr("version; DROP TABLE", ">", "1.0.0")
		require.Error(t, err)
		assert.Nil(t, expr)
		assert.Contains(t, err.Error(), "semverToIntArray")
	})

	t.Run("value is parameterized not concatenated", func(t *testing.T) {
		dangerousValue := "1.0.0'; DROP TABLE package;--"

		expr, err := versionCompareExpr("p.version", ">", dangerousValue)
		require.NoError(t, err)
		require.NotNil(t, expr)

		sql, args, err := goqu.From("package").Prepared(true).Where(expr).ToSQL()
		require.NoError(t, err)
		assert.NotContains(t, sql, dangerousValue)
		assert.Contains(t, args, dangerousValue)
	})
}

func TestGetRequiredChannelFloorsGuardClauses(t *testing.T) {
	q := &Queries{maxFloorsPerResponse: 5}

	t.Run("nil channel", func(t *testing.T) {
		pkgs, err := q.GetRequiredChannelFloors(nil, "1.0.0")
		require.Error(t, err)
		assert.Nil(t, pkgs)
		assert.True(t, errors.Is(err, ErrNoPackageFound))
	})

	t.Run("nil channel package", func(t *testing.T) {
		channel := &types.Channel{ID: "channel-1"}
		pkgs, err := q.GetRequiredChannelFloors(channel, "1.0.0")
		require.Error(t, err)
		assert.Nil(t, pkgs)
		assert.True(t, errors.Is(err, ErrNoPackageFound))
	})

	t.Run("empty instance version", func(t *testing.T) {
		channel := &types.Channel{
			ID: "channel-1",
			Package: &types.Package{
				Version: "2.0.0",
			},
		}
		pkgs, err := q.GetRequiredChannelFloors(channel, "")
		require.Error(t, err)
		assert.Nil(t, pkgs)
		assert.EqualError(t, err, "instance version cannot be empty")
	})
}
