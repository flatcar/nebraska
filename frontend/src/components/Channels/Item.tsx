import { Box, Grid, makeStyles, Tooltip, useTheme } from '@material-ui/core';
import ListItem from '@material-ui/core/ListItem';
import ListItemSecondaryAction from '@material-ui/core/ListItemSecondaryAction';
import ListItemText from '@material-ui/core/ListItemText';
import ScheduleIcon from '@material-ui/icons/Schedule';
import React from 'react';
import { useTranslation } from 'react-i18next';
import { Channel, Package } from '../../api/apiDataTypes';
import { applicationsStore } from '../../stores/Stores';
import { ARCHES, cleanSemverVersion } from '../../utils/helpers';
import MoreMenu from '../common/MoreMenu';
import ChannelAvatar from './ChannelAvatar';

const useStyles = makeStyles({
  root: {
    margin: '0px',
  },
});

function Item(props: {
  channel: Channel;
  packages?: Package[];
  showArch?: boolean;
  isAppView?: boolean;
  handleUpdateChannel?: (channelID: string) => void;
}) {
  const theme = useTheme();
  const classes = useStyles();
  const { t } = useTranslation();
  const { channel, showArch = true, isAppView = false, ...others } = props;
  const name = channel.name;
  const version = channel.package
    ? cleanSemverVersion(channel.package.version)
    : t('channels|No package');

  function deleteChannel() {
    const confirmationText = t('channels|Are you sure you want to delete this channel?');
    if (window.confirm(confirmationText)) {
      applicationsStore.deleteChannel(channel.application_id, channel.id);
    }
  }

  function updateChannel() {
    if (props.handleUpdateChannel) {
      props.handleUpdateChannel(channel.id);
    }
  }

  function getSecondaryText() {
    let text = '';

    if (version) {
      text = cleanSemverVersion(version);
    }

    if (showArch) {
      if (text !== '') {
        text += ' ';
      }

      text += `(${ARCHES[channel.arch]})`;
    }
    const date = channel.package ? new Date(channel.package.created_ts) : null;
    return (
      <Box display="flex" ml={1}>
        <Box>{text}</Box>
        {date && (
          <Box pl={2}>
            <Box display="flex">
              <Box>
                <Tooltip title={t('channels|Release date') || ''}>
                  <ScheduleIcon fontSize="small" />
                </Tooltip>
              </Box>
              <Box pl={1}>{t('{{date, date}}', { date: date })}</Box>
            </Box>
          </Box>
        )}
      </Box>
    );
  }
  return (
    <ListItem component="div" {...others}>
      <Grid container spacing={2}>
        <Grid item>
          {isAppView ? (
            <ChannelAvatar color={channel.color} size={theme.spacing(1)} />
          ) : (
            <ChannelAvatar color={channel.color}>{name[0]}</ChannelAvatar>
          )}
        </Grid>
        <Grid item>
          <ListItemText
            primary={
              <Box display="flex" alignItems="center">
                <Box pl={1} display="inline-block">
                  {name}
                </Box>
              </Box>
            }
            secondary={getSecondaryText()}
            className={classes.root}
            disableTypography
          />
        </Grid>
      </Grid>
      {props.handleUpdateChannel && (
        <ListItemSecondaryAction>
          <MoreMenu
            options={[
              { label: t('frequent|Edit'), action: updateChannel },
              { label: t('frequent|Delete'), action: deleteChannel },
            ]}
          />
        </ListItemSecondaryAction>
      )}
    </ListItem>
  );
}

export default Item;
