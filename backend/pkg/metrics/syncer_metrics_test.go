package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestSyncerLastSuccessTimestamp(t *testing.T) {
	SyncerLastSuccessTimestamp("stable", "amd64")

	got := testutil.ToFloat64(syncerLastSuccessTimestamp.WithLabelValues("stable", "amd64"))
	if got <= 0 {
		t.Fatalf("expected a positive unix timestamp, got %v", got)
	}
}

func TestSyncerCheckFailure(t *testing.T) {
	before := testutil.ToFloat64(syncerCheckFailuresTotal.WithLabelValues("beta", "arm64"))

	SyncerCheckFailure("beta", "arm64")

	after := testutil.ToFloat64(syncerCheckFailuresTotal.WithLabelValues("beta", "arm64"))
	if after != before+1 {
		t.Fatalf("expected counter to increment by 1, went from %v to %v", before, after)
	}
}

func TestSyncerCheckDuration(t *testing.T) {
	SyncerCheckDuration("alpha", "amd64", 1.5)

	count := testutil.CollectAndCount(syncerCheckDurationSeconds)
	if count == 0 {
		t.Fatal("expected histogram to have at least one observation registered")
	}
}

func TestSyncerPackageCreated(t *testing.T) {
	before := testutil.ToFloat64(syncerPackagesCreatedTotal.WithLabelValues("edge", "amd64"))

	SyncerPackageCreated("edge", "amd64")

	after := testutil.ToFloat64(syncerPackagesCreatedTotal.WithLabelValues("edge", "amd64"))
	if after != before+1 {
		t.Fatalf("expected counter to increment by 1, went from %v to %v", before, after)
	}
}