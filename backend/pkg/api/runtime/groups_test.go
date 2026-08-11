package runtime

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"

	"github.com/flatcar/nebraska/backend/pkg/api/internal/types"
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

	version := "4.0.0"
	tTeam, _ := as.AddTeam(&types.Team{Name: "test_team"})
	tApp, _ := as.AddApp(&types.Application{Name: "test_app", TeamID: tTeam.ID})
	tPkg, _ := as.AddPackage(&types.Package{Type: types.PkgTypeOther, URL: "http://sample.url/pkg", Version: "12.1.0", ApplicationID: tApp.ID})
	tChannel, _ := as.AddChannel(&types.Channel{Name: "test_channel", Color: "blue", ApplicationID: tApp.ID, PackageID: null.StringFrom(tPkg.ID)})
	tGroup, _ := as.AddGroup(&types.Group{Name: "test_group1", ApplicationID: tApp.ID, ChannelID: null.StringFrom(tChannel.ID), PolicyUpdatesEnabled: true, PolicySafeMode: true, PolicyPeriodInterval: "15 minutes", PolicyMaxUpdatesPerPeriod: 2, PolicyUpdateTimeout: "60 minutes"})
	instanceID1 := uuid.New().String()
	instanceID2 := uuid.New().String()

	_, _ = rs.RegisterInstance(types.Instance{ID: instanceID1, IP: "10.0.0.1"}, NewInstanceApplication(tApp.ID, tGroup.ID, version))

	instance1, err := a.GetInstance(instanceID1, tApp.ID)
	assert.NoError(t, err)

	_ = rs.grantUpdate(instance1, version)
	_ = rs.updateInstanceStatus(instanceID1, tApp.ID, types.InstanceStatusComplete)
	_, _ = rs.RegisterInstance(types.Instance{ID: instanceID2, IP: "10.0.0.2"}, NewInstanceApplication(tApp.ID, tGroup.ID, version))

	instance2, err := a.GetInstance(instanceID2, tApp.ID)
	assert.NoError(t, err)

	_ = rs.grantUpdate(instance2, version)
	_ = rs.updateInstanceStatus(instanceID2, tApp.ID, types.InstanceStatusDownloading)

	// get StatusCountTimeline from 1 hr before now
	statusTimelineMap, err := a.GetGroupStatusCountTimeline(tGroup.ID, "1h")
	assert.NoError(t, err)
	// for 1h we generate timestamp for every 15 minute so total timeline should have 5 timestamps
	assert.Equal(t, len(statusTimelineMap), 5)
	var statusInstanceCountMap = make(map[int]uint64)
	for _, statusMap := range statusTimelineMap {
		statusInstanceCountMap[types.InstanceStatusComplete] += statusMap[types.InstanceStatusComplete][version]
		statusInstanceCountMap[types.InstanceStatusDownloading] += statusMap[types.InstanceStatusDownloading][version]
	}
	// as we registered two instances with version 4.0.0 with statuses 4 and 7
	// so our status breakdown should have count 1 for both status 4 and 7
	assert.Equal(t, statusInstanceCountMap[types.InstanceStatusComplete], uint64(1))
	assert.Equal(t, statusInstanceCountMap[types.InstanceStatusDownloading], uint64(1))

	statusTimelineMap, err = a.GetGroupStatusCountTimeline(tGroup.ID, "1d")
	assert.NoError(t, err)
	// for 1d we generate timestamp for each hour so total timeline should have 25 timestamps
	assert.Equal(t, len(statusTimelineMap), 25)

	statusTimelineMap, err = a.GetGroupStatusCountTimeline(tGroup.ID, "7d")
	assert.NoError(t, err)
	// for 7d we generate timestamp for each day so total timeline should have 8 timestamps
	assert.Equal(t, len(statusTimelineMap), 8)

	statusTimelineMap, err = a.GetGroupStatusCountTimeline(tGroup.ID, "30d")
	assert.NoError(t, err)
	// for 30d we generate timestamp after each 3days so total timeline should have 11 timestamps
	assert.Equal(t, len(statusTimelineMap), 11)
}
