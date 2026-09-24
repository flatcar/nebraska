/*
Copyright 2025 The Nebraska Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package api

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestAcquireSyncLeadership_ConcurrentReplicas verifies that when two
// "replicas" (independent API instances sharing the same underlying
// database) race to acquire sync leadership concurrently, exactly one
// wins. This is the scenario behind #388: previously every replica ran
// its own independent syncer with no coordination, causing concurrent
// writes and duplicate-key violations under HA deployments.
func TestAcquireSyncLeadership_ConcurrentReplicas(t *testing.T) {
	replicaA := newForTest(t)
	defer replicaA.Close()
	replicaB := newForTest(t)
	defer replicaB.Close()

	var wg sync.WaitGroup
	results := make([]bool, 2)
	leaderships := make([]*SyncLeadership, 2)
	errs := make([]error, 2)

	acquire := func(i int, a *API) {
		defer wg.Done()
		l, acquired, err := a.AcquireSyncLeadership(context.Background())
		results[i] = acquired
		errs[i] = err
		if acquired {
			leaderships[i] = l
		}
	}

	wg.Add(2)
	go acquire(0, replicaA)
	go acquire(1, replicaB)
	wg.Wait()

	require.NoError(t, errs[0])
	require.NoError(t, errs[1])
	require.NotEqual(t, results[0], results[1], "exactly one replica should acquire sync leadership, got: A=%v B=%v", results[0], results[1])

	for _, l := range leaderships {
		if l != nil {
			require.NoError(t, l.Close())
		}
	}

	// after release, a third attempt should succeed again
	l3, acquired3, err := replicaA.AcquireSyncLeadership(context.Background())
	require.NoError(t, err)
	require.True(t, acquired3, "sync leadership should be acquirable again after release")
	require.NoError(t, l3.Close())
}
