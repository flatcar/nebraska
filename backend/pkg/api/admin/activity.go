package admin

import (
	"github.com/doug-martin/goqu/v9"
)

// newChannelActivityEntry creates a new admin_activity entry related to a
// specific channel.
func (s *Service) newChannelActivityEntry(class int, severity int, version, appID, channelID string) error {
	query, _, err := goqu.Insert("admin_activity").
		Cols("class", "severity", "version", "application_id", "channel_id").
		Vals(goqu.Vals{class, severity, version, appID, channelID}).
		ToSQL()
	if err != nil {
		return err
	}
	_, err = s.db.Exec(query)

	if err != nil {
		return err
	}

	return nil
}
