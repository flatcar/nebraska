import Grid from '@material-ui/core/Grid';
import MuiList from '@material-ui/core/List';
import ListSubheader from '@material-ui/core/ListSubheader';
import Typography from '@material-ui/core/Typography';
import PropTypes from 'prop-types';
import React from 'react';
import _ from 'underscore';
import { ARCHES } from '../../constants/helpers';
import { applicationsStore } from '../../stores/Stores';
import Loader from '../Common/Loader';
import ModalButton from '../Common/ModalButton.react';
import SectionPaper from '../Common/SectionPaper';
import EditDialog from './EditDialog';
import Item from './Item.react';

function ChannelList(props) {
  const {application, onEdit} = props;

  function getChannelsPerArch() {
    const perArch = {};
    application.channels.forEach(channel => {
      if (!perArch[channel.arch]) {
        perArch[channel.arch] = [];
      }
      perArch[channel.arch].push(channel);
    });

    return perArch;
  }

  return (
    <React.Fragment>
      {Object.entries(getChannelsPerArch()).map(([arch, channels]) =>
        <MuiList
          key={arch}
          subheader={<ListSubheader disableSticky >{ARCHES[arch]}</ListSubheader>}
          dense
        >
          {channels.map(channel =>
            <Item
              key={'channelID_' + channel.id}
              channel={channel}
              packages={application.packages || []}
              showArch={false}
              handleUpdateChannel={onEdit}
            />
          )}
        </MuiList>
      )}
    </React.Fragment>
  );
}

class List extends React.Component {

  constructor(props) {
    super(props);
    this.onChange = this.onChange.bind(this);
    this.closeUpdateChannelModal = this.closeUpdateChannelModal.bind(this);
    this.openUpdateChannelModal = this.openUpdateChannelModal.bind(this);

    this.state = {
      application: applicationsStore.getCachedApplication(props.appID),
      updateChannelModalVisible: false,
      updateChannelIDModal: null
    };
  }

  closeUpdateChannelModal() {
    this.setState({updateChannelModalVisible: false});
  }

  openUpdateChannelModal(channelID) {
    this.setState({updateChannelModalVisible: true, updateChannelIDModal: channelID});
  }

  componentDidMount() {
    applicationsStore.addChangeListener(this.onChange);
  }

  componentWillUnmount() {
    applicationsStore.removeChangeListener(this.onChange);
  }

  onChange() {
    this.setState({
      application: applicationsStore.getCachedApplication(this.props.appID)
    });
  }

  render() {
    const application = this.state.application;
    let channels = [];
    let packages = [];

    if (application) {
      channels = application.channels ? application.channels : [];
      packages = application.packages ? application.packages : [];
    }

    const channelToUpdate =  !_.isEmpty(channels) && this.state.updateChannelIDModal ? _.findWhere(channels, {id: this.state.updateChannelIDModal}) : null;

    return (
      <SectionPaper>
        <Grid
          container
          alignItems="center"
          justify="space-between"
        >
          <Grid item>
            <Typography variant="h5">Channels</Typography>
          </Grid>
          <Grid item>
            <ModalButton
              modalToOpen="AddChannelModal"
              data={{
                packages: packages,
                applicationID: this.props.appID
              }}
            />
          </Grid>
        </Grid>
        {!application ?
          <Loader />
          :
          <ChannelList
            application={application}
            onEdit={this.openUpdateChannelModal}
          />
        }
        {channelToUpdate &&
          <EditDialog
            data={{packages: packages, applicationID: this.props.appID, channel: channelToUpdate}}
            show={this.state.updateChannelModalVisible}
            onHide={this.closeUpdateChannelModal}
          />
        }
      </SectionPaper>
    );
  }
}

List.propTypes = {
  appID: PropTypes.string.isRequired
};

export default List;
