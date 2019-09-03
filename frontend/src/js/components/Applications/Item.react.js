import Grid from '@material-ui/core/Grid';
import ListItem from '@material-ui/core/ListItem';
import { makeStyles } from '@material-ui/core/styles';
import PropTypes from 'prop-types';
import React from 'react';
import { applicationsStore } from '../../stores/Stores';
import { CardFeatureLabel, CardHeader } from '../Common/Card';
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
  let description = props.application.description || 'No description provided';
  let channels = props.application.channels || [];
  let groups = props.application.groups || [];
  let instances = props.application.instances.count || 0;
  let appID = props.application ? props.application.id : '';

  function updateApplication() {
    props.handleUpdateApplication(props.application.id)
  }

  function deleteApplication() {
    let confirmationText = "Are you sure you want to delete this application?"
    if (confirm(confirmationText)) {
      applicationsStore.deleteApplication(props.application.id)
    }
  }

  return(
    <ListItem disableGutters divider>
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
                {instances}
            </Grid>
            <Grid item xs={6}>
              <CardFeatureLabel>Groups:</CardFeatureLabel>&nbsp;
                {groups.length}
              <GroupsList
                groups={groups}
                appID={props.application.id}
                appName={props.application.name} />
            </Grid>
            <Grid item xs={12}>
              <CardFeatureLabel>Channels:</CardFeatureLabel>
              <ChannelsList channels={channels} />
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
}

export default Item
