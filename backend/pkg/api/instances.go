package api

import "github.com/flatcar/nebraska/backend/pkg/api/internal/types"

const (
	InstanceStatusUndefined     = types.InstanceStatusUndefined
	InstanceStatusUpdateGranted = types.InstanceStatusUpdateGranted
	InstanceStatusError         = types.InstanceStatusError
	InstanceStatusComplete      = types.InstanceStatusComplete
	InstanceStatusInstalled     = types.InstanceStatusInstalled
	InstanceStatusDownloaded    = types.InstanceStatusDownloaded
	InstanceStatusDownloading   = types.InstanceStatusDownloading
	InstanceStatusOnHold        = types.InstanceStatusOnHold
)

type (
	Instance                   = types.Instance
	InstancesWithTotal         = types.InstancesWithTotal
	InstanceApplication        = types.InstanceApplication
	InstanceStatusHistoryEntry = types.InstanceStatusHistoryEntry
	InstancesQueryParams       = types.InstancesQueryParams
	InstanceStats              = types.InstanceStats
)
