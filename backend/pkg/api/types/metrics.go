package types

type AppInstancesPerChannelMetric struct {
	ApplicationName string `db:"app_name" json:"app_name"`
	Version         string `db:"version" json:"version"`
	ChannelName     string `db:"channel_name" json:"channel_name"`
	// Arch is the numeric value of the channel's Arch (see arch.go). It is
	// -1 when the instance's group has no channel (or no group) assigned,
	// in which case ChannelName is also "".
	Arch           int `db:"arch" json:"arch"`
	InstancesCount int `db:"instances_count" json:"instances_count"`
}

type FailedUpdatesMetric struct {
	ApplicationName string `db:"app_name" json:"app_name"`
	FailureCount    int    `db:"fail_count" json:"fail_count"`
}
