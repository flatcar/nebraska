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

// InstancesPerOEMMetric holds the count of active instances grouped by
// application and OEM platform (e.g. "aws", "azure", "vmware").
type InstancesPerOEMMetric struct {
	ApplicationName string `db:"app_name"`
	OEM             string `db:"oem"`
	InstancesCount  int    `db:"instances_count"`
}
