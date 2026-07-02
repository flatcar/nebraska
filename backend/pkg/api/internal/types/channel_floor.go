package types

import "gopkg.in/guregu/null.v4"

// ChannelFloorInfo contains a channel and its floor reason for a specific package
type ChannelFloorInfo struct {
	Channel     *Channel    `json:"channel"`
	FloorReason null.String `json:"floor_reason"`
}
