import { Box, Grid, makeStyles, Tooltip, useTheme } from '@material-ui/core';
import ListItem from '@material-ui/core/ListItem';
import ListItemSecondaryAction from '@material-ui/core/ListItemSecondaryAction';
import ListItemText from '@material-ui/core/ListItemText';
import ScheduleIcon from '@material-ui/icons/Schedule';
import { useTranslation } from 'react-i18next';
import { Channel } from '../../api/apiDataTypes';
import { applicationsStore } from '../../stores/Stores';
import { ARCHES, cleanSemverVersion } from '../../utils/helpers';
import MoreMenu from '../common/MoreMenu';
import ChannelAvatar from './ChannelAvatar';

const useStyles = makeStyles({
  root: {
    margin: '0px',
  },
});

export interface ChannelItemProps {
  channel: Channel;
  /** The default is to display the arch. */
  showArch?: boolean;
  /** The default is to not display the arch. */
  isAppView?: boolean;
  /** When an update to the channel happens. */
  onChannelUpdate?: (channelID: string) => void;
}

export default function ChannelItem(props: ChannelItemProps) {
  const theme = useTheme();
  const classes = useStyles();
  const { t } = useTranslation();
  const { channel, showArch = true, isAppView = false, onChannelUpdate = null, ...others } = props;
  const name = channel.name;
  const version = channel.package
    ? cleanSemverVersion(channel.package.version)
    : t('channels|No package');

  function deleteChannel() {
    const confirmationText = t('channels|Are you sure you want to delete this channel?');
    if (window.confirm(confirmationText)) {
      applicationsStore().deleteChannel(channel.application_id, channel.id);
    }
  }

  function updateChannel() {
    if (onChannelUpdate) {
      onChannelUpdate(channel.id);
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
      {onChannelUpdate && (
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
