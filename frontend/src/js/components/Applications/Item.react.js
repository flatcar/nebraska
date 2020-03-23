import Grid from '@material-ui/core/Grid';
import { makeStyles } from '@material-ui/core/styles';
import PropTypes from 'prop-types';
import React from 'react';
import { applicationsStore } from '../../stores/Stores';
import { CardFeatureLabel, CardHeader, CardLabel } from '../Common/Card';
import ListItem from '../Common/ListItem';
import MoreMenu from '../Common/MoreMenu';
import ChannelsList from './ApplicationItemChannelsList.react';
import GroupsList from './ApplicationItemGroupsList.react';

const useStyles = makeStyles(theme => ({
  itemSection: {
    padding: '1em'
  },
}));

function Item(props) {
  const classes = useStyles();
  const description = props.application.description || 'No description provided';
  const channels = props.application.channels || [];
  const groups = props.application.groups || [];
  const instances = props.application.instances.count || 'None';
  const appID = props.application ? props.application.id : '';

  function updateApplication() {
    props.handleUpdateApplication(props.application.id);
  }

  function deleteApplication() {
    const confirmationText = 'Are you sure you want to delete this application?';
    if (window.confirm(confirmationText)) {
      applicationsStore.deleteApplication(props.application.id);
    }
  }

  return(
    <ListItem>
      <Grid container>
        <Grid item xs={12}>
          <CardHeader
            cardMainLinkLabel={props.application.name}
            cardMainLinkPath={{pathname: `/apps/${appID}`}}
            cardId={appID}
            cardDescription={description}
          >
            <MoreMenu options={[
              {
              'label': 'Edit',
              'action': updateApplication,
              },
              {
                'label': 'Delete',
                'action': deleteApplication,
              }
            ]} />
          </CardHeader>
        </Grid>
        <Grid item xs={12}>
          <Grid
            container
            spacing={2}
            justify="space-between"
            className={classes.itemSection}
          >
            <Grid item xs={6}>
              <CardFeatureLabel>Instances:</CardFeatureLabel>&nbsp;
              <CardLabel>{instances}</CardLabel>
            </Grid>
            <Grid item xs={6}>
              <CardFeatureLabel>Groups:</CardFeatureLabel>&nbsp;
              <CardLabel>{groups.length == 0 ? 'None' : groups.length}</CardLabel>
              <GroupsList
                groups={groups}
                appID={props.application.id}
                appName={props.application.name} />
            </Grid>
            <Grid item xs={12}>
              <CardFeatureLabel>Channels:</CardFeatureLabel>&nbsp;
              {channels.length > 0 ?
                <ChannelsList channels={channels} />
                :
                <CardLabel>None</CardLabel>
              }
            </Grid>
          </Grid>
        </Grid>
      </Grid>
    </ListItem>
  );
}

Item.propTypes = {
  application: PropTypes.object.isRequired,
  handleUpdateApplication: PropTypes.func.isRequired
};

export default Item;
