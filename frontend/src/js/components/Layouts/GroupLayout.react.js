import { applicationsStore } from "../../stores/Stores"
import React from "react"
import _ from "underscore"
import GroupExtended from "../Groups/ItemExtended.react"
import SectionHeader from '../Common/SectionHeader';
import EditDialog from "../Groups/EditDialog";

class GroupLayout extends React.Component {

 constructor(props) {
    super(props);
    this.onChange = this.onChange.bind(this);
    this.openUpdateGroupModal = this.openUpdateGroupModal.bind(this);
    this.closeUpdateGroupModal = this.closeUpdateGroupModal.bind(this)
    
    let appID = props.match.params.appID,
        groupID = props.match.params.groupID
    this.state = {
      appID: appID,
      groupID: groupID,
      applications: applicationsStore.getCachedApplications(),
      updateGroupModalVisible: false,
    }
  }

  componentWillMount() {
    applicationsStore.getApplication(this.props.match.params.appID)
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
  
  openUpdateGroupModal() {
    this.setState({updateGroupModalVisible: true})
  }

  closeUpdateGroupModal() {
    this.setState({updateGroupModalVisible: false})
  }

  render() {
    let appName = "",
        groupName = ""

    let applications = this.state.applications ? this.state.applications : [],
        application = _.findWhere(applications, {id: this.state.appID})
    let groups = [];
    let channels = [];
    
    if (application) {
      appName = application.name
      groups = application.groups; 
      channels = application.channels ? application.channels : [];
      let group = _.findWhere(application.groups, {id: this.state.groupID})
      if (group) {
        groupName = group.name
      }
    }
   
    const groupToUpdate = _.findWhere(groups, {id: this.state.groupID});
    
    return(
      <div>
        <SectionHeader
          title={groupName}
          breadcrumbs={[
            {
              path: '/apps',
              label: 'Applications'
            },
            {
              path: `/apps/${this.state.appID}`,
              label: appName
            }
          ]}
        />
        <GroupExtended appID={this.state.appID} groupID={this.state.groupID} handleUpdateGroup={this.openUpdateGroupModal}/>
        <EditDialog data={{group: groupToUpdate, channels: channels}} show={this.state.updateGroupModalVisible} onHide={this.closeUpdateGroupModal} />
      </div>
    )
  }

}

export default GroupLayout
