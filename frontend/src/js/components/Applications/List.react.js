import { applicationsStore, modalStore } from "../../stores/Stores"
import React from "react"
import Grid from '@material-ui/core/Grid';
import ModalButton from "../Common/ModalButton.react"
import Item from "./Item.react"
import _ from "underscore"
import Loader from "react-spinners/ScaleLoader"
import SearchInput from '../Common/ListSearch'
import EditDialog from "./EditDialog"
import Typography from '@material-ui/core/Typography';

class List extends React.Component {

  constructor(props) {
    super(props)
    this.onChange = this.onChange.bind(this)
    this.searchUpdated = this.searchUpdated.bind(this)
    this.openUpdateAppModal = this.openUpdateAppModal.bind(this)
    this.closeUpdateAppModal = this.closeUpdateAppModal.bind(this)

    this.state = {
      applications: applicationsStore.getCachedApplications(),
      searchTerm: "",
      updateAppModalVisible: false,
      updateAppIDModal: null
    }
  }

  closeUpdateAppModal() {
    this.setState({updateAppModalVisible: false})
  }

  openUpdateAppModal(appID) {
    this.setState({updateAppModalVisible: true, updateAppIDModal: appID})
  }

  componentDidMount() {
    applicationsStore.addChangeListener(this.onChange)
  }

  componentWillUnmount() {
    applicationsStore.removeChangeListener(this.onChange)
  }

  onChange() {
    this.setState({
      applications: applicationsStore.getCachedApplications()
    })
  }

  searchUpdated(event) {
    const {name, value} = event.currentTarget;
    this.setState({searchTerm: value.toLowerCase()})
  }

  render() {
    let applications = this.state.applications,
        entries = ""

    if (this.state.searchTerm) {
      applications = applications.filter(app => app.name.toLowerCase().includes(this.state.searchTerm));
    }

    if (_.isNull(applications)) {
      entries = <div className="icon-loading-container"><Loader color="#00AEEF" size="35px" margin="2px"/></div>
    } else {
      if (_.isEmpty(applications)) {
        if (this.state.searchTerm) {
          entries = <div className="emptyBox">No results found.</div>
        } else {
          entries = <div className="emptyBox">Ops, it looks like you have not created any application yet..<br/><br/> Now is a great time to create your first one, just click on the plus symbol above.</div>
        }
      } else {
        entries = _.map(applications, (application, i) => {
          return <Grid item><Item key={application.id} application={application} handleUpdateApplication={this.openUpdateAppModal} /></Grid>
        })
      }
    }

    const appToUpdate =  applications && this.state.updateAppIDModal ? _.findWhere(applications, {id: this.state.updateAppIDModal}) : null

    return(
      <Grid container alignItems="stretch">
        <Grid item xs={8}>
          <ModalButton icon="plus" modalToOpen="AddApplicationModal" data={{applications: applications}} />
        </Grid>
        <Grid item xs={4}>
          <SearchInput
            onChange={this.searchUpdated} placeholder="Search..."
            fullWidth
          />
        </Grid>
        <Grid
          container
          alignItems="stretch"
          direction="column"
          spacing={2}
          className="apps--container">
          {entries}
        </Grid>
        {/* Update app modal */}
        {appToUpdate &&
          <EditDialog
            data={appToUpdate}
            show={this.state.updateAppModalVisible}
            onHide={this.closeUpdateAppModal} />
        }
      </Grid>
    )
  }

}

export default List
