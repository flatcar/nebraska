package runtime

import (
	"context"
	"fmt"
	"time"
)

const (
	defaultPruneBatchSize  = 500
	defaultPruneMaxBatches = 20
	defaultPruneTimeout    = 2 * time.Minute
)

// PruneStaleInstancesResult is the outcome of one prune pass.
type PruneStaleInstancesResult struct {
	// Candidates is the number of instances matching the cutoff. In a live
	// run this equals Deleted; in a dry-run it is the would-delete count.
	Candidates int64
	// Deleted is the number of instance rows actually removed. Zero in dry-run.
	Deleted int64
	// MoreRemain is true when this pass hit the per-run batch cap and stale
	// rows are likely still present for a later tick.
	MoreRemain bool
}

// PruneStaleInstances deletes (or counts, when dryRun is true) instances whose
// newest last_check_for_updates is older than cutoff. Instances that never
// registered an application fall back to instance.created_ts.
//
// Related rows in instance_application, instance_status_history, event and
// activity are removed by ON DELETE CASCADE. Each DELETE is limited to
// batchSize rows, and a single call processes at most a handful of batches so
// a first enable against years of dead nodes cannot lock the hot tables for
// the whole hour.
func (s *Service) PruneStaleInstances(cutoff time.Time, batchSize int, dryRun bool) (PruneStaleInstancesResult, error) {
	return s.pruneStaleInstances(cutoff, batchSize, defaultPruneMaxBatches, dryRun)
}

func (s *Service) pruneStaleInstances(cutoff time.Time, batchSize, maxBatches int, dryRun bool) (PruneStaleInstancesResult, error) {
	if batchSize <= 0 {
		batchSize = defaultPruneBatchSize
	}
	if maxBatches <= 0 {
		maxBatches = defaultPruneMaxBatches
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultPruneTimeout)
	defer cancel()

	if dryRun {
		var count int64
		err := s.db.QueryRowxContext(ctx, `
			SELECT COUNT(*) FROM (
				SELECT i.id
				FROM instance i
				LEFT JOIN instance_application ia ON ia.instance_id = i.id
				GROUP BY i.id, i.created_ts
				HAVING COALESCE(MAX(ia.last_check_for_updates), i.created_ts) < $1
			) stale
		`, cutoff).Scan(&count)
		if err != nil {
			return PruneStaleInstancesResult{}, fmt.Errorf("count stale instances: %w", err)
		}
		return PruneStaleInstancesResult{Candidates: count}, nil
	}

	var deleted int64
	var moreRemain bool
	for batch := 0; batch < maxBatches; batch++ {
		result, err := s.db.ExecContext(ctx, `
			WITH stale AS (
				SELECT i.id
				FROM instance i
				LEFT JOIN instance_application ia ON ia.instance_id = i.id
				GROUP BY i.id, i.created_ts
				HAVING COALESCE(MAX(ia.last_check_for_updates), i.created_ts) < $1
				ORDER BY COALESCE(MAX(ia.last_check_for_updates), i.created_ts) ASC, i.id ASC
				LIMIT $2
			)
			DELETE FROM instance i
			USING stale
			WHERE i.id = stale.id
			  AND COALESCE(
					(SELECT MAX(ia2.last_check_for_updates) FROM instance_application ia2 WHERE ia2.instance_id = i.id),
					i.created_ts
			  ) < $1
		`, cutoff, batchSize)
		if err != nil {
			return PruneStaleInstancesResult{Candidates: deleted, Deleted: deleted}, fmt.Errorf("delete stale instances: %w", err)
		}
		n, err := result.RowsAffected()
		if err != nil {
			return PruneStaleInstancesResult{Candidates: deleted, Deleted: deleted}, fmt.Errorf("stale instance rows affected: %w", err)
		}
		deleted += n
		if n < int64(batchSize) {
			moreRemain = false
			break
		}
		if batch == maxBatches-1 {
			moreRemain = true
		}
	}

	return PruneStaleInstancesResult{Candidates: deleted, Deleted: deleted, MoreRemain: moreRemain}, nil
}
