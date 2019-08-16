import PropTypes from 'prop-types';
import { applicationsStore } from "../../stores/Stores"
import React from "react"
import _ from "underscore"
import ModalButton from "../Common/ModalButton.react"
import ModalUpdate from "./ModalUpdate.react"
import Item from "./Item.react"
import Loader from "react-spinners/ScaleLoader"
import Card from '@material-ui/core/Card';
import CardContent from '@material-ui/core/CardContent';
import Grid from '@material-ui/core/Grid';
import Typography from '@material-ui/core/Typography';

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
        entries = <div className="emptyBox">This application does not have any channel yet</div>;
      } else {
        entries = _.map(channels, (channel, i) => {
          return <Item key={"channelID_" + channel.id} channel={channel} packages={packages} handleUpdateChannel={this.openUpdateChannelModal} />
        })
      }
    } else {
      entries = <div className="icon-loading-container"><Loader color="#00AEEF" size="35px" margin="2px"/></div>
    }

    const channelToUpdate =  !_.isEmpty(channels) && this.state.updateChannelIDModal ? _.findWhere(channels, {id: this.state.updateChannelIDModal}) : null

    return (
      <Grid container spacing={1}>
        <Grid item xs={12}>
          <Typography variant="h4" className="displayInline mainTitle">Channels</Typography>
          <ModalButton
              icon="plus"
              modalToOpen="AddChannelModal"
              data={{packages: packages, applicationID: this.props.appID}} />
        </Grid>
        <Grid item xs={12}>
          <Card>
            <CardContent className="groups--packagesList">
              {entries}
              {/* Update channel modal */}
              {channelToUpdate &&
                <ModalUpdate
                  data={{packages: packages, applicationID: this.props.appID, channel: channelToUpdate}}
                  modalVisible={this.state.updateChannelModalVisible}
                  onHide={this.closeUpdateChannelModal} />
              }
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    )
  }
}

List.propTypes = {
  appID: PropTypes.string.isRequired
}

export default List
