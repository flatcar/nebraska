import coreosIcon from '@iconify/icons-logos/coreos-icon';
import cancelIcon from '@iconify/icons-mdi/cancel';
import cubeOutline from '@iconify/icons-mdi/cube-outline';
import { InlineIcon } from '@iconify/react';
import Grid from '@material-ui/core/Grid';
import ListItem from '@material-ui/core/ListItem';
import ListItemIcon from '@material-ui/core/ListItemIcon';
import ListItemSecondaryAction from '@material-ui/core/ListItemSecondaryAction';
import ListItemText from '@material-ui/core/ListItemText';
import Typography from '@material-ui/core/Typography';
import makeStyles from '@material-ui/styles/makeStyles';
import moment from 'moment';
import PropTypes from 'prop-types';
import React from 'react';
import _ from 'underscore';
import { cleanSemverVersion } from '../../constants/helpers';
import { applicationsStore } from '../../stores/Stores';
import Label from '../Common/Label';
import MoreMenu from '../Common/MoreMenu';
import VersionBullet from '../Common/VersionBullet.react';

const useStyles = makeStyles(theme => ({
  subtitle: {
    fontSize: '.9em',
    textTransform: 'uppercase',
    fontWeight: '300',
    paddingRight: '.05em',
    color: theme.palette.grey['500'],
  },
  packageIcon: {
    minWidth: '40px',
  },
  channelLabel: {
    marginRight: '5px',
  }
}));

const containerIcons = {
  1: {icon: coreosIcon, name: 'CoreOS'},
  other: {icon: cubeOutline, name: 'Other'},
};

function Item(props) {
  const classes = useStyles();
  let date = moment.utc(props.packageItem.created_ts).local().format("hh:mma, DD/MM");
  let type = props.packageItem.type || 1;
  let processedChannels = _.where(props.channels, {package_id: props.packageItem.id});
  let blacklistInfo = null;
  let item = type in containerIcons ? containerIcons[type] : containerIcons.other;

  if (props.packageItem.channels_blacklist) {
    let channelsList = _.map(props.packageItem.channels_blacklist, (channel, index) => {
      return (_.findWhere(props.channels, {id: channel})).name;
    })
    blacklistInfo = channelsList.join(' - ');
  }

  function deletePackage() {
    let confirmationText = "Are you sure you want to delete this package?"
    if (confirm(confirmationText)) {
      applicationsStore.deletePackage(props.packageItem.application_id, props.packageItem.id);
    }
  }

  function updatePackage() {
    props.handleUpdatePackage(props.packageItem.id);
  }

  function makeItemSecondaryInfo() {
    return (
      <Grid container direction="column">
        <Grid item>
          <Typography component="span" className={classes.subtitle}>Version:</Typography>&nbsp;
          {cleanSemverVersion(props.packageItem.version)}
        </Grid>
        {processedChannels.length > 0 &&
          <Grid item>
            <Typography component="span" className={classes.subtitle}>Channels:</Typography>&nbsp;
            {processedChannels.map((channel, i) => {
              return (<span className={classes.channelLabel}>
                        <VersionBullet channel={channel} key={"packageItemBullet_" + i} />
                        {channel.name}
                      </span>
              );
              })
            }
          </Grid>
        }
        <Grid item>
          <Typography component="span" className={classes.subtitle}>Released:</Typography>&nbsp;
          {date}
        </Grid>
        {props.packageItem.channels_blacklist &&
          <Grid item>
            {props.packageItem.channels_blacklist &&
              <Label><InlineIcon icon={cancelIcon} width="10" height="10" /> { blacklistInfo }</Label>
            }
          </Grid>
        }
      </Grid>
    );
  }

  return (
    <ListItem dense alignItems="flex-start">
      <ListItemIcon className={classes.packageIcon}>
        <InlineIcon icon={item.icon} width="25" height="25" />
      </ListItemIcon>
      <ListItemText
        primary={item.name}
        secondary={makeItemSecondaryInfo()}
      />
      <ListItemSecondaryAction>
        <MoreMenu options={[
            {label: 'Edit', action: updatePackage},
            {label: 'Delete', action: deletePackage}
          ]}
        />
      </ListItemSecondaryAction>
    </ListItem>
  );
}

Item.propTypes = {
  packageItem: PropTypes.object.isRequired,
  channels: PropTypes.array,
  handleUpdatePackage: PropTypes.func.isRequired
}

export default Item
