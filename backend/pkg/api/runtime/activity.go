package runtime

import (
	"github.com/doug-martin/goqu/v9"
)

// newGroupActivityEntry creates a new activity entry related to a specific
// group.
func (s *Service) newGroupActivityEntry(class int, severity int, version, appID, groupID string) error {
	query, _, err := goqu.Insert("activity").
		Cols("class", "severity", "version", "application_id", "group_id").
		Vals(goqu.Vals{class, severity, version, appID, groupID}).
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

// newInstanceActivityEntry creates a new activity entry related to a specific
// instance.
func (s *Service) newInstanceActivityEntry(class int, severity int, version, appID, groupID, instanceID string) error {
	query, _, err := goqu.Insert("activity").
		Cols("class", "severity", "version", "application_id", "group_id", "instance_id").
		Vals(goqu.Vals{class, severity, version, appID, groupID, instanceID}).
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
