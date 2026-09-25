package admin

import (
	"database/sql"
	"fmt"
	"regexp"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
	"gopkg.in/guregu/null.v4"

	"github.com/flatcar/nebraska/backend/pkg/api/types"
)

// addApp validates and inserts the application using the transaction
// provided. The caller owns the transaction and is responsible for committing
// it and for invalidating the application cache afterwards.
func (s *Service) addApp(app *types.Application, tx *sqlx.Tx) error {
	if err := validateProductID(app.ProductID); err != nil {
		return fmt.Errorf("cannot add application %v: %w", app.ID, err)
	}
	query, _, err := goqu.Insert("application").
		Cols("name", "product_id", "description", "team_id").
		Vals(goqu.Vals{app.Name, app.ProductID, app.Description, app.TeamID}).
		Returning(goqu.T("application").All()).
		ToSQL()
	if err != nil {
		return err
	}
	return tx.QueryRowx(query).StructScan(app)
}

// AddApp registers the provided application.
func (s *Service) AddApp(app *types.Application) (*types.Application, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			l.Error().Err(err).Msg("AddApp - could not roll back")
		}
	}()

	if err := s.addApp(app, tx); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	s.ClearCachedAppIDs()
	return app, nil
}

// AddAppCloning registers the provided application, cloning the groups and
// channels from an existing application. Channels' packages will be set to null
// as packages won't be cloned.
//
// The application and every cloned channel and group are written in a single
// transaction, so a failure to copy any one of them rolls the whole clone back
// and returns the error. A successful return therefore always means a complete
// copy, never an application that silently lost some of its channels or has
// groups pointing at channels that were never created.
func (s *Service) AddAppCloning(app *types.Application, sourceAppID string) (*types.Application, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			l.Error().Err(err).Msg("AddAppCloning - could not roll back")
		}
	}()

	if err := s.addApp(app, tx); err != nil {
		return nil, err
	}

	if sourceAppID != "" {
		// The source application is pre-existing, committed data and isn't
		// touched by this transaction, so it's read on the shared connection.
		sourceApp, err := s.GetApp(sourceAppID)
		if err != nil {
			return nil, fmt.Errorf("cannot clone application: could not get source app %v: %w", sourceAppID, err)
		}

		channelsIDsMappings := make(map[string]null.String)

		for _, channel := range sourceApp.Channels {
			originalChannelID := channel.ID
			channel.ApplicationID = app.ID
			channel.PackageID = null.String{}
			if err := s.addChannel(channel, tx); err != nil {
				return nil, fmt.Errorf("cannot clone application: could not copy channel %v: %w", originalChannelID, err)
			}
			channelsIDsMappings[originalChannelID] = null.StringFrom(channel.ID)
		}

		for _, group := range sourceApp.Groups {
			originalGroupID := group.ID
			group.ApplicationID = app.ID
			if group.ChannelID.String != "" {
				group.ChannelID = channelsIDsMappings[group.ChannelID.String]
			}
			group.PolicyUpdatesEnabled = true
			group.ID = ""
			if err := s.addGroup(group, tx); err != nil {
				return nil, fmt.Errorf("cannot clone application: could not copy group %v: %w", originalGroupID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	s.ClearCachedAppIDs()
	s.UpdateCachedGroups()
	return app, nil
}

func validateProductID(productID null.String) error {
	if productID.Ptr() == nil {
		return nil
	}

	if len(*productID.Ptr()) > 155 {
		return fmt.Errorf("product ID %v is not valid (max length 155)", *productID.Ptr())
	}

	// This regex matches an ID that matches
	// * At least two segments.
	// * All characters must be alphanumeric, a dash.
	// Each segment must start with a letter.
	// Each segment must not end with a dash.
	regMatcher := "^[a-zA-Z]+([a-zA-Z0-9\\-]*[a-zA-Z0-9])*(\\.[a-zA-Z]+([a-zA-Z0-1\\-]*[a-zA-Z0-9])*)+$"
	matches, err := regexp.MatchString(regMatcher, *productID.Ptr())
	if err != nil {
		return err
	}

	if !matches {
		return fmt.Errorf("product ID %v is not valid (has to be in the form e.g. io.example.App)", *productID.Ptr())
	}

	return nil
}

// UpdateApp updates an existing application using the content of the
// application provided.
func (s *Service) UpdateApp(app *types.Application) error {
	if err := validateProductID(app.ProductID); err != nil {
		return fmt.Errorf("cannot add application %v: %w", app.ID, err)
	}

	query, _, err := goqu.Update("application").
		Set(
			goqu.Record{
				"name":        app.Name,
				"product_id":  app.ProductID,
				"description": app.Description,
			},
		).
		Where(goqu.C("id").Eq(app.ID)).
		ToSQL()
	if err != nil {
		return err
	}
	result, err := s.db.Exec(query)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return types.ErrNoRowsAffected
	}

	s.ClearCachedAppIDs()
	return nil
}

// DeleteApp removes the application identified by the id provided.
func (s *Service) DeleteApp(appID string) error {
	query, _, err := goqu.Delete("application").Where(goqu.C("id").Eq(appID)).ToSQL()
	if err != nil {
		return err
	}
	result, err := s.db.Exec(query)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return types.ErrNoRowsAffected
	}

	s.ClearCachedAppIDs()
	return nil
}
