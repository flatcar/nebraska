package admin

import (
	"fmt"
	"regexp"

	"github.com/doug-martin/goqu/v9"
	"gopkg.in/guregu/null.v4"

	"github.com/flatcar/nebraska/backend/pkg/api/types"
)

// AddApp registers the provided application.
func (s *Service) AddApp(app *types.Application) (*types.Application, error) {
	if err := validateProductID(app.ProductID); err != nil {
		return nil, fmt.Errorf("cannot add application %v: %w", app.ID, err)
	}
	query, _, err := goqu.Insert("application").
		Cols("name", "product_id", "description", "team_id").
		Vals(goqu.Vals{app.Name, app.ProductID, app.Description, app.TeamID}).
		Returning(goqu.T("application").All()).
		ToSQL()
	if err != nil {
		return nil, err
	}
	err = s.db.QueryRowx(query).StructScan(app)
	if err != nil {
		return nil, err
	}

	s.ClearCachedAppIDs()
	return app, nil
}

// AddAppCloning registers the provided application, cloning the groups and
// channels from an existing application. Channels' packages will be set to null
// as packages won't be cloned.
func (s *Service) AddAppCloning(app *types.Application, sourceAppID string) (*types.Application, error) {
	app, err := s.AddApp(app)
	if err != nil {
		return nil, err
	}

	// NOTE: cloning operation is not transactional and something could go wrong

	if sourceAppID != "" {
		sourceApp, err := s.GetApp(sourceAppID)
		if err != nil {
			l.Error().Err(err).Msg("AddAppCloning - could not get source app")
			return app, nil
		}

		channelsIDsMappings := make(map[string]null.String)

		for _, channel := range sourceApp.Channels {
			originalChannelID := channel.ID
			channel.ApplicationID = app.ID
			channel.PackageID = null.String{}
			channelCopy, err := s.AddChannel(channel)
			if err != nil {
				l.Error().Err(err).Msg("AddAppCloning - could not add channel")
				return app, nil // FIXME - think about what we should return to the caller
			}
			channelsIDsMappings[originalChannelID] = null.StringFrom(channelCopy.ID)
		}

		for _, group := range sourceApp.Groups {
			group.ApplicationID = app.ID
			if group.ChannelID.String != "" {
				group.ChannelID = channelsIDsMappings[group.ChannelID.String]
			}
			group.PolicyUpdatesEnabled = true
			group.ID = ""
			if _, err := s.AddGroup(group); err != nil {
				l.Error().Err(err).Msg("AddAppCloning - could not add group")
				return app, nil // FIXME - think about what we should return to the caller
			}
		}
	}
	// Even though AddApp will invalidate the cache, we need to do it again here
	// to prevent eventual race issues.
	s.ClearCachedAppIDs()
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
