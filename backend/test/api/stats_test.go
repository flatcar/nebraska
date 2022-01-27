package api_test

import (
	"fmt"
	"net/http"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kinvolk/nebraska/backend/pkg/api"
)

func TestGroupVersionTimeline(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// set timezone
		loc, err := time.LoadLocation("Etc/UTC")
		require.NoError(t, err)
		time.Local = loc

		// establish DB connection
		db := newDBForTest(t)

		// get app that has instance
		appWithInstance := getAppWithInstance(t, db)

		// fetch group version timeline from API
		url := fmt.Sprintf("%s/api/apps/%s/groups/%s/version_timeline?duration=1d", os.Getenv("NEBRASKA_TEST_SERVER_URL"), appWithInstance.ID, appWithInstance.Groups[0].ID)
		method := "GET"

		// response
		var timelineResponse map[time.Time](api.VersionCountMap)

		httpDo(t, url, method, nil, http.StatusOK, "json", &timelineResponse)

		// get group version timeline from DB
		timelineDB, _, err := db.GetGroupVersionCountTimeline(appWithInstance.Groups[0].ID, "1d")
		require.NoError(t, err)
		require.NotNil(t, timelineDB)

		assert.Equal(t, len(timelineDB), len(timelineResponse))

		// Since the response from DB and API doesn't match as they are relative to the time at which
		// they were calculated the response result is converted to arrays and they are matched.

		dbKeys := []time.Time{}
		for k := range timelineDB {
			dbKeys = append(dbKeys, k)
		}

		sort.Slice(dbKeys, func(i, j int) bool {
			return dbKeys[i].Before(dbKeys[j])
		})

		var dbVersionCountMap []api.VersionCountMap
		for _, k := range dbKeys {
			dbVersionCountMap = append(dbVersionCountMap, timelineDB[k])
		}

		respKeys := []time.Time{}

		for k := range timelineResponse {
			respKeys = append(respKeys, k)
		}

		sort.Slice(respKeys, func(i, j int) bool {
			return respKeys[i].Before(respKeys[j])
		})

		var respVersionCountMap []api.VersionCountMap
		for _, k := range respKeys {
			respVersionCountMap = append(respVersionCountMap, timelineResponse[k])
		}

		assert.Equal(t, dbVersionCountMap, respVersionCountMap)
	})
}

func TestGroupVersionBreakdown(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)

		// get app that has instance
		appWithInstance := getAppWithInstance(t, db)

		// fetch version breakdown from DB
		breakdownDB, err := db.GetGroupVersionBreakdown(appWithInstance.Groups[0].ID)
		require.NoError(t, err)
		require.NotNil(t, breakdownDB)

		// fetch version breakdown from API
		url := fmt.Sprintf("%s/api/apps/%s/groups/%s/version_breakdown", os.Getenv("NEBRASKA_TEST_SERVER_URL"), appWithInstance.ID, appWithInstance.Groups[0].ID)
		method := "GET"

		// response
		var breakdownResp []*api.VersionBreakdownEntry

		httpDo(t, url, method, nil, http.StatusOK, "json", &breakdownResp)

		assert.Equal(t, breakdownDB, breakdownResp)
	})
}

func TestGroupStatusTimeline(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// set timezone
		loc, err := time.LoadLocation("Etc/UTC")
		require.NoError(t, err)
		time.Local = loc

		// establish DB connection
		db := newDBForTest(t)

		// get random app
		app := getRandomApp(t, db)

		// create instance for app[0]
		instanceID := uuid.New()
		instanceDB, err := db.RegisterInstance(instanceID.String(), "alias", "0.0.0.0", "0.0.1", app.ID, app.Groups[0].ID)
		require.NoError(t, err)

		// GetUpdatePackage
		_, err = db.GetUpdatePackage(instanceDB.ID, instanceDB.Alias, instanceDB.IP, instanceDB.Application.Version, app.ID, app.Groups[0].ID)
		require.NoError(t, err)

		// create event for instance
		err = db.RegisterEvent(instanceDB.ID, app.ID, app.Groups[0].ID, api.EventUpdateComplete, api.ResultSuccessReboot, "0.0.0", "0")
		require.NoError(t, err)

		// get group status timeline from DB

		groupStatusCountTimelineDB, err := db.GetGroupStatusCountTimeline(app.Groups[0].ID, "1d")
		require.NoError(t, err)
		require.NotNil(t, groupStatusCountTimelineDB)

		// fetch group status timeleine from API
		url := fmt.Sprintf("%s/api/apps/%s/groups/%s/status_timeline?duration=1d", os.Getenv("NEBRASKA_TEST_SERVER_URL"), app.ID, app.Groups[0].ID)
		method := "GET"

		// response
		var groupStatusCountTimelineResp map[time.Time](map[int](api.VersionCountMap))

		httpDo(t, url, method, nil, http.StatusOK, "json", &groupStatusCountTimelineResp)

		assert.Equal(t, len(groupStatusCountTimelineDB), len(groupStatusCountTimelineResp))

		// Since the response from DB and API doesn't match as they are relative to the time at which
		// they were calculated the response result is converted to arrays and they are matched.

		dbKeys := []time.Time{}
		for k := range groupStatusCountTimelineDB {
			dbKeys = append(dbKeys, k)
		}

		sort.Slice(dbKeys, func(i, j int) bool {
			return dbKeys[i].Before(dbKeys[j])
		})

		var dbStatusCountMap []map[int](api.VersionCountMap)
		for _, k := range dbKeys {
			dbStatusCountMap = append(dbStatusCountMap, groupStatusCountTimelineDB[k])
		}

		respKeys := []time.Time{}

		for k := range groupStatusCountTimelineResp {
			respKeys = append(respKeys, k)
		}

		sort.Slice(respKeys, func(i, j int) bool {
			return respKeys[i].Before(respKeys[j])
		})

		var respStatusCountMap []map[int](api.VersionCountMap)
		for _, k := range respKeys {
			respStatusCountMap = append(respStatusCountMap, groupStatusCountTimelineResp[k])
		}

		assert.Equal(t, dbStatusCountMap, respStatusCountMap)
	})
}

func TestGroupInstanceStats(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// establish DB connection
		db := newDBForTest(t)

		// get app that has instance
		appWithInstance := getAppWithInstance(t, db)

		// get instance stats from DB
		instanceStatsDB, err := db.GetGroupInstancesStats(appWithInstance.Groups[0].ID, "1d")
		require.NoError(t, err)
		require.NotNil(t, instanceStatsDB)

		// fetch instances stats from API
		url := fmt.Sprintf("%s/api/apps/%s/groups/%s/instances_stats?duration=1d", os.Getenv("NEBRASKA_TEST_SERVER_URL"), appWithInstance.ID, appWithInstance.Groups[0].ID)
		method := "GET"

		// response
		var instanceStatsResp api.InstancesStatusStats

		httpDo(t, url, method, nil, http.StatusOK, "json", &instanceStatsResp)

		assert.Equal(t, *instanceStatsDB, instanceStatsResp)
	})
}
