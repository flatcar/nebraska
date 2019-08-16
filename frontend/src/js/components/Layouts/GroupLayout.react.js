import { applicationsStore } from "../../stores/Stores"
import React from "react"
import { Row, Col } from "react-bootstrap"
import _ from "underscore"
import GroupExtended from "../Groups/ItemExtended.react"
import SectionHeader from '../Common/SectionHeader';

class GroupLayout extends React.Component {

 constructor(props) {
    super(props);
    this.onChange = this.onChange.bind(this);

    let appID = props.match.params.appID,
        groupID = props.match.params.groupID
    this.state = {
      appID: appID,
      groupID: groupID,
      applications: applicationsStore.getCachedApplications()
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

  render() {
    let appName = "",
        groupName = ""

    let applications = this.state.applications ? this.state.applications : [],
        application = _.findWhere(applications, {id: this.state.appID})

    if (application) {
      appName = application.name

      let group = _.findWhere(application.groups, {id: this.state.groupID})
      if (group) {
        groupName = group.name
      }
    }

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
        <GroupExtended appID={this.state.appID} groupID={this.state.groupID} />
     </div>
    )
  }

}

export default GroupLayout
