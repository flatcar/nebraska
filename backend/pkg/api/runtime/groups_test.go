package runtime

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/flatcar/nebraska/backend/pkg/api/types"
)

func TestGetVersionCountTimeline(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)
	rs := runtimeSvc(a)

	// Set cache lifespan to 50ms for testing and restore when done
	cacheManager := NewTestCacheManager()
	oldLifespan := cacheManager.SetCacheLifespanForTest(50 * time.Millisecond)
	defer cacheManager.RestoreCacheLifespan(oldLifespan)

	version := "4.0.0"
	tTeam, _ := as.AddTeam(&types.Team{Name: "test_team"})
	tApp, _ := as.AddApp(&types.Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := as.AddPackage(&types.Package{Type: types.PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := as.AddChannel(&types.Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := as.AddGroup(&types.Group{Name: "test_group1", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	instanceID := uuid.New().String()

	_, _ = rs.RegisterInstance(types.Instance{ID: instanceID, IP: "10.0.0.1"}, NewInstanceApplication(tApp.ID, tGroup.ID, version))

	instance, err := a.GetInstance(instanceID, tApp.ID)
	assert.NoError(t, err)
	_ = rs.grantUpdate(instance, version)
	instance, err = a.GetInstance(instanceID, tApp.ID)
	assert.NoError(t, err)
	_ = rs.updateInstanceObjStatus(instance, types.InstanceStatusComplete)

	var versionTimelineMap map[time.Time](types.VersionCountMap)
	var isCache bool

	// get VersionCountTimeline from 1 hr before now
	_, isCache, err = a.GetGroupVersionCountTimeline(tGroup.ID, "1h")
	assert.NoError(t, err)

	// the first time the cache is not hit
	assert.Equal(t, false, isCache)

	// wait a moment for the cache to be set (it's set asynchronously)
	time.Sleep(10 * time.Millisecond)

	// call again - should hit cache (within 50ms lifespan)
	_, isCache, err = a.GetGroupVersionCountTimeline(tGroup.ID, "1h")
	assert.NoError(t, err)

	// the cache must be hit
	assert.Equal(t, true, isCache)

	// wait for cache to expire (100ms > 50ms lifespan)
	time.Sleep(100 * time.Millisecond)

	versionTimelineMap, isCache, err = a.GetGroupVersionCountTimeline(tGroup.ID, "1h")
	assert.NoError(t, err)

	// the cache must be stale as we waited longer than the lifespan
	assert.Equal(t, false, isCache)

	var totalInstances uint64
	for _, versionBreakdown := range versionTimelineMap {
		totalInstances += versionBreakdown[version]
	}
	assert.Equal(t, totalInstances, uint64(1))
	// for 1h we generate timestamp for every 15 minute so total timeline should have 5 timestamps
	assert.Equal(t, len(versionTimelineMap), 5)

	versionTimelineMap, isCache, err = a.GetGroupVersionCountTimeline(tGroup.ID, "1d")
	assert.NoError(t, err)
	// for 1d we generate timestamp for each hour so total timeline should have 25 timestamps
	assert.Equal(t, len(versionTimelineMap), 25)
	// the first time the cache is not hit
	assert.Equal(t, false, isCache)

	versionTimelineMap, isCache, err = a.GetGroupVersionCountTimeline(tGroup.ID, "7d")
	assert.NoError(t, err)
	// for 7d we generate timestamp for each day so total timeline should have 8 timestamps
	assert.Equal(t, len(versionTimelineMap), 8)
	// the first time the cache is not hit
	assert.Equal(t, false, isCache)

	versionTimelineMap, isCache, err = a.GetGroupVersionCountTimeline(tGroup.ID, "30d")
	assert.NoError(t, err)
	// for 30d we generate timestamp after each 3days so total timeline should have 11 timestamps
	assert.Equal(t, len(versionTimelineMap), 11)
	// the first time the cache is not hit
	assert.Equal(t, false, isCache)

	// An unknown group ID returns all 5 time buckets with a nil error instead of sql.ErrNoRows.
	// Because generate_series always emits 5 buckets, check that each bucket is actually empty.
	versionTimelineMap, isCache, err = a.GetGroupVersionCountTimeline(uuid.New().String(), "1h")
	assert.NoError(t, err)
	assert.False(t, isCache)
	assert.Equal(t, 5, len(versionTimelineMap))
	for _, versionBreakdown := range versionTimelineMap {
		assert.Empty(t, versionBreakdown)
	}
}

func TestGetStatusCountTimeline(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)
	rs := runtimeSvc(a)

	const (
		dummyIPAddress = "10.0.0.1"
		initialVersion = "1.0.0"
		targetVersion  = "1.0.1"
	)

	registerInstanceWithStatusHistory := func(t *testing.T, appID, groupID string, status int, age time.Duration, customID ...string) {
		t.Helper()

		id := uuid.New().String()
		if len(customID) > 0 && customID[0] != "" {
			id = customID[0]
		}
		_, err := rs.RegisterInstance(types.Instance{ID: id, IP: dummyIPAddress}, NewInstanceApplication(appID, groupID, initialVersion))
		require.NoError(t, err)

		instance, err := a.GetInstance(id, appID)
		require.NoError(t, err)

		// FIXME: updateInstanceObjStatus implicitly requires grantUpdate to populate last_update_version.
		require.NoError(t, rs.grantUpdate(instance, targetVersion))
		instance, err = a.GetInstance(id, appID)
		require.NoError(t, err)
		require.NoError(t, rs.updateInstanceObjStatus(instance, status))

		if age > 0 {
			backdateInstanceStatusHistory(t, a, id, age)
		}
	}

	t.Run("latest_bucket_and_duration_counts", func(t *testing.T) {
		tTeam, err := as.AddTeam(&types.Team{Name: "test_team"})
		require.NoError(t, err)
		tApp, err := as.AddApp(&types.Application{Name: "test_app", TeamID: tTeam.ID})
		require.NoError(t, err)
		tGroup, err := as.AddGroup(&types.Group{
			Name:                      "test_group1",
			ApplicationID:             tApp.ID,
			PolicyPeriodInterval:      "15 minutes",
			PolicyMaxUpdatesPerPeriod: 2,
			PolicyUpdateTimeout:       "60 minutes",
		})
		require.NoError(t, err)

		// Register two instances with distinct statuses to verify breakdown aggregation
		registerInstanceWithStatusHistory(t, tApp.ID, tGroup.ID, types.InstanceStatusComplete, 0)
		registerInstanceWithStatusHistory(t, tApp.ID, tGroup.ID, types.InstanceStatusDownloading, 0)

		// 1. Verify "1h" duration: 5 buckets (15m interval is hardcoded by durationCodeToPostgresTimings)
		statusTimelineMap, err := a.GetGroupStatusCountTimeline(tGroup.ID, "1h")
		require.NoError(t, err)
		assert.Equal(t, 5, len(statusTimelineMap), "1h window produces 5 15-minute buckets (60m/15m + 1)")

		// Find the latest timestamp bucket
		var latestTimestamp time.Time
		for ts := range statusTimelineMap {
			if ts.After(latestTimestamp) {
				latestTimestamp = ts
			}
		}

		// Verify events reside strictly in the latest bucket, and earlier buckets are zero
		for ts, statusBreakdown := range statusTimelineMap {
			if ts.Equal(latestTimestamp) {
				assert.Equal(t, uint64(1), statusBreakdown[types.InstanceStatusComplete][targetVersion], "expected 1 complete instance in latest bucket")
				assert.Equal(t, uint64(1), statusBreakdown[types.InstanceStatusDownloading][targetVersion], "expected 1 downloading instance in latest bucket")
			} else {
				assert.Equal(t, uint64(0), statusBreakdown[types.InstanceStatusComplete][targetVersion], "earlier buckets must have 0 complete instances")
				assert.Equal(t, uint64(0), statusBreakdown[types.InstanceStatusDownloading][targetVersion], "earlier buckets must have 0 downloading instances")
			}
		}

		// 2. Verify bucket counts for remaining durations (PostgreSQL generate_series includes both boundaries: intervals + 1)
		statusTimelineMap, err = a.GetGroupStatusCountTimeline(tGroup.ID, "1d")
		require.NoError(t, err)
		assert.Equal(t, 25, len(statusTimelineMap), "1d window produces 25 hourly buckets (24h/1h + 1)")

		statusTimelineMap, err = a.GetGroupStatusCountTimeline(tGroup.ID, "7d")
		require.NoError(t, err)
		assert.Equal(t, 8, len(statusTimelineMap), "7d window produces 8 daily buckets (7d/1d + 1)")

		statusTimelineMap, err = a.GetGroupStatusCountTimeline(tGroup.ID, "30d")
		require.NoError(t, err)
		assert.Equal(t, 11, len(statusTimelineMap), "30d window produces 11 3-day buckets (30d/3d + 1)")

		// An unknown group ID returns all 5 time buckets with a nil error instead of sql.ErrNoRows.
		// Because generate_series always emits 5 buckets, check that each bucket is actually empty.
		statusTimelineMap, err = a.GetGroupStatusCountTimeline(uuid.New().String(), "1h")
		require.NoError(t, err)
		assert.Equal(t, 5, len(statusTimelineMap))
		for _, statusBreakdown := range statusTimelineMap {
			assert.Empty(t, statusBreakdown)
		}
	})

	t.Run("historical_multi_bucket_and_filtering", func(t *testing.T) {
		tTeam, err := as.AddTeam(&types.Team{Name: "test_team_historical"})
		require.NoError(t, err)
		tApp, err := as.AddApp(&types.Application{Name: "test_app_historical", TeamID: tTeam.ID})
		require.NoError(t, err)
		tGroup, err := as.AddGroup(&types.Group{
			Name:                      "test_group_historical",
			ApplicationID:             tApp.ID,
			PolicyPeriodInterval:      "15 minutes",
			PolicyMaxUpdatesPerPeriod: 2,
			PolicyUpdateTimeout:       "60 minutes",
		})
		require.NoError(t, err)

		// Window filtering depends strictly on Age.
		//
		// | Inst     | Tracer      | Age | 1d | 7d | 30d | Notes                  |
		// |----------|-------------|:---:|:--:|:--:|:---:|------------------------|
		// | instNow  | Complete    |  0  |  1 |  1 |  1  | Latest bucket only     |
		// | inst5d   | Downloading |  5d |  0 |  1 |  1  | 5d > 1d window         |
		// | inst15d  | Installed   | 15d |  0 |  0 |  1  | 15d > 7d window        |
		// | inst45d  | Error       | 45d |  0 |  0 |  0  | 45d > 30d window       |
		// | fakeInst | Complete    |  0  |  0 |  0 |  0  | Filtered by ID '{%}'   |
		registerInstanceWithStatusHistory(t, tApp.ID, tGroup.ID, types.InstanceStatusComplete, 0)
		registerInstanceWithStatusHistory(t, tApp.ID, tGroup.ID, types.InstanceStatusDownloading, 5*day)
		registerInstanceWithStatusHistory(t, tApp.ID, tGroup.ID, types.InstanceStatusInstalled, 15*day)
		registerInstanceWithStatusHistory(t, tApp.ID, tGroup.ID, types.InstanceStatusError, 45*day)
		registerInstanceWithStatusHistory(t, tApp.ID, tGroup.ID, types.InstanceStatusComplete, 0, "{12345678-1234-1234-1234-123456789012}")

		countStatusInWindow := func(timeline map[time.Time]map[int]types.VersionCountMap, status int) uint64 {
			var total uint64
			for _, statusMap := range timeline {
				total += statusMap[status][targetVersion]
			}
			return total
		}

		// 2. Test "1d" duration: only 0d Complete present; 5d, 15d, 45d absent
		timeline1d, err := a.GetGroupStatusCountTimeline(tGroup.ID, "1d")
		require.NoError(t, err)
		require.Equal(t, 25, len(timeline1d), "1d window produces 25 hourly buckets (24h/1h + 1)")
		assert.Equal(t, uint64(1), countStatusInWindow(timeline1d, types.InstanceStatusComplete), "1d window should only count latest Complete event")
		assert.Equal(t, uint64(0), countStatusInWindow(timeline1d, types.InstanceStatusDownloading), "5d Downloading event should not be in 1d window")
		assert.Equal(t, uint64(0), countStatusInWindow(timeline1d, types.InstanceStatusInstalled), "15d Installed event should not be in 1d window")
		assert.Equal(t, uint64(0), countStatusInWindow(timeline1d, types.InstanceStatusError), "45d Error event should not be in 1d window")

		// 3. Test "7d" duration: 0d Complete and 5d Downloading present; 15d Installed and 45d Error absent
		timeline7d, err := a.GetGroupStatusCountTimeline(tGroup.ID, "7d")
		require.NoError(t, err)
		require.Equal(t, 8, len(timeline7d), "7d window produces 8 daily buckets (7d/1d + 1)")
		assert.Equal(t, uint64(1), countStatusInWindow(timeline7d, types.InstanceStatusComplete), "7d window should count latest Complete event")
		assert.Equal(t, uint64(1), countStatusInWindow(timeline7d, types.InstanceStatusDownloading), "7d window should count 5d Downloading event")
		assert.Equal(t, uint64(0), countStatusInWindow(timeline7d, types.InstanceStatusInstalled), "15d Installed event should not be in 7d window")
		assert.Equal(t, uint64(0), countStatusInWindow(timeline7d, types.InstanceStatusError), "45d Error event should not be in 7d window")

		// 4. Test "30d" duration: 0d Complete, 5d Downloading, 15d Installed present; 45d Error (45d ago) strictly excluded
		timeline30d, err := a.GetGroupStatusCountTimeline(tGroup.ID, "30d")
		require.NoError(t, err)
		require.Equal(t, 11, len(timeline30d), "30d window produces 11 3-day buckets (30d/3d + 1)")
		assert.Equal(t, uint64(1), countStatusInWindow(timeline30d, types.InstanceStatusComplete), "30d window should count latest Complete event")
		assert.Equal(t, uint64(1), countStatusInWindow(timeline30d, types.InstanceStatusDownloading), "30d window should count 5d Downloading event")
		assert.Equal(t, uint64(1), countStatusInWindow(timeline30d, types.InstanceStatusInstalled), "30d window should count 15d Installed event")
		assert.Equal(t, uint64(0), countStatusInWindow(timeline30d, types.InstanceStatusError), "45d Error event is past 30d and must be excluded")
	})
}

func TestUpdateLocalOverrides(t *testing.T) {
	a := newForTest(t)
	defer a.Close()
	as := adminSvc(a)
	rs := runtimeSvc(a)

	tTeam, _ := as.AddTeam(&types.Team{Name: "test_team"})
	tApp, _ := as.AddApp(&types.Application{Name: "test_app", TeamID: tTeam.ID})
	tGroup, err := as.AddGroup(&types.Group{Name: "test_group", ApplicationID: tApp.ID, PolicyUpdatesEnabled: true, PolicyTimezone: null.StringFrom("UTC"), PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	require.NoError(t, err)

	update := func(t *testing.T, body string) error {
		t.Helper()

		var patch types.GroupPolicyOverridesPatch
		require.NoError(t, json.Unmarshal([]byte(body), &patch))
		return rs.UpdateLocalOverrides(tGroup.ID, &patch)
	}
	read := func(t *testing.T) *types.Group {
		t.Helper()

		group, err := rs.GetGroup(tGroup.ID)
		require.NoError(t, err)
		return group
	}
	reset := func(t *testing.T) {
		t.Helper()

		require.NoError(t, update(t, `{"policy_updates_enabled": null, "policy_safe_mode": null, "policy_office_hours": null,
			"policy_timezone": null, "policy_period_interval": null, "policy_max_updates_per_period": null, "policy_update_timeout": null}`))
	}
	adminDefault := types.GroupPolicy{
		PolicyUpdatesEnabled:      true,
		PolicyTimezone:            null.StringFrom("UTC"),
		PolicyPeriodInterval:      "15 minutes",
		PolicyMaxUpdatesPerPeriod: 2,
		PolicyUpdateTimeout:       "60 minutes",
	}

	t.Run("sets_overrides_without_touching_the_default", func(t *testing.T) {
		t.Cleanup(func() { reset(t) })

		require.NoError(t, update(t, `{"policy_safe_mode": true, "policy_max_updates_per_period": 5, "policy_period_interval": "2 hours"}`))

		group := read(t)
		// The rollout engine reads the top-level fields.
		assert.True(t, group.PolicySafeMode)
		assert.Equal(t, 5, group.PolicyMaxUpdatesPerPeriod)
		assert.Equal(t, "2 hours", group.PolicyPeriodInterval)
		assert.Equal(t, "60 minutes", group.PolicyUpdateTimeout)
		assert.Equal(t, null.IntFrom(5), group.LocalOverrides.PolicyMaxUpdatesPerPeriod)
		assert.False(t, group.LocalOverrides.PolicyUpdateTimeout.Valid)
		assert.Equal(t, adminDefault, group.AdminPolicy)
	})

	t.Run("leaves_out_what_it_does_not_name", func(t *testing.T) {
		t.Cleanup(func() { reset(t) })

		require.NoError(t, update(t, `{"policy_safe_mode": true}`))
		// The rollout engine trips the brake while the operator edits something else.
		require.NoError(t, rs.disableUpdates(tGroup.ID))
		require.NoError(t, update(t, `{"policy_safe_mode": null, "policy_update_timeout": "2 hours"}`))

		group := read(t)
		assert.False(t, group.PolicyUpdatesEnabled, "the brake must stay")
		assert.False(t, group.LocalOverrides.PolicySafeMode.Valid)
		assert.Equal(t, null.StringFrom("2 hours"), group.LocalOverrides.PolicyUpdateTimeout)
	})

	t.Run("pauses_and_resumes_this_node", func(t *testing.T) {
		t.Cleanup(func() { reset(t) })

		require.NoError(t, update(t, `{"policy_updates_enabled": false}`))
		assert.False(t, read(t).PolicyUpdatesEnabled)

		require.NoError(t, update(t, `{"policy_updates_enabled": null}`))
		assert.True(t, read(t).PolicyUpdatesEnabled, "back to the default")
	})

	t.Run("an_office_hours_override_carries_its_timezone", func(t *testing.T) {
		t.Cleanup(func() { reset(t) })

		assert.ErrorIs(t, update(t, `{"policy_office_hours": true}`), types.ErrInvalidLocalOverrides)

		require.NoError(t, update(t, `{"policy_office_hours": true, "policy_timezone": "Asia/Tokyo"}`))
		group := read(t)
		assert.True(t, group.PolicyOfficeHours)
		assert.Equal(t, null.StringFrom("Asia/Tokyo"), group.PolicyTimezone)

		assert.ErrorIs(t, update(t, `{"policy_timezone": null}`), types.ErrInvalidLocalOverrides)
		require.NoError(t, update(t, `{"policy_office_hours": false}`))
		require.NoError(t, update(t, `{"policy_timezone": null}`))
	})

	t.Run("a_timezone_override_goes_with_office_hours", func(t *testing.T) {
		t.Cleanup(func() { reset(t) })

		// Left alone it would pair with office hours the default may turn on.
		assert.ErrorIs(t, update(t, `{"policy_timezone": "Asia/Tokyo"}`), types.ErrInvalidLocalOverrides)

		require.NoError(t, update(t, `{"policy_office_hours": true, "policy_timezone": "Asia/Tokyo"}`))
		assert.ErrorIs(t, update(t, `{"policy_office_hours": null}`), types.ErrInvalidLocalOverrides)
		require.NoError(t, update(t, `{"policy_office_hours": null, "policy_timezone": null}`))
		assert.Equal(t, types.GroupPolicyOverrides{}, *read(t).LocalOverrides)
	})

	t.Run("refuses_what_it_cannot_apply", func(t *testing.T) {
		assert.ErrorIs(t, update(t, `{"policy_safe_mode": true, "policy_updates_enabled": true}`), types.ErrInvalidLocalOverrides)

		group := read(t)
		assert.Equal(t, types.GroupPolicyOverrides{}, *group.LocalOverrides, "a refused change writes nothing")
	})

	// The rollout engine computes now() less the period on every update check
	// while an instance of the group is active.
	t.Run("the_longest_interval_still_works_in_an_update_check", func(t *testing.T) {
		t.Cleanup(func() { reset(t) })

		_, err := rs.RegisterInstance(types.Instance{ID: uuid.New().String(), IP: "10.0.0.1"}, NewInstanceApplication(tApp.ID, tGroup.ID, "1.0.0"))
		require.NoError(t, err)

		require.NoError(t, update(t, `{"policy_period_interval": "999999 days", "policy_update_timeout": "999999 days"}`))
		group := read(t)
		_, err = rs.GetGroupRolloutPolicyStats(group, group.PolicyMaxUpdatesPerPeriod, group.PolicySafeMode)
		assert.NoError(t, err)

		// One the bound refuses would fail every update check in the group.
		_, err = rs.db.Exec(`update group_local set policy_period_interval_override = '3000000 days' where group_id = $1`, tGroup.ID)
		require.NoError(t, err)
		group = read(t)
		_, err = rs.GetGroupRolloutPolicyStats(group, group.PolicyMaxUpdatesPerPeriod, group.PolicySafeMode)
		assert.ErrorContains(t, err, "out of range")
	})

	// A row may hold a timezone override without office hours, such as one
	// changed by hand. A change that leaves both alone still applies.
	t.Run("a_pause_never_fails_on_account_of_office_hours", func(t *testing.T) {
		t.Cleanup(func() { reset(t) })

		_, err := rs.db.Exec(`update group_local set policy_timezone_override = 'Asia/Tokyo' where group_id = $1`, tGroup.ID)
		require.NoError(t, err)

		for _, body := range []string{
			`{"policy_updates_enabled": false}`,
			`{"policy_updates_enabled": null}`,
			`{"policy_safe_mode": true}`,
			`{}`,
		} {
			assert.NoError(t, update(t, body), body)
		}

		// A change to either still has to leave them paired.
		assert.ErrorIs(t, update(t, `{"policy_timezone": "Europe/Paris"}`), types.ErrInvalidLocalOverrides)
		require.NoError(t, update(t, `{"policy_office_hours": null, "policy_timezone": null}`))
		assert.Equal(t, types.GroupPolicyOverrides{PolicySafeMode: null.BoolFrom(true)}, *read(t).LocalOverrides)
	})

	t.Run("unknown_group", func(t *testing.T) {
		var patch types.GroupPolicyOverridesPatch
		require.NoError(t, json.Unmarshal([]byte(`{}`), &patch))

		assert.ErrorIs(t, rs.UpdateLocalOverrides(uuid.New().String(), &patch), types.ErrNoRowsAffected)
	})
}
