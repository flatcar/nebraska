package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/doug-martin/goqu/v9"
	"gopkg.in/guregu/null.v4"
)

const (
	activityPackageNotFound int = 1 + iota
	activityRolloutStarted
	activityRolloutFinished
	activityRolloutFailed
	activityInstanceUpdateFailed
	activityChannelPackageUpdated
)

const (
	activitySuccess int = 1 + iota
	activityInfo
	activityWarning
	activityError
)

// activityContext represents the context of a given activity entry.
type activityContext struct {
	appID      string
	groupID    string
	channelID  string
	instanceID string
}

// Activity represents a Nebraska activity entry.
type Activity struct {
	AppID           null.String `db:"application_id" json:"app_id"`
	GroupID         null.String `db:"group_id" json:"group_id"`
	CreatedTs       time.Time   `db:"created_ts" json:"created_ts"`
	Class           int         `db:"class" json:"class"`
	Severity        int         `db:"severity" json:"severity"`
	Version         string      `db:"version" json:"version"`
	ApplicationName string      `db:"application_name" json:"application_name"`
	GroupName       null.String `db:"group_name" json:"group_name"`
	ChannelName     null.String `db:"channel_name" json:"channel_name"`
	InstanceID      null.String `db:"instance_id" json:"instance_id"`
}

// ActivityQueryParams represents a helper structure used to pass a set of
// parameters when querying activity entries.
type ActivityQueryParams struct {
	AppID      string    `db:"application_id"`
	GroupID    string    `db:"group_id"`
	ChannelID  string    `db:"channel_id"`
	InstanceID string    `db:"instance_id"`
	Version    string    `db:"version"`
	Severity   int       `db:"severity"`
	Start      time.Time `db:"start"`
	End        time.Time `db:"end"`
	Page       uint64    `json:"page"`
	PerPage    uint64    `json:"perpage"`
}

// GetActivity returns a list of activity entries that match the specified
// criteria in the query parameters.
func (api *API) GetActivity(teamID string, p ActivityQueryParams) ([]*Activity, error) {
	var activityEntries []*Activity
	query, _, err := api.activityQuery(teamID, p).ToSQL()
	if err != nil {
		return nil, err
	}
	rows, err := api.db.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		activityEntry := &Activity{}
		err := rows.StructScan(activityEntry)
		if err != nil {
			return nil, err
		}
		activityEntries = append(activityEntries, activityEntry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return activityEntries, nil
}

// activityQuery returns a SelectDataset prepared to return all activity
// entries that match the criteria provided in ActivityQueryParams.
func (api *API) activityQuery(teamID string, p ActivityQueryParams) *goqu.SelectDataset {
	p.Page, p.PerPage = validatePaginationParams(p.Page, p.PerPage)

	var start, end time.Time
	if !p.Start.IsZero() {
		start = p.Start.UTC()
	} else {
		start = time.Now().UTC().AddDate(0, 0, -3)
	}
	if !p.End.IsZero() {
		end = p.End.UTC()
	} else {
		end = time.Now().UTC()
	}
	query := goqu.From(goqu.L(`
	activity AS a 
	INNER JOIN application AS app ON (a.application_id = app.id)
	LEFT JOIN groups AS g ON (a.group_id = g.id)
	LEFT JOIN channel AS c ON (a.channel_id = c.id)
`)).Select("a.application_id", "a.group_id", "a.created_ts", "a.class", "a.severity", "a.version", "a.instance_id",
		goqu.I("app.name").As("application_name"), goqu.I("g.name").
			As("group_name"), goqu.I("c.name").As("channel_name")).
		Where(goqu.I("app.team_id").Eq(teamID), goqu.And(goqu.I("a.created_ts").Gte(start),
			goqu.I("a.created_ts").Lt(end)))

	if p.AppID != "" {
		query = query.Where(goqu.I("app.id").Eq(p.AppID))
	}

	if p.GroupID != "" {
		query = query.Where(goqu.I("g.id").Eq(p.GroupID))
	}

	if p.ChannelID != "" {
		query = query.Where(goqu.I("c.id").Eq(p.ChannelID))
	}

	if p.InstanceID != "" {
		query = query.Where(goqu.I("a.instance_id").Eq(p.InstanceID))
	} else {
		query = query.Where(goqu.L(ignoreFakeInstanceCondition("a.instance_id")))
	}

	if p.Version != "" {
		query = query.Where(goqu.I("a.version").Eq(p.Version))
	}

	if p.Severity != 0 {
		query = query.Where(goqu.I("a.severity").Eq(p.Severity))
	}
	limit, offset := sqlPaginate(p.Page, p.PerPage)
	query = query.Limit(limit).
		Offset(offset).Order(goqu.I("a.created_ts").Desc())

	return query
}

func (api *API) hasRecentActivity(class int, p ActivityQueryParams) bool {
	recent := time.Now().UTC().Add(time.Duration(-24*60) * time.Minute)

	query := goqu.From("activity").
		Select("*").
		Where(goqu.C("class").Eq(class)).
		Where(goqu.C("created_ts").Gt(recent))

	if p.Severity != 0 {
		query = query.Where(goqu.C("severity").Eq(p.Severity))
	}

	if p.Version != "" {
		query = query.Where(goqu.C("version").Eq(p.Version))
	}

	if p.GroupID != "" {
		query = query.Where(goqu.C("group_id").Eq(p.GroupID))
	}

	if p.AppID != "" {
		query = query.Where(goqu.I("application_id").Eq(p.AppID))
	}

	if p.ChannelID != "" {
		query = query.Where(goqu.I("channel_id").Eq(p.ChannelID))
	}

	if p.InstanceID != "" {
		query = query.Where(goqu.C("instance_id").Eq(p.InstanceID))
	} else {
		query = query.Where(goqu.L(ignoreFakeInstanceCondition("instance_id")))
	}

	sql, _, err := query.ToSQL()
	if err != nil {
		return false
	}

	rows, err := api.db.Queryx(sql)
	if err == nil {
		rows.Close()
	}
	return err == nil
}

// newGroupActivityEntry creates a new activity entry related to a specific
// group.
func (api *API) newGroupActivityEntry(class int, severity int, version, appID, groupID string) error {
	query, _, err := goqu.Insert("activity").
		Cols("class", "severity", "version", "application_id", "group_id").
		Vals(goqu.Vals{class, severity, version, appID, groupID}).
		ToSQL()
	if err != nil {
		return err
	}
	_, err = api.db.Exec(query)

	if err != nil {
		return err
	}

	ctx := &activityContext{
		appID:   appID,
		groupID: groupID,
	}
	go api.postHipchat(class, severity, version, ctx)

	return nil
}

// newChannelActivityEntry creates a new activity entry related to a specific
// channel.
func (api *API) newChannelActivityEntry(class int, severity int, version, appID, channelID string) error {
	query, _, err := goqu.Insert("activity").
		Cols("class", "severity", "version", "application_id", "channel_id").
		Vals(goqu.Vals{class, severity, version, appID, channelID}).
		ToSQL()
	if err != nil {
		return err
	}
	_, err = api.db.Exec(query)

	if err != nil {
		return err
	}

	ctx := &activityContext{
		appID:     appID,
		channelID: channelID,
	}
	go api.postHipchat(class, severity, version, ctx)

	return nil
}

// newInstanceActivityEntry creates a new activity entry related to a specific
// instance.
func (api *API) newInstanceActivityEntry(class int, severity int, version, appID, groupID, instanceID string) error {
	query, _, err := goqu.Insert("activity").
		Cols("class", "severity", "version", "application_id", "group_id", "instance_id").
		Vals(goqu.Vals{class, severity, version, appID, groupID, instanceID}).
		ToSQL()
	if err != nil {
		return err
	}
	_, err = api.db.Exec(query)

	if err != nil {
		return err
	}

	ctx := &activityContext{
		appID:      appID,
		groupID:    groupID,
		instanceID: instanceID,
	}
	go api.postHipchat(class, severity, version, ctx)

	return nil
}

// EXPERIMENTAL! This is an experiment playing a bit with Hipchat
//
// postHipchat builds and posts a message representing the activity entry
// provided to Hipchat. This is just an experiment.
func (api *API) postHipchat(class, severity int, version string, ctx *activityContext) {
	room := os.Getenv("CR_HIPCHAT_ROOM")
	token := os.Getenv("CR_HIPCHAT_TOKEN")
	if room == "" || token == "" {
		return
	}

	var msg bytes.Buffer
	var color string

	app, _ := api.GetApp(ctx.appID)
	fmt.Fprintf(&msg, "<b>%s</b> ", app.Name)

	if ctx.groupID != "" {
		group, _ := api.GetGroup(ctx.groupID)
		fmt.Fprintf(&msg, "> <b>%s</b>", group.Name)
	}
	fmt.Fprint(&msg, "<br/>")

	switch class {
	case activityPackageNotFound:
		fmt.Fprint(&msg, "An update request could not be processed because the group's channel is not linked to any package")
		color = "red"
	case activityRolloutStarted:
		fmt.Fprintf(&msg, "Version <i>%s</i> roll out started", version)
		color = "purple"
	case activityRolloutFinished:
		fmt.Fprintf(&msg, "Version <i>%s</i> successfully rolled out", version)
		color = "green"
	case activityRolloutFailed:
		fmt.Fprintf(&msg, "There was an error rolling out version <i>%s</i> as the first update attempt failed. Group's updates have been disabled", version)
		color = "red"
	case activityInstanceUpdateFailed:
		instance, _ := api.GetInstance(ctx.instanceID, ctx.appID)
		fmt.Fprintf(&msg, "Instance <i>%s</i> reported an error while processing update to version <i>%s</i>", instance.IP, version)
		color = "yellow"
	case activityChannelPackageUpdated:
		channel, _ := api.GetChannel(ctx.channelID)
		fmt.Fprintf(&msg, "Channel <i>%s</i> is now pointing to version <i>%s</i>", channel.Name, version)
		color = "purple"
	}

	body := map[string]interface{}{
		"message_format": "html",
		"message":        msg.String(),
		"color":          color,
		"notify":         true,
	}
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return
	}

	url := "https://api.hipchat.com/v2/room/%s/notification?auth_token=%s"
	resp, err := http.Post(fmt.Sprintf(url, room, token), "application/json", bytes.NewReader(bodyJSON))
	if err != nil {
		return
	}
	resp.Body.Close()
}
