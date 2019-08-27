import PropTypes from 'prop-types';
import React from "react"
import ApplicationEditDialog from "../Applications/EditDialog"
import GroupEditDialog from "../Groups/EditDialog"
import ChannelEditDialog from "../Channels/EditDialog"
import PackageEditDialog from '../Packages/EditDialog';

class ModalButton extends React.Component {

  constructor(props) {
    super(props)
    this.close = this.close.bind(this)
    this.open = this.open.bind(this)

    this.state = {showModal: false}
  }

  close() {
    this.setState({showModal: false})
  }

  open() {
    this.setState({showModal: true})
  }

  render() {
    var options = {
      show: this.state.showModal,
      data: this.props.data
    }

    switch (this.props.modalToOpen) {
      case "AddApplicationModal":
        var modal = <ApplicationEditDialog create {...options} onHide={this.close} />
        break
      case "AddGroupModal":
        var modal = <GroupEditDialog create {...options} onHide={this.close} />
        break
      case "AddChannelModal":
        var modal = <ChannelEditDialog create {...options} onHide={this.close} />
        break
      case "AddPackageModal":
        var modal = <PackageEditDialog create {...options} onHide={this.close} />
        break
    }

    return(
      <a className={"cr-button displayInline fa fa-" + this.props.icon} href="javascript:void(0)" onClick={this.open.bind()} id={"openModal-" + this.props.modalToOpen}>
        {modal}
      </a>
    )
  }

}

ModalButton.propTypes = {
  icon: PropTypes.string.isRequired,
  modalToOpen: PropTypes.string.isRequired,
  data: PropTypes.object
}

export default ModalButton
