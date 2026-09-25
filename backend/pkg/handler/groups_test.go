package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flatcar/nebraska/backend/pkg/api"
	"github.com/flatcar/nebraska/backend/pkg/api/admin"
	"github.com/flatcar/nebraska/backend/pkg/api/runtime"
	"github.com/flatcar/nebraska/backend/pkg/api/types"
	"github.com/flatcar/nebraska/backend/pkg/codegen"
)

const (
	testAppID   = "b6458005-8f40-4627-b33b-be70a718c48e"
	testGroupID = "bcaa68bc-5f82-11e5-9d70-feff819cdc9f"
)

func setupHandler(t *testing.T) (*Handler, *api.API) {
	t.Helper()
	a, err := api.NewForTest(api.OptionInitDB)
	require.NoError(t, err)

	adminSvc := admin.NewService(a.Conn(), a.Reads())
	runtimeSvc := runtime.NewService(a.Conn(), a.Reads(), runtime.Config{DisableUpdatesOnFailedRollout: true})

	h := &Handler{
		db:      a,
		admin:   adminSvc,
		runtime: runtimeSvc,
	}
	return h, a
}

func TestPaginateGroups(t *testing.T) {
	h, a := setupHandler(t)
	defer a.Close()

	e := echo.New()

	// 1. Success
	req := httptest.NewRequest(http.MethodGet, "/api/apps/"+testAppID+"/groups", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	page := 1
	perPage := 5
	err := h.PaginateGroups(c, testAppID, codegen.PaginateGroupsParams{
		Page:    &page,
		Perpage: &perPage,
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var pg groupsPage
	err = json.Unmarshal(rec.Body.Bytes(), &pg)
	assert.NoError(t, err)
	assert.Greater(t, pg.TotalCount, 0)
	assert.NotEmpty(t, pg.Groups)

	// 2. Malformed/Validation (non-existent/invalid App ID)
	req2 := httptest.NewRequest(http.MethodGet, "/api/apps/invalid-app/groups", nil)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	err = h.PaginateGroups(c2, "invalid-app", codegen.PaginateGroupsParams{})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec2.Code)
}

func TestCreateGroup(t *testing.T) {
	h, a := setupHandler(t)
	defer a.Close()

	e := echo.New()
	track := "stable"
	desc := "Test group created via handler tests"

	// 1. Success
	bodyPayload := codegen.GroupConfig{
		Name:                      "New Test Group",
		Description:               &desc,
		PolicyMaxUpdatesPerPeriod: 5,
		PolicyPeriodInterval:      "30 minutes",
		PolicyTimezone:            "Europe/Berlin",
		PolicyUpdateTimeout:       "15 minutes",
		Track:                     &track,
	}
	bodyBytes, err := json.Marshal(bodyPayload)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/apps/"+testAppID+"/groups", bytes.NewReader(bodyBytes))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = h.CreateGroup(c, testAppID)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var g types.Group
	err = json.Unmarshal(rec.Body.Bytes(), &g)
	assert.NoError(t, err)
	assert.Equal(t, "New Test Group", g.Name)
	assert.Equal(t, testAppID, g.ApplicationID)

	// 2. App Not Found
	req2 := httptest.NewRequest(http.MethodPost, "/api/apps/invalid-app/groups", bytes.NewReader(bodyBytes))
	req2.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	err = h.CreateGroup(c2, "invalid-app")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec2.Code)

	// 3. Malformed JSON Body
	req3 := httptest.NewRequest(http.MethodPost, "/api/apps/"+testAppID+"/groups", bytes.NewReader([]byte("{invalid-json")))
	req3.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec3 := httptest.NewRecorder()
	c3 := e.NewContext(req3, rec3)

	err = h.CreateGroup(c3, testAppID)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec3.Code)
}

func TestGetGroup(t *testing.T) {
	h, a := setupHandler(t)
	defer a.Close()

	e := echo.New()

	// 1. Success
	req := httptest.NewRequest(http.MethodGet, "/api/apps/"+testAppID+"/groups/"+testGroupID, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.GetGroup(c, testAppID, testGroupID)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var g types.Group
	err = json.Unmarshal(rec.Body.Bytes(), &g)
	assert.NoError(t, err)
	assert.Equal(t, testGroupID, g.ID)

	// 2. Not Found
	req2 := httptest.NewRequest(http.MethodGet, "/api/apps/"+testAppID+"/groups/00000000-0000-0000-0000-000000000000", nil)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	err = h.GetGroup(c2, testAppID, "00000000-0000-0000-0000-000000000000")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec2.Code)

	// 3. Malformed ID (invalid UUID)
	req3 := httptest.NewRequest(http.MethodGet, "/api/apps/"+testAppID+"/groups/not-a-uuid", nil)
	rec3 := httptest.NewRecorder()
	c3 := e.NewContext(req3, rec3)

	err = h.GetGroup(c3, testAppID, "not-a-uuid")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec3.Code)
}

func TestUpdateGroup(t *testing.T) {
	h, a := setupHandler(t)
	defer a.Close()

	e := echo.New()
	track := "stable"
	desc := "Updated description"

	bodyPayload := codegen.GroupConfig{
		Name:                      "Updated Group Name",
		Description:               &desc,
		PolicyMaxUpdatesPerPeriod: 4,
		PolicyPeriodInterval:      "60 minutes",
		PolicyTimezone:            "Europe/Berlin",
		PolicyUpdateTimeout:       "15 minutes",
		Track:                     &track,
	}
	bodyBytes, err := json.Marshal(bodyPayload)
	require.NoError(t, err)

	// 1. Success
	req := httptest.NewRequest(http.MethodPut, "/api/apps/"+testAppID+"/groups/"+testGroupID, bytes.NewReader(bodyBytes))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = h.UpdateGroup(c, testAppID, testGroupID)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var g types.Group
	err = json.Unmarshal(rec.Body.Bytes(), &g)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Group Name", g.Name)

	// 2. Not Found
	req2 := httptest.NewRequest(http.MethodPut, "/api/apps/"+testAppID+"/groups/00000000-0000-0000-0000-000000000000", bytes.NewReader(bodyBytes))
	req2.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	err = h.UpdateGroup(c2, testAppID, "00000000-0000-0000-0000-000000000000")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec2.Code)

	// 3. App Not Found
	req3 := httptest.NewRequest(http.MethodPut, "/api/apps/invalid-app/groups/"+testGroupID, bytes.NewReader(bodyBytes))
	req3.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec3 := httptest.NewRecorder()
	c3 := e.NewContext(req3, rec3)

	err = h.UpdateGroup(c3, "invalid-app", testGroupID)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec3.Code)

	// 4. Malformed JSON Body
	req4 := httptest.NewRequest(http.MethodPut, "/api/apps/"+testAppID+"/groups/"+testGroupID, bytes.NewReader([]byte("{invalid-json")))
	req4.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec4 := httptest.NewRecorder()
	c4 := e.NewContext(req4, rec4)

	err = h.UpdateGroup(c4, testAppID, testGroupID)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec4.Code)
}

func TestDeleteGroup(t *testing.T) {
	h, a := setupHandler(t)
	defer a.Close()

	e := echo.New()

	// Helper: create a group to delete so we don't disrupt other tests.
	track := "stable"
	bodyPayload := codegen.GroupConfig{
		Name:                      "To Delete",
		PolicyMaxUpdatesPerPeriod: 1,
		PolicyPeriodInterval:      "15 minutes",
		PolicyTimezone:            "UTC",
		PolicyUpdateTimeout:       "15 minutes",
		Track:                     &track,
	}
	bodyBytes, err := json.Marshal(bodyPayload)
	require.NoError(t, err)

	reqCreate := httptest.NewRequest(http.MethodPost, "/api/apps/"+testAppID+"/groups", bytes.NewReader(bodyBytes))
	reqCreate.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recCreate := httptest.NewRecorder()
	cCreate := e.NewContext(reqCreate, recCreate)

	err = h.CreateGroup(cCreate, testAppID)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, recCreate.Code)

	var g types.Group
	err = json.Unmarshal(recCreate.Body.Bytes(), &g)
	require.NoError(t, err)
	createdGroupID := g.ID

	// 1. Success
	req := httptest.NewRequest(http.MethodDelete, "/api/apps/"+testAppID+"/groups/"+createdGroupID, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = h.DeleteGroup(c, testAppID, createdGroupID)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)

	// 2. Not Found
	req2 := httptest.NewRequest(http.MethodDelete, "/api/apps/"+testAppID+"/groups/00000000-0000-0000-0000-000000000000", nil)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	err = h.DeleteGroup(c2, testAppID, "00000000-0000-0000-0000-000000000000")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec2.Code)

	// 3. Malformed ID
	req3 := httptest.NewRequest(http.MethodDelete, "/api/apps/"+testAppID+"/groups/not-a-uuid", nil)
	rec3 := httptest.NewRecorder()
	c3 := e.NewContext(req3, rec3)

	err = h.DeleteGroup(c3, testAppID, "not-a-uuid")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec3.Code)
}

func TestGetGroupVersionTimeline(t *testing.T) {
	h, a := setupHandler(t)
	defer a.Close()

	e := echo.New()

	// 1. Success
	req := httptest.NewRequest(http.MethodGet, "/api/apps/"+testAppID+"/groups/"+testGroupID+"/version_timeline", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.GetGroupVersionTimeline(c, testAppID, testGroupID, codegen.GetGroupVersionTimelineParams{
		Duration: "30d",
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "MISS", rec.Header().Get("X-Cache"))

	// Verify caching on second call
	recCached := httptest.NewRecorder()
	cCached := e.NewContext(req, recCached)
	err = h.GetGroupVersionTimeline(cCached, testAppID, testGroupID, codegen.GetGroupVersionTimelineParams{
		Duration: "30d",
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, recCached.Code)
	assert.Equal(t, "HIT", recCached.Header().Get("X-Cache"))

	// 2. No data for an unknown group
	req2 := httptest.NewRequest(http.MethodGet, "/api/apps/"+testAppID+"/groups/00000000-0000-0000-0000-000000000000/version_timeline", nil)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	err = h.GetGroupVersionTimeline(c2, testAppID, "00000000-0000-0000-0000-000000000000", codegen.GetGroupVersionTimelineParams{
		Duration: "30d",
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec2.Code)

	// 3. Malformed/Validation (Invalid duration)
	req3 := httptest.NewRequest(http.MethodGet, "/api/apps/"+testAppID+"/groups/"+testGroupID+"/version_timeline?duration=invalid", nil)
	rec3 := httptest.NewRecorder()
	c3 := e.NewContext(req3, rec3)

	err = h.GetGroupVersionTimeline(c3, testAppID, testGroupID, codegen.GetGroupVersionTimelineParams{
		Duration: "invalid",
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec3.Code)
}

func TestGetGroupStatusTimeline(t *testing.T) {
	h, a := setupHandler(t)
	defer a.Close()

	e := echo.New()

	// 1. Success
	req := httptest.NewRequest(http.MethodGet, "/api/apps/"+testAppID+"/groups/"+testGroupID+"/status_timeline", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.GetGroupStatusTimeline(c, testAppID, testGroupID, codegen.GetGroupStatusTimelineParams{
		Duration: "30d",
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// 2. No data for an unknown group
	req2 := httptest.NewRequest(http.MethodGet, "/api/apps/"+testAppID+"/groups/00000000-0000-0000-0000-000000000000/status_timeline", nil)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	err = h.GetGroupStatusTimeline(c2, testAppID, "00000000-0000-0000-0000-000000000000", codegen.GetGroupStatusTimelineParams{
		Duration: "30d",
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec2.Code)

	// 3. Malformed/Validation (Invalid duration)
	req3 := httptest.NewRequest(http.MethodGet, "/api/apps/"+testAppID+"/groups/"+testGroupID+"/status_timeline?duration=invalid", nil)
	rec3 := httptest.NewRecorder()
	c3 := e.NewContext(req3, rec3)

	err = h.GetGroupStatusTimeline(c3, testAppID, testGroupID, codegen.GetGroupStatusTimelineParams{
		Duration: "invalid",
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec3.Code)
}

func TestGetGroupInstanceStats(t *testing.T) {
	h, a := setupHandler(t)
	defer a.Close()

	e := echo.New()

	// 1. Success
	req := httptest.NewRequest(http.MethodGet, "/api/apps/"+testAppID+"/groups/"+testGroupID+"/instances_stats", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.GetGroupInstanceStats(c, testAppID, testGroupID, codegen.GetGroupInstanceStatsParams{
		Duration: "30d",
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// 2. Not Found
	req2 := httptest.NewRequest(http.MethodGet, "/api/apps/"+testAppID+"/groups/00000000-0000-0000-0000-000000000000/instances_stats", nil)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	err = h.GetGroupInstanceStats(c2, testAppID, "00000000-0000-0000-0000-000000000000", codegen.GetGroupInstanceStatsParams{
		Duration: "30d",
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec2.Code)

	// 3. Malformed/Validation (Invalid duration)
	req3 := httptest.NewRequest(http.MethodGet, "/api/apps/"+testAppID+"/groups/"+testGroupID+"/instances_stats?duration=invalid", nil)
	rec3 := httptest.NewRecorder()
	c3 := e.NewContext(req3, rec3)

	err = h.GetGroupInstanceStats(c3, testAppID, testGroupID, codegen.GetGroupInstanceStatsParams{
		Duration: "invalid",
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec3.Code)
}

func TestGetGroupVersionBreakdown(t *testing.T) {
	h, a := setupHandler(t)
	defer a.Close()

	e := echo.New()

	// 1. Success
	req := httptest.NewRequest(http.MethodGet, "/api/apps/"+testAppID+"/groups/"+testGroupID+"/version_breakdown", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.GetGroupVersionBreakdown(c, testAppID, testGroupID)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// 2. No data for an unknown group
	req2 := httptest.NewRequest(http.MethodGet, "/api/apps/"+testAppID+"/groups/00000000-0000-0000-0000-000000000000/version_breakdown", nil)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	err = h.GetGroupVersionBreakdown(c2, testAppID, "00000000-0000-0000-0000-000000000000")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec2.Code)
}

func TestGetGroupInstances(t *testing.T) {
	h, a := setupHandler(t)
	defer a.Close()

	e := echo.New()

	// 1. Success
	req := httptest.NewRequest(http.MethodGet, "/api/apps/"+testAppID+"/groups/"+testGroupID+"/instances", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	page := 1
	perPage := 5
	status := 4 // complete
	err := h.GetGroupInstances(c, testAppID, testGroupID, codegen.GetGroupInstancesParams{
		Page:     &page,
		Perpage:  &perPage,
		Status:   status,
		Duration: "30d",
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// 2. App Not Found
	req2 := httptest.NewRequest(http.MethodGet, "/api/apps/invalid-app/groups/"+testGroupID+"/instances", nil)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	err = h.GetGroupInstances(c2, "invalid-app", testGroupID, codegen.GetGroupInstancesParams{
		Duration: "30d",
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec2.Code)

	// 3. Malformed/Validation (Invalid duration)
	req3 := httptest.NewRequest(http.MethodGet, "/api/apps/"+testAppID+"/groups/"+testGroupID+"/instances?duration=invalid", nil)
	rec3 := httptest.NewRecorder()
	c3 := e.NewContext(req3, rec3)

	err = h.GetGroupInstances(c3, testAppID, testGroupID, codegen.GetGroupInstancesParams{
		Duration: "invalid",
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec3.Code)
}

func TestGetGroupInstancesCount(t *testing.T) {
	h, a := setupHandler(t)
	defer a.Close()

	e := echo.New()

	// 1. Success
	req := httptest.NewRequest(http.MethodGet, "/api/apps/"+testAppID+"/groups/"+testGroupID+"/instances/count", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.GetGroupInstancesCount(c, testAppID, testGroupID, codegen.GetGroupInstancesCountParams{
		Duration: "30d",
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// 2. App Not Found
	req2 := httptest.NewRequest(http.MethodGet, "/api/apps/invalid-app/groups/"+testGroupID+"/instances/count", nil)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	err = h.GetGroupInstancesCount(c2, "invalid-app", testGroupID, codegen.GetGroupInstancesCountParams{
		Duration: "30d",
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec2.Code)

	// 3. Malformed/Validation (Invalid duration)
	req3 := httptest.NewRequest(http.MethodGet, "/api/apps/"+testAppID+"/groups/"+testGroupID+"/instances/count?duration=invalid", nil)
	rec3 := httptest.NewRecorder()
	c3 := e.NewContext(req3, rec3)

	err = h.GetGroupInstancesCount(c3, testAppID, testGroupID, codegen.GetGroupInstancesCountParams{
		Duration: "invalid",
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec3.Code)
}
