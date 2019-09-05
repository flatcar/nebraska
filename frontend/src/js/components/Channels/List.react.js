import PropTypes from 'prop-types';
import { applicationsStore } from "../../stores/Stores"
import React from "react"
import _ from "underscore"
import ModalButton from "../Common/ModalButton.react"
import EditDialog from "./EditDialog"
import Item from "./Item.react"
import Loader from '../Common/Loader';
import Grid from '@material-ui/core/Grid';
import Typography from '@material-ui/core/Typography';
import MuiList from '@material-ui/core/List';
import Empty from '../Common/EmptyContent';
import SectionPaper from '../Common/SectionPaper';

class List extends React.Component {

  constructor(props) {
    super(props)
    this.onChange = this.onChange.bind(this)
    this.closeUpdateChannelModal = this.closeUpdateChannelModal.bind(this)
    this.openUpdateChannelModal = this.openUpdateChannelModal.bind(this)

    this.state = {
      application: applicationsStore.getCachedApplication(props.appID),
      updateChannelModalVisible: false,
      updateChannelIDModal: null
    }
  }

  closeUpdateChannelModal() {
    this.setState({updateChannelModalVisible: false})
  }

  openUpdateChannelModal(channelID) {
    this.setState({updateChannelModalVisible: true, updateChannelIDModal: channelID})
  }

  componentDidMount() {
    applicationsStore.addChangeListener(this.onChange)
  }

  componentWillUnmount() {
    applicationsStore.removeChangeListener(this.onChange)
  }

  onChange() {
    this.setState({
      application: applicationsStore.getCachedApplication(this.props.appID)
    })
  }

  render() {
    let application = this.state.application,
        channels = [],
        packages = [],
        entries = ""

    if (application) {
      channels = application.channels ? application.channels : []
      packages = application.packages ? application.packages : []

      if (_.isEmpty(channels)) {
        entries = <Empty>This application does not have any channel yet</Empty>;
      } else {
        entries = _.map(channels, (channel, i) => {
          return <Item key={"channelID_" + channel.id} channel={channel} packages={packages} handleUpdateChannel={this.openUpdateChannelModal} />
        })
      }
    } else {
      entries = <Loader />
    }

    const channelToUpdate =  !_.isEmpty(channels) && this.state.updateChannelIDModal ? _.findWhere(channels, {id: this.state.updateChannelIDModal}) : null

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
              icon="plus"
              modalToOpen="AddChannelModal"
              data={{
                packages: packages,
                applicationID: this.props.appID
              }}
            />
          </Grid>
        </Grid>
        <MuiList dense>
          {entries}
        </MuiList>
        {channelToUpdate &&
          <EditDialog
            data={{packages: packages, applicationID: this.props.appID, channel: channelToUpdate}}
            show={this.state.updateChannelModalVisible}
            onHide={this.closeUpdateChannelModal} />
        }
      </SectionPaper>
    )
  }
}

List.propTypes = {
  appID: PropTypes.string.isRequired
}

export default List
