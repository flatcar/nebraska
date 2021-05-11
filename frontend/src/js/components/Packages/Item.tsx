import cancelIcon from '@iconify/icons-mdi/cancel';
import cubeOutline from '@iconify/icons-mdi/cube-outline';
import { InlineIcon } from '@iconify/react';
import { Theme } from '@material-ui/core';
import Grid from '@material-ui/core/Grid';
import ListItem from '@material-ui/core/ListItem';
import ListItemIcon from '@material-ui/core/ListItemIcon';
import ListItemSecondaryAction from '@material-ui/core/ListItemSecondaryAction';
import ListItemText from '@material-ui/core/ListItemText';
import Typography from '@material-ui/core/Typography';
import makeStyles from '@material-ui/styles/makeStyles';
import PropTypes from 'prop-types';
import React from 'react';
import _ from 'underscore';
import { Channel, Package } from '../../api/apiDataTypes';
import flatcarIcon from '../../icons/flatcar-logo.json';
import { applicationsStore } from '../../stores/Stores';
import { ARCHES, cleanSemverVersion } from '../../utils/helpers';
import ChannelAvatar from '../Channels/ChannelAvatar';
import Label from '../Common/Label';
import MoreMenu from '../Common/MoreMenu';

//@todo visit this again
//@ts-ignore
const useStyles = makeStyles((theme: Theme) => ({
  packageName: {
    fontSize: '1.1em',
  },
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
  },
}));

const containerIcons: {
  [key: string]: any;
} = {
  1: { icon: flatcarIcon, name: 'Flatcar' },
  other: { icon: cubeOutline, name: 'Other' },
};

function Item(props: {
  packageItem: Package;
  channels: Channel[];
  handleUpdatePackage: (pkgId: string) => void;
}) {
  const classes = useStyles();
  const createdDate = new Date(props.packageItem.created_ts as string);
  const date = createdDate.toLocaleString('default', {
    year: 'numeric',
    month: 'numeric',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
  const type = props.packageItem.type || 1;
  const processedChannels = _.where(props.channels, { package_id: props.packageItem.id });
  let blacklistInfo: string | null = null;
  const item = type in containerIcons ? containerIcons[type] : containerIcons.other;

  if (props.packageItem.channels_blacklist) {
    const channelsList = _.map(props.packageItem.channels_blacklist, (channel, index) => {
      return _.findWhere(props.channels, { id: channel })?.name;
    });
    blacklistInfo = channelsList.join(' - ');
  }

  function deletePackage() {
    const confirmationText = 'Are you sure you want to delete this package?';
    if (window.confirm(confirmationText)) {
      applicationsStore.deletePackage(
        props.packageItem.application_id,
        props.packageItem.id as string
      );
    }
  }

  function updatePackage() {
    props.handleUpdatePackage(props.packageItem.id as string);
  }

  function makeItemSecondaryInfo() {
    return (
      <Grid container direction="column">
        <Grid item>
          <Typography component="span" className={classes.subtitle}>
            Version:
          </Typography>
          &nbsp;
          {`${cleanSemverVersion(props.packageItem.version)} (${ARCHES[props.packageItem.arch]})`}
        </Grid>
        {processedChannels.length > 0 && (
          <Grid item>
            <Typography component="span" className={classes.subtitle}>
              Channels:
            </Typography>
            &nbsp;
            {processedChannels.map((channel, i) => {
              return (
                <span className={classes.channelLabel} key={i}>
                  <ChannelAvatar color={channel.color} size="10px" />
                  &nbsp;
                  {channel.name}
                </span>
              );
            })}
          </Grid>
        )}
        <Grid item>
          <Typography component="span" className={classes.subtitle}>
            Released:
          </Typography>
          &nbsp;
          {date}
        </Grid>
        {props.packageItem.channels_blacklist && (
          <Grid item>
            {props.packageItem.channels_blacklist && (
              <Label>
                <InlineIcon icon={cancelIcon} width="10" height="10" /> {blacklistInfo}
              </Label>
            )}
          </Grid>
        )}
      </Grid>
    );
  }

  return (
    <ListItem dense alignItems="flex-start">
      <ListItemIcon className={classes.packageIcon}>
        <InlineIcon icon={item.icon} width="35" height="35" />
      </ListItemIcon>
      <ListItemText
        disableTypography
        primaryTypographyProps={{ className: classes.packageName }}
        primary={<Typography>{item.name}</Typography>}
        secondary={makeItemSecondaryInfo()}
      />
      <ListItemSecondaryAction>
        <MoreMenu
          options={[
            { label: 'Edit', action: updatePackage },
            { label: 'Delete', action: deletePackage },
          ]}
        />
      </ListItemSecondaryAction>
    </ListItem>
  );
}

Item.propTypes = {
  packageItem: PropTypes.object.isRequired,
  channels: PropTypes.array,
  handleUpdatePackage: PropTypes.func.isRequired,
};

export default Item;
