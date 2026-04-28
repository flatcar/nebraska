import { InlineIcon } from '@iconify/react';
import Cancel from '@mui/icons-material/Cancel';
import ViewInArOutlined from '@mui/icons-material/ViewInArOutlined';
import Box from '@mui/material/Box';
import ListItem from '@mui/material/ListItem';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemSecondaryAction from '@mui/material/ListItemSecondaryAction';
import ListItemText from '@mui/material/ListItemText';
import Stack from '@mui/material/Stack';
import { styled } from '@mui/material/styles';
import Typography from '@mui/material/Typography';
import type React from 'react';
import { useTranslation } from 'react-i18next';
import _ from 'underscore';

import { Channel, Package } from '../../api/apiDataTypes';
import { makeLocaleTime } from '../../i18n/dateTime';
import flatcarIcon from '../../icons/flatcar-logo.json';
import { applicationsStore } from '../../stores/Stores';
import { ARCHES, cleanSemverVersion } from '../../utils/helpers';
import ChannelAvatar from '../Channels/ChannelAvatar';
import Label from '../common/Label';
import MoreMenu from '../common/MoreMenu';

const PREFIX = 'Item';

const classes = {
  packageName: `${PREFIX}-packageName`,
  subtitle: `${PREFIX}-subtitle`,
  packageIcon: `${PREFIX}-packageIcon`,
  channelLabel: `${PREFIX}-channelLabel`,
};

const StyledListItem = styled(ListItem)({
  [`& .${classes.packageName}`]: {
    fontSize: '1.1em',
  },
  [`& .${classes.subtitle}`]: {
    fontSize: '.9em',
    textTransform: 'uppercase',
    fontWeight: 300,
    paddingRight: '.05em',
    color: '#595959',
  },
  [`& .${classes.packageIcon}`]: {
    minWidth: '40px',
  },
  [`& .${classes.channelLabel}`]: {
    marginRight: '5px',
  },
});

const containerIcons: {
  [key: string]: { icon: React.ReactNode; name: string };
} = {
  1: { icon: <InlineIcon icon={flatcarIcon} width="35" height="35" />, name: 'Flatcar' },
  other: { icon: <ViewInArOutlined sx={{ fontSize: 35 }} />, name: 'Other' },
};

interface ItemProps {
  packageItem: Package;
  channels: Channel[];
  handleUpdatePackage(pkgId: string): void;
}

function Item(props: ItemProps) {
  const { t } = useTranslation();

  const type = props.packageItem.type || 1;
  const processedChannels = _.where(props.channels, { package_id: props.packageItem.id });
  let blacklistInfo: string | null = null;
  const item = type in containerIcons ? containerIcons[type] : containerIcons.other;

  if (props.packageItem.channels_blacklist) {
    const channelsList = _.map(props.packageItem.channels_blacklist, channel => {
      return _.findWhere(props.channels, { id: channel })?.name;
    });
    blacklistInfo = channelsList.join(' - ');
  }

  function deletePackage() {
    const confirmationText = t('packages|confirm_delete_package');
    if (window.confirm(confirmationText)) {
      applicationsStore().deletePackage(
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
      <Stack direction="column">
        <Box>
          <Typography component="span" className={classes.subtitle}>
            {`${t('packages|version')}:`}
          </Typography>
          &nbsp;
          {`${cleanSemverVersion(props.packageItem.version)} (${ARCHES[props.packageItem.arch]})`}
        </Box>
        {processedChannels.length > 0 && (
          <Box>
            <Typography component="span" className={classes.subtitle}>
              {`${t('packages|channels')}:`}
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
          </Box>
        )}
        <Box>
          <Typography component="span" className={classes.subtitle}>
            {`${t('packages|released')}:`}
          </Typography>
          &nbsp;
          {makeLocaleTime(props.packageItem.created_ts, {
            showTime: false,
            dateFormat: { year: 'numeric', month: 'numeric', day: 'numeric' },
          })}
        </Box>
        {props.packageItem.channels_blacklist && (
          <Box>
            {props.packageItem.channels_blacklist && (
              <Label>
                <Cancel sx={{ fontSize: 10 }} /> {blacklistInfo}
              </Label>
            )}
          </Box>
        )}
        {props.packageItem.is_floor && (
          <Box>
            <Label>
              <ViewInArOutlined sx={{ fontSize: 10 }} />{' '}
              {t('packages|floor_package', { defaultValue: 'Floor Package' })}
            </Label>
          </Box>
        )}
      </Stack>
    );
  }

  return (
    <StyledListItem dense alignItems="flex-start">
      <ListItemIcon className={classes.packageIcon}>{item.icon}</ListItemIcon>
      <ListItemText
        disableTypography
        slotProps={{ primary: { className: classes.packageName } }}
        primary={<Typography>{item.name}</Typography>}
        secondary={makeItemSecondaryInfo()}
      />
      <ListItemSecondaryAction>
        <MoreMenu
          options={[
            { label: t('frequent|edit'), action: updatePackage },
            { label: t('frequent|delete'), action: deletePackage },
          ]}
        />
      </ListItemSecondaryAction>
    </StyledListItem>
  );
}

export default Item;
