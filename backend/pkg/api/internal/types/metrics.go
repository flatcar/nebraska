package types

type AppInstancesPerChannelMetric struct {
	ApplicationName string `db:"app_name" json:"app_name"`
	Version         string `db:"version" json:"version"`
	ChannelName     string `db:"channel_name" json:"channel_name"`
	InstancesCount  int    `db:"instances_count" json:"instances_count"`
}

type FailedUpdatesMetric struct {
	ApplicationName string `db:"app_name" json:"app_name"`
	FailureCount    int    `db:"fail_count" json:"fail_count"`
}

// GroupUpdatesStatsMetric represents rollout progress and update tracking metrics for a group.
// This surfaces data Nebraska already tracks internally for rollout policy enforcement,
// making it observable via Prometheus for monitoring and alerting.
type GroupUpdatesStatsMetric struct {
	ApplicationID  string `db:"application_id" json:"application_id"`
	GroupID        string `db:"group_id" json:"group_id"`
	ApplicationName string `db:"application_name" json:"application_name"`
	GroupName      string `db:"group_name" json:"group_name"`
	ChannelName    string `db:"channel_name" json:"channel_name"`
	// Update progress stats
	TotalInstances                   int `db:"total_instances" json:"total_instances"`
	UpdatesToCurrentVersionGranted   int `db:"updates_to_current_version_granted" json:"updates_to_current_version_granted"`
	UpdatesToCurrentVersionAttempted int `db:"updates_to_current_version_attempted" json:"updates_to_current_version_attempted"`
	UpdatesToCurrentVersionSucceeded int `db:"updates_to_current_version_succeeded" json:"updates_to_current_version_succeeded"`
	UpdatesToCurrentVersionFailed    int `db:"updates_to_current_version_failed" json:"updates_to_current_version_failed"`
	UpdatesGrantedInLastPeriod       int `db:"updates_granted_in_last_period" json:"updates_granted_in_last_period"`
	UpdatesInProgress                int `db:"updates_in_progress" json:"updates_in_progress"`
	UpdatesTimedOut                  int `db:"updates_timed_out" json:"updates_timed_out"`
}

// InstancesByOEMMetric represents the distribution of instances by OEM/platform.
// This metric is useful for understanding fleet composition across different
// hardware vendors and cloud platforms (e.g., aws, azure, gcp, equinixmetal, etc.)
type InstancesByOEMMetric struct {
	OEM            string `db:"oem" json:"oem"`
	InstancesCount int    `db:"instances_count" json:"instances_count"`
}
