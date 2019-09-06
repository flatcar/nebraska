import PropTypes from 'prop-types';
import React from "react"
import ApplicationEditDialog from "../Applications/EditDialog"
import GroupEditDialog from "../Groups/EditDialog"
import ChannelEditDialog from "../Channels/EditDialog"
import PackageEditDialog from '../Packages/EditDialog';
import AddIcon from '@material-ui/icons/Add';
import Fab from '@material-ui/core/Fab';

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

    // @todo: verify whether aria-label should be more specific (in which
    // case it should be set from the caller).
    return(
      <div>
        <Fab size="small" aria-label="add" onClick={this.open.bind()}>
          <AddIcon />
        </Fab>
        {modal}
      </div>
    )
  }

}

ModalButton.propTypes = {
  icon: PropTypes.string.isRequired,
  modalToOpen: PropTypes.string.isRequired,
  data: PropTypes.object
}

export default ModalButton
