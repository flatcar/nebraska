import { Box } from '@mui/material';
import Grid from '@mui/material/Grid';
import MuiList from '@mui/material/List';
import ListSubheader from '@mui/material/ListSubheader';
import Typography from '@mui/material/Typography';
import React from 'react';
import { useTranslation } from 'react-i18next';
import _ from 'underscore';

import API from '../../api/API';
import { Channel, Package } from '../../api/apiDataTypes';
import { applicationsStore } from '../../stores/Stores';
import { ARCHES } from '../../utils/helpers';
import Empty from '../common/EmptyContent';
import Loader from '../common/Loader';
import ModalButton from '../common/ModalButton';
import SectionPaper from '../common/SectionPaper';
import ChannelEdit from './ChannelEdit';
import ChannelItem from './ChannelItem';

function Channels(props: { channels: null | Channel[]; onEdit: (channelId: string) => void }) {
  const { channels, onEdit } = props;
  const { t } = useTranslation();

  const channelsPerArch = (function () {
    const perArch: {
      [key: number]: Channel[];
    } = {};

    (channels ? channels : []).forEach((channel: Channel) => {
      if (!perArch[channel.arch]) {
        perArch[channel.arch] = [];
      }
      perArch[channel.arch].push(channel);
    });

    return perArch;
  })();

  const noChannels = !Object.values(channelsPerArch).find(
    channels => !!channels && channels.length > 0
  );

  if (noChannels) {
    return <Empty>{t('channels|no_channels_created')}</Empty>;
  }

  return (
    <React.Fragment>
      {Object.entries(channelsPerArch).map(([arch, channels]) => (
        <MuiList
          key={arch}
          subheader={<ListSubheader disableSticky>{ARCHES[parseInt(arch)]}</ListSubheader>}
          dense
          sx={{
            '& > hr:first-of-type': {
              display: 'none',
            },
          }}
        >
          {channels.map(channel => (
            <ChannelItem
              key={'channelID_' + channel.id}
              channel={channel}
              showArch={false}
              onChannelUpdate={onEdit}
            />
          ))}
        </MuiList>
      ))}
    </React.Fragment>
  );
}

export interface ChannelListProps {
  appID: string;
}
export default function ChannelList(props: ChannelListProps) {
  const { appID } = props;
  const [application, setApplication] = React.useState(
    applicationsStore().getCachedApplication(appID)
  );
  const [packages, setPackages] = React.useState<null | Package[]>(null);

  function onStoreChange() {
    setApplication(applicationsStore().getCachedApplication(appID));
  }

  React.useEffect(() => {
    applicationsStore().addChangeListener(onStoreChange);

    // In case the application was not yet cached, we fetch it here
    if (application === null) {
      applicationsStore().getApplication(props.appID);
    } else {
      // Fetch packages
      API.getPackages(application.id)
        .then(result => {
          if (_.isNull(result.packages)) {
            setPackages([]);
          } else {
            setPackages(result.packages);
          }
        })
        .catch(err => {
          console.error('Error getting the packages for the channel: ', err);
        });
    }

    return function cleanup() {
      applicationsStore().removeChangeListener(onStoreChange);
    };
  }, [application]);

  const channels = application ? (application.channels ? application.channels : []) : [];
  const loading = !application || packages === null;

  return (
    <ChannelListPure
      channels={channels}
      appID={appID}
      packages={packages ? packages : []}
      loading={loading}
    />
  );
}

export interface ChannelListPureProps {
  /** Application ID for these channels. */
  appID: string;
  /** The Packages to choose from when adding or editing a channel. */
  packages: Package[];
  /** The channels to list. */
  channels: Channel[];
  /** If we are waiting on channels or packages data. */
  loading: boolean;
}

export function ChannelListPure(props: ChannelListPureProps) {
  const [channelToEdit, setChannelToEdit] = React.useState<null | Channel>(null);
  const { t } = useTranslation();
  const { packages, appID, channels, loading } = props;

  function onChannelEditOpen(channelID: string) {
    const channelToUpdate =
      !_.isEmpty(channels) && channelID
        ? _.findWhere(channels ? channels : [], { id: channelID }) || null
        : null;

    setChannelToEdit(channelToUpdate);
  }

  function onChannelEditClose() {
    setChannelToEdit(null);
  }

  return (
    <Box mt={2}>
      <Box mb={2}>
        <Grid container alignItems="center" justifyContent="space-between">
          <Grid>
            <Typography variant="h1">{t('channels|channels')}</Typography>
          </Grid>
          <Grid>
            <ModalButton
              modalToOpen="AddChannelModal"
              data={{
                packages: packages,
                applicationID: appID,
              }}
            />
          </Grid>
        </Grid>
      </Box>
      <SectionPaper>
        {loading ? <Loader /> : <Channels channels={channels} onEdit={onChannelEditOpen} />}
        {channelToEdit && (
          <ChannelEdit
            data={{ packages: packages, applicationID: appID, channel: channelToEdit }}
            show={channelToEdit !== null}
            onHide={onChannelEditClose}
          />
        )}
      </SectionPaper>
    </Box>
  );
}
