package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"regexp"
	"time"
	"unicode/utf8"

	"gopkg.in/guregu/null.v4"
)

// ErrInvalidLocalOverrides indicates a change to a node's overrides of a group's
// policy that the node cannot apply.
var ErrInvalidLocalOverrides = errors.New("nebraska: invalid local overrides")

// maxTimezoneLength is the size of the group_local column that holds a timezone
// override.
const maxTimezoneLength = 40

// OverrideIntervalPattern is the form of an interval override, as in the API
// spec. The bound keeps now() minus the interval within PostgreSQL's range.
const OverrideIntervalPattern = `^[1-9][0-9]{0,5} (minute|hour|day)s?$`

var overrideInterval = regexp.MustCompile(OverrideIntervalPattern)

// GroupPolicyOverridesPatch changes a node's overrides of a group's policy. A
// nil field keeps its override, a null one removes it and a value sets it.
type GroupPolicyOverridesPatch struct {
	PolicyUpdatesEnabled      *null.Bool   `json:"policy_updates_enabled,omitempty"`
	PolicySafeMode            *null.Bool   `json:"policy_safe_mode,omitempty"`
	PolicyOfficeHours         *null.Bool   `json:"policy_office_hours,omitempty"`
	PolicyTimezone            *null.String `json:"policy_timezone,omitempty"`
	PolicyPeriodInterval      *null.String `json:"policy_period_interval,omitempty"`
	PolicyMaxUpdatesPerPeriod *null.Int    `json:"policy_max_updates_per_period,omitempty"`
	PolicyUpdateTimeout       *null.String `json:"policy_update_timeout,omitempty"`
}

// UnmarshalJSON tells a field that is left out from one that is null, which
// encoding/json does not, and refuses unknown fields.
func (p *GroupPolicyOverridesPatch) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidLocalOverrides, err)
	}
	if raw == nil {
		return fmt.Errorf("%w: expected an object", ErrInvalidLocalOverrides)
	}

	var patch GroupPolicyOverridesPatch
	for name, value := range raw {
		var err error
		switch name {
		case "policy_updates_enabled":
			patch.PolicyUpdatesEnabled = new(null.Bool)
			err = json.Unmarshal(value, patch.PolicyUpdatesEnabled)
		case "policy_safe_mode":
			patch.PolicySafeMode = new(null.Bool)
			err = json.Unmarshal(value, patch.PolicySafeMode)
		case "policy_office_hours":
			patch.PolicyOfficeHours = new(null.Bool)
			err = json.Unmarshal(value, patch.PolicyOfficeHours)
		case "policy_timezone":
			patch.PolicyTimezone = new(null.String)
			err = json.Unmarshal(value, patch.PolicyTimezone)
		case "policy_period_interval":
			patch.PolicyPeriodInterval = new(null.String)
			err = json.Unmarshal(value, patch.PolicyPeriodInterval)
		case "policy_max_updates_per_period":
			patch.PolicyMaxUpdatesPerPeriod = new(null.Int)
			err = json.Unmarshal(value, patch.PolicyMaxUpdatesPerPeriod)
		case "policy_update_timeout":
			patch.PolicyUpdateTimeout = new(null.String)
			err = json.Unmarshal(value, patch.PolicyUpdateTimeout)
		default:
			return fmt.Errorf("%w: unknown field %q", ErrInvalidLocalOverrides, name)
		}
		if err != nil {
			return fmt.Errorf("%w: %s: %v", ErrInvalidLocalOverrides, name, err)
		}
	}

	*p = patch
	return nil
}

// Changes returns the overrides the patch changes, by JSON name, with a null
// value for an override it removes.
func (p *GroupPolicyOverridesPatch) Changes() map[string]any {
	changes := map[string]any{}
	if p.PolicyUpdatesEnabled != nil {
		changes["policy_updates_enabled"] = *p.PolicyUpdatesEnabled
	}
	if p.PolicySafeMode != nil {
		changes["policy_safe_mode"] = *p.PolicySafeMode
	}
	if p.PolicyOfficeHours != nil {
		changes["policy_office_hours"] = *p.PolicyOfficeHours
	}
	if p.PolicyTimezone != nil {
		changes["policy_timezone"] = *p.PolicyTimezone
	}
	if p.PolicyPeriodInterval != nil {
		changes["policy_period_interval"] = *p.PolicyPeriodInterval
	}
	if p.PolicyMaxUpdatesPerPeriod != nil {
		changes["policy_max_updates_per_period"] = *p.PolicyMaxUpdatesPerPeriod
	}
	if p.PolicyUpdateTimeout != nil {
		changes["policy_update_timeout"] = *p.PolicyUpdateTimeout
	}
	return changes
}

// Validate checks each override the patch sets on its own.
func (p *GroupPolicyOverridesPatch) Validate() error {
	// The group default is the pause every node honors, so an override can
	// pause a node but cannot enable updates the default disables.
	if v := p.PolicyUpdatesEnabled; v != nil && v.Valid && v.Bool {
		return fmt.Errorf("%w: policy_updates_enabled can only be false, which pauses updates on this node, or null", ErrInvalidLocalOverrides)
	}
	if v := p.PolicyTimezone; v != nil && v.Valid && !validTimezone(v.String) {
		return fmt.Errorf("%w: policy_timezone %q is not a known time zone", ErrInvalidLocalOverrides, v.String)
	}
	if v := p.PolicyMaxUpdatesPerPeriod; v != nil && v.Valid && (v.Int64 < 1 || v.Int64 > math.MaxInt32) {
		return fmt.Errorf("%w: policy_max_updates_per_period must be from 1 to %d", ErrInvalidLocalOverrides, math.MaxInt32)
	}
	for _, f := range []struct {
		name  string
		value *null.String
	}{
		{"policy_period_interval", p.PolicyPeriodInterval},
		{"policy_update_timeout", p.PolicyUpdateTimeout},
	} {
		if v := f.value; v != nil && v.Valid && !overrideInterval.MatchString(v.String) {
			return fmt.Errorf("%w: %s %q must be a whole number of minutes, hours or days from 1 to 999999, such as \"15 minutes\"", ErrInvalidLocalOverrides, f.name, v.String)
		}
	}
	return nil
}

func validTimezone(tz string) bool {
	// LoadLocation reads "" as UTC and "Local" as the server's own zone, neither
	// of which names a zone.
	if tz == "" || tz == "Local" || utf8.RuneCountInString(tz) > maxTimezoneLength {
		return false
	}
	_, err := time.LoadLocation(tz)
	return err == nil
}
