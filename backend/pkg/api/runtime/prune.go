package runtime

import (
	"fmt"
	"time"
)

const defaultPruneBatchSize = 500

// PruneStaleInstancesResult is the outcome of one prune pass.
type PruneStaleInstancesResult struct {
	// Candidates is the number of instances matching the cutoff. In a live
	// run this equals Deleted; in a dry-run it is the would-delete count.
	Candidates int64
	// Deleted is the number of instance rows actually removed. Zero in dry-run.
	Deleted int64
}

// PruneStaleInstances deletes (or counts, when dryRun is true) instances whose
// newest last_check_for_updates is older than cutoff. Instances that never
// registered an application fall back to instance.created_ts.
//
// Related rows in instance_application, instance_status_history, event and
// activity are removed by ON DELETE CASCADE. Work happens in batches so a
// large fleet does not hold a long lock on instance.
func (s *Service) PruneStaleInstances(cutoff time.Time, batchSize int, dryRun bool) (PruneStaleInstancesResult, error) {
	if batchSize <= 0 {
		batchSize = defaultPruneBatchSize
	}

	if dryRun {
		var count int64
		err := s.db.QueryRowx(`
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
	for {
		result, err := s.db.Exec(`
			DELETE FROM instance
			WHERE id IN (
				SELECT i.id
				FROM instance i
				LEFT JOIN instance_application ia ON ia.instance_id = i.id
				GROUP BY i.id, i.created_ts
				HAVING COALESCE(MAX(ia.last_check_for_updates), i.created_ts) < $1
				LIMIT $2
			)
		`, cutoff, batchSize)
		if err != nil {
			return PruneStaleInstancesResult{}, fmt.Errorf("delete stale instances: %w", err)
		}
		n, err := result.RowsAffected()
		if err != nil {
			return PruneStaleInstancesResult{}, fmt.Errorf("stale instance rows affected: %w", err)
		}
		deleted += n
		if n < int64(batchSize) {
			break
		}
	}

	return PruneStaleInstancesResult{Candidates: deleted, Deleted: deleted}, nil
}
