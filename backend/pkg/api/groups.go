package api

import "github.com/flatcar/nebraska/backend/pkg/api/internal/types"

var (
	// ErrInvalidChannel error indicates that a channel doesn't belong to the
	// application it was supposed to belong to.
	ErrInvalidChannel = types.ErrInvalidChannel

	// ErrExpectingValidTimezone error indicates that a valid timezone wasn't
	// provided when enabling the flag PolicyOfficeHours.
	ErrExpectingValidTimezone = types.ErrExpectingValidTimezone
)

type (
	GroupDescriptor                 = types.GroupDescriptor
	Group                           = types.Group
	VersionBreakdownEntry           = types.VersionBreakdownEntry
	VersionCountTimelineEntry       = types.VersionCountTimelineEntry
	StatusVersionCountTimelineEntry = types.StatusVersionCountTimelineEntry
	VersionCountMap                 = types.VersionCountMap
	InstancesStatusStats            = types.InstancesStatusStats
	UpdatesStats                    = types.UpdatesStats
)
