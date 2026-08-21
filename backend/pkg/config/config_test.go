package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestValidateInstanceRetention(t *testing.T) {
	t.Run("disabled by default", func(t *testing.T) {
		c := &Config{AuthMode: "noop"}
		assert.NoError(t, c.Validate())
	})

	t.Run("enabled requires batch size", func(t *testing.T) {
		c := &Config{AuthMode: "noop", InstanceRetention: time.Hour}
		assert.Error(t, c.Validate())
	})

	t.Run("enabled with batch size", func(t *testing.T) {
		c := &Config{AuthMode: "noop", InstanceRetention: time.Hour, InstanceRetentionBatchSize: 500}
		assert.NoError(t, c.Validate())
	})

	t.Run("negative retention", func(t *testing.T) {
		c := &Config{AuthMode: "noop", InstanceRetention: -time.Hour, InstanceRetentionBatchSize: 500}
		assert.Error(t, c.Validate())
	})
}
