import cancelIcon from '@iconify/icons-mdi/cancel';
import cubeOutline from '@iconify/icons-mdi/cube-outline';
import { InlineIcon } from '@iconify/react';
import Grid from '@mui/material/Grid';
import ListItem from '@mui/material/ListItem';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemSecondaryAction from '@mui/material/ListItemSecondaryAction';
import ListItemText from '@mui/material/ListItemText';
import { styled } from '@mui/material/styles';
import Typography from '@mui/material/Typography';
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
  [key: string]: any;
} = {
  1: { icon: flatcarIcon, name: 'Flatcar' },
  other: { icon: cubeOutline, name: 'Other' },
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
      <Grid container direction="column">
        <Grid>
          <Typography component="span" className={classes.subtitle}>
            {`${t('packages|version')}:`}
          </Typography>
          &nbsp;
          {`${cleanSemverVersion(props.packageItem.version)} (${ARCHES[props.packageItem.arch]})`}
        </Grid>
        {processedChannels.length > 0 && (
          <Grid>
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
          </Grid>
        )}
        <Grid>
          <Typography component="span" className={classes.subtitle}>
            {`${t('packages|released')}:`}
          </Typography>
          &nbsp;
          {makeLocaleTime(props.packageItem.created_ts, {
            showTime: false,
            dateFormat: { year: 'numeric', month: 'numeric', day: 'numeric' },
          })}
        </Grid>
        {props.packageItem.channels_blacklist && (
          <Grid>
            {props.packageItem.channels_blacklist && (
              <Label>
                <InlineIcon icon={cancelIcon} width="10" height="10" /> {blacklistInfo}
              </Label>
            )}
          </Grid>
        )}
        {props.packageItem.is_floor && (
          <Grid>
            <Label>
              <InlineIcon icon={cubeOutline} width="10" height="10" />{' '}
              {t('packages|floor_package', { defaultValue: 'Floor Package' })}
            </Label>
          </Grid>
        )}
      </Grid>
    );
  }

  return (
    <StyledListItem dense alignItems="flex-start">
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
            { label: t('frequent|edit'), action: updatePackage },
            { label: t('frequent|delete'), action: deletePackage },
          ]}
        />
      </ListItemSecondaryAction>
    </StyledListItem>
  );
}

export default Item;
