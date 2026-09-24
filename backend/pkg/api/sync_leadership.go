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
	"database/sql"
)

// syncLeadershipLockID is a fixed, arbitrary Postgres advisory lock id used
// to coordinate the syncer's periodic sync tick across replicas. There is
// only ever one syncer per Nebraska deployment (it is not per-tenant), so a
// single fixed id is sufficient - unlike per-TMS locks, no derivation from
// configuration is needed here.
const syncLeadershipLockID = 5548415426157

// SyncLeadership represents held leadership for one sync tick, backed by a
// session-scoped Postgres advisory lock on a dedicated connection. Close
// releases the lock and the connection. If the process holding it crashes,
// Postgres detects the dead connection and releases the lock automatically.
type SyncLeadership struct {
	conn *sql.Conn
}

// Close releases the advisory lock and the underlying connection.
func (l *SyncLeadership) Close() error {
	defer func() { _ = l.conn.Close() }()

	_, err := l.conn.ExecContext(context.Background(), "SELECT pg_advisory_unlock($1)", syncLeadershipLockID)

	return err
}

// AcquireSyncLeadership attempts to acquire the sync leadership advisory
// lock, so only one replica runs the periodic sync at a time.
//
// Previously every replica ran its own independent syncer with no
// coordination: in HA deployments (e.g. via the Helm chart's
// replicaCount), each replica's syncer ticked and wrote to the same shared
// tables concurrently, causing duplicate-key violations and panics under
// load. See https://github.com/flatcar/nebraska/issues/388.
//
// Returns acquired=false and a nil leadership if another replica currently
// holds the lock; the caller should skip that tick rather than block.
func (api *API) AcquireSyncLeadership(ctx context.Context) (*SyncLeadership, bool, error) {
	conn, err := api.db.Conn(ctx)
	if err != nil {
		return nil, false, err
	}

	var acquired bool
	if err := conn.QueryRowContext(ctx, "SELECT pg_try_advisory_lock($1)", syncLeadershipLockID).Scan(&acquired); err != nil {
		_ = conn.Close()

		return nil, false, err
	}

	if !acquired {
		_ = conn.Close()

		return nil, false, nil
	}

	return &SyncLeadership{conn: conn}, true, nil
}
