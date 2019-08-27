import PropTypes from 'prop-types';
import { applicationsStore } from "../../stores/Stores"
import React from "react"
import Grid from '@material-ui/core/Grid';
import _ from "underscore"
import Item from "./Item.react"
import ModalButton from "../Common/ModalButton.react"
import SearchInput from "../Common/ListSearch"
import Loader from "react-spinners/ScaleLoader"
import MiniLoader from "react-spinners/PulseLoader"
import Typography from '@material-ui/core/Typography';
import EditDialog from './EditDialog';

class List extends React.Component {

  constructor(props) {
    super(props)
    this.onChange = this.onChange.bind(this)
    this.searchUpdated = this.searchUpdated.bind(this)
    this.openUpdateGroupModal = this.openUpdateGroupModal.bind(this)
    this.closeUpdateGroupModal = this.closeUpdateGroupModal.bind(this)

    this.state = {
      application: applicationsStore.getCachedApplication(this.props.appID),
      searchTerm: "",
      updateGroupModalVisible: false,
      updateGroupIDModal: null,
      updateAppIDModal: null
    }
  }

  closeUpdateGroupModal() {
    this.setState({updateGroupModalVisible: false})
  }

  openUpdateGroupModal(appID, groupID) {
    this.setState({updateGroupModalVisible: true, updateGroupIDModal: groupID, updateAppIDModal: appID})
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

  searchUpdated(event) {
    const {name, value} = event.currentTarget;
    this.setState({searchTerm: value.toLowerCase()})
  }

  render() {
    let application = this.state.application

    let channels = [],
        groups = [],
        packages = [],
        instances = 0,
        name = "",
        entries = ""

    const miniLoader = <div className="icon-loading-container"><MiniLoader color="#00AEEF" size="12px" /></div>

    if (application) {
      name = application.name
      groups = application.groups ? application.groups : []
      packages = application.packages ? application.packages : []
      instances = application.instances ? application.instances : []
      channels = application.channels ? application.channels : []

      if (this.state.searchTerm) {
        groups = groups.filter(app => app.name.toLowerCase().includes(this.state.searchTerm));
      }

      if (_.isEmpty(groups)) {
        if (this.state.searchTerm) {
          entries = <div className="emptyBox">No results found.</div>
        } else {
          entries = <div className="emptyBox">There are no groups for this application yet.<br/><br/>Groups help you control how you want to distribute updates to a specific set of instances.</div>
        }
      } else {
        entries = _.map(groups, (group, i) => {
          return <Item key={"groupID_" + group.id} group={group} appName={name} channels={channels} handleUpdateGroup={this.openUpdateGroupModal} />
        })
      }

    } else {
      entries = <div className="icon-loading-container"><Loader color="#00AEEF" size="35px" margin="2px"/></div>
    }

    const groupToUpdate =  !_.isEmpty(groups) && this.state.updateGroupIDModal ? _.findWhere(groups, {id: this.state.updateGroupIDModal}) : null

		return (
      <Grid container alignItems="stretch">
        <Grid item xs={7}>
          <Typography variant="h4" className="displayInline mainTitle">Groups</Typography>
          <ModalButton icon="plus" modalToOpen="AddGroupModal" data={{channels: channels, appID: this.props.appID}} />
        </Grid>
        <Grid item xs={5}>
          <SearchInput onChange={this.searchUpdated} placeholder="Search..." />
        </Grid>
        <Grid
          container
          alignItems="stretch"
          direction="column"
          className="groups--container">
          {entries}
        </Grid>
        {/* Update group modal */}
        {groupToUpdate &&
          <EditDialog
            data={{group: groupToUpdate, channels: channels}}
            show={this.state.updateGroupModalVisible}
            onHide={this.closeUpdateGroupModal} />
        }
      </Grid>
		)
  }

}

List.propTypes = {
  appID: PropTypes.string.isRequired
}

export default List
