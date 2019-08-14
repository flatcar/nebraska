import PropTypes from 'prop-types';
import { applicationsStore } from "../../stores/Stores"
import React from "react"
import { Row, Col } from "react-bootstrap"
import { Link } from "react-router-dom"
import _ from "underscore"
import Item from "./Item.react"
import ModalButton from "../Common/ModalButton.react"
import SearchInput from "../Common/ListSearch"
import Loader from "react-spinners/ScaleLoader"
import MiniLoader from "react-spinners/PulseLoader"
import ModalUpdate from "./ModalUpdate.react"

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
      <div>
        <Col xs={8}>
          <Row>
            <Col xs={5}>
              <h1 className="displayInline mainTitle">Groups</h1>
              <ModalButton icon="plus" modalToOpen="AddGroupModal" data={{channels: channels, appID: this.props.appID}} />
            </Col>
            <Col xs={7} className="alignRight">
              <SearchInput onChange={this.searchUpdated} placeholder="Search..." />
            </Col>
          </Row>
          <Row>
            <Col xs={12} className="groups--container">
              {entries}
            </Col>
          </Row>
        </Col>
        {/* Update group modal */}
        {groupToUpdate &&
          <ModalUpdate
            data={{group: groupToUpdate, channels: channels}}
            modalVisible={this.state.updateGroupModalVisible}
            onHide={this.closeUpdateGroupModal} />
        }
      </div>
		)
  }

}

List.propTypes = {
  appID: PropTypes.string.isRequired
}

export default List
