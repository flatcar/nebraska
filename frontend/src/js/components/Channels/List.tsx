import { Box, makeStyles } from '@material-ui/core';
import Grid from '@material-ui/core/Grid';
import MuiList from '@material-ui/core/List';
import ListSubheader from '@material-ui/core/ListSubheader';
import Typography from '@material-ui/core/Typography';
import React from 'react';
import _ from 'underscore';
import API from '../../api/API';
import { Application, Channel, Package } from '../../api/apiDataTypes';
import { applicationsStore } from '../../stores/Stores';
import { ARCHES } from '../../utils/helpers';
import Loader from '../Common/Loader';
import ModalButton from '../Common/ModalButton';
import SectionPaper from '../Common/SectionPaper';
import EditDialog from './EditDialog';
import Item from './Item';

const useStyles = makeStyles({
  root: {
    '& > hr:first-child': {
      display: 'none',
    },
  },
});

interface PackageChannelApplication extends Application {
  packages: Package[];
  channels: Channel[];
}

function ChannelList(props: {
  application: PackageChannelApplication;
  onEdit: (channelId: string) => void;
}) {
  const { application, onEdit } = props;
  const classes = useStyles();

  function getChannelsPerArch() {
    const perArch: {
      [key: number]: any[];
    } = {};

    // If application doesn't have any channel return empty object.
    if (application.channels === null) {
      return perArch;
    }

    application.channels.forEach((channel: Channel) => {
      if (!perArch[channel.arch]) {
        perArch[channel.arch] = [];
      }
      perArch[channel.arch].push(channel);
    });

    return perArch;
  }

  return (
    <React.Fragment>
      {Object.entries(getChannelsPerArch()).map(([arch, channels]) => (
        <MuiList
          key={arch}
          subheader={<ListSubheader disableSticky>{ARCHES[parseInt(arch)]}</ListSubheader>}
          dense
          className={classes.root}
        >
          {channels.map(channel => (
            <Item
              key={'channelID_' + channel.id}
              channel={channel}
              packages={application.packages || []}
              showArch={false}
              handleUpdateChannel={onEdit}
            />
          ))}
        </MuiList>
      ))}
    </React.Fragment>
  );
}

function List(props: { appID: string }) {
  const { appID } = props;
  const [application, setApplication] = React.useState(
    applicationsStore.getCachedApplication(appID)
  );
  const [packages, setPackages] = React.useState<null | Package[]>(null);
  const [channelToEdit, setChannelToEdit] = React.useState<null | Channel>(null);

  React.useEffect(() => {
    applicationsStore.addChangeListener(onStoreChange);

    // In case the application was not yet cached, we fetch it here
    if (application === null) {
      applicationsStore.getApplication(props.appID);
    }

    // Fetch packages
    if (!packages) {
      API.getPackages(props.appID).then(result => {
        if (_.isNull(result)) {
          setPackages([]);
          return;
        }
        setPackages(result);
      });
    }

    return function cleanup() {
      applicationsStore.removeChangeListener(onStoreChange);
    };
  }, [appID]);

  function onStoreChange() {
    setApplication(applicationsStore.getCachedApplication(appID));
  }

  function onChannelEditOpen(channelID: string) {
    let channels = [];
    if (application) {
      channels = application.channels ? application.channels : [];
    }

    const channelToUpdate =
      !_.isEmpty(channels) && channelID ? _.findWhere(channels, { id: channelID }) : null;

    setChannelToEdit(channelToUpdate);
  }

  function onChannelEditClose() {
    setChannelToEdit(null);
  }

  return (
    <Box mt={2}>
      <Box mb={2}>
        <Grid container alignItems="center" justify="space-between">
          <Grid item>
            <Typography variant="h4">Channels</Typography>
          </Grid>
          <Grid item>
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
        {!application ? (
          <Loader />
        ) : (
          <ChannelList application={application} onEdit={onChannelEditOpen} />
        )}
        {channelToEdit && (
          <EditDialog
            data={{ packages: packages, applicationID: appID, channel: channelToEdit }}
            show={channelToEdit !== null}
            onHide={onChannelEditClose}
          />
        )}
      </SectionPaper>
    </Box>
  );
}

export default List;
