import PropTypes from 'prop-types';
import { applicationsStore } from "../../stores/Stores"
import React from "react"
import _ from "underscore"
import ModalButton from "../Common/ModalButton.react"
import Item from "./Item.react"
import ModalUpdate from "./ModalUpdate.react"
import Loader from "react-spinners/ScaleLoader"
import Card from '@material-ui/core/Card';
import CardContent from '@material-ui/core/CardContent';
import Grid from '@material-ui/core/Grid';
import {CardDescriptionLabel} from '../Common/Card';
import Typography from '@material-ui/core/Typography';

class List extends React.Component {

  constructor(props) {
    super(props)
    this.onChange = this.onChange.bind(this)
    this.closeUpdatePackageModal = this.closeUpdatePackageModal.bind(this)
    this.openUpdatePackageModal = this.openUpdatePackageModal.bind(this)

    this.state = {
      application: applicationsStore.getCachedApplication(props.appID),
      updatePackageModalVisible: false,
      updatePackageIDModal: null
    }
  }

  closeUpdatePackageModal() {
    this.setState({updatePackageModalVisible: false})
  }

  openUpdatePackageModal(packageID) {
    this.setState({updatePackageModalVisible: true, updatePackageIDModal: packageID})
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

      if (_.isEmpty(packages)) {
        entries = <CardDescriptionLabel>This application does not have any package yet</CardDescriptionLabel>
      } else {
        entries = _.map(packages, (packageItem, i) => {
          return <Item key={"packageItemID_" + packageItem.id} packageItem={packageItem} channels={channels} handleUpdatePackage={this.openUpdatePackageModal} />
        })
      }
    } else {
      entries = <Loader color="#00AEEF" size="35px" margin="2px"/>
    }

    const packageToUpdate =  !_.isEmpty(packages) && this.state.updatePackageIDModal ? _.findWhere(packages, {id: this.state.updatePackageIDModal}) : null

    return (
      <Grid container spacing={1}>
        <Grid item xs={12}>
          <Typography variant="h4" className="displayInline mainTitle">Packages</Typography>
          <ModalButton icon="plus" modalToOpen="AddPackageModal"
            data={{channels: channels, appID: this.props.appID}} />
        </Grid>
        <Grid item xs={12}>
          <Card>
            <CardContent className="groups--packagesList">
              {entries}
              {/* Update package modal */}
              {packageToUpdate &&
                <ModalUpdate
                  data={{channels: channels, channel: packageToUpdate}}
                  modalVisible={this.state.updatePackageModalVisible}
                  onHide={this.closeUpdatePackageModal} />
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
