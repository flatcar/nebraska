package runtime

import (
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
	_ = rs.updateInstanceStatus(instanceID, tApp.ID, types.InstanceStatusComplete)

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
	for _, versionMap := range versionTimelineMap {
		totalInstances += versionMap[version]
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

	registerInstanceWithStatusHistory := func(appID, groupID string, status int, age time.Duration, customID ...string) {
		id := uuid.New().String()
		if len(customID) > 0 && customID[0] != "" {
			id = customID[0]
		}
		_, err := rs.RegisterInstance(types.Instance{ID: id, IP: dummyIPAddress}, NewInstanceApplication(appID, groupID, initialVersion))
		require.NoError(t, err)

		instance, err := a.GetInstance(id, appID)
		require.NoError(t, err)

		// FIXME: updateInstanceStatus implicitly requires grantUpdate to populate last_update_version.
		require.NoError(t, rs.grantUpdate(instance, targetVersion))
		require.NoError(t, rs.updateInstanceStatus(id, appID, status))

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
		registerInstanceWithStatusHistory(tApp.ID, tGroup.ID, types.InstanceStatusComplete, 0)
		registerInstanceWithStatusHistory(tApp.ID, tGroup.ID, types.InstanceStatusDownloading, 0)

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
		registerInstanceWithStatusHistory(tApp.ID, tGroup.ID, types.InstanceStatusComplete, 0)
		registerInstanceWithStatusHistory(tApp.ID, tGroup.ID, types.InstanceStatusDownloading, 5*day)
		registerInstanceWithStatusHistory(tApp.ID, tGroup.ID, types.InstanceStatusInstalled, 15*day)
		registerInstanceWithStatusHistory(tApp.ID, tGroup.ID, types.InstanceStatusError, 45*day)
		registerInstanceWithStatusHistory(tApp.ID, tGroup.ID, types.InstanceStatusComplete, 0, "{12345678-1234-1234-1234-123456789012}")

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
