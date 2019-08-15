import PropTypes from 'prop-types';
import { applicationsStore } from "../../stores/Stores"
import React from "react"
import GroupsList from "./ApplicationItemGroupsList.react"
import ChannelsList from "./ApplicationItemChannelsList.react"
import Grid from '@material-ui/core/Grid';
import Card from '@material-ui/core/Card';
import CardContent from '@material-ui/core/CardContent';
import {CardFeatureLabel, CardHeader} from '../Common/Card';

class Item extends React.Component {

  constructor(props) {
    super(props)
    this.updateApplication = this.updateApplication.bind(this)
    this.deleteApplication = this.deleteApplication.bind(this)
  }

  updateApplication() {
    this.props.handleUpdateApplication(this.props.application.id)
  }

  deleteApplication() {
    let confirmationText = "Are you sure you want to delete this application?"
    if (confirm(confirmationText)) {
      applicationsStore.deleteApplication(this.props.application.id)
    }
  }

  render() {
    let application = this.props.application ? this.props.application : {},
        description = this.props.application.description ? this.props.application.description : "No description provided",
        styleDescription = this.props.application.description ? "" : " italicText",
        channels = this.props.application.channels ? this.props.application.channels : [],
        groups = this.props.application.groups ? this.props.application.groups : [],
        instances = this.props.application.instances.count ? this.props.application.instances.count : 0,
        appID = this.props.application ? this.props.application.id : "",
        popoverContent = {
          type: "application",
          appID: appID
        }

    return(
      <Card>
        <CardHeader
          cardMainLinkLabel={this.props.application.name}
          cardMainLinkPath={{pathname: `/apps/${appID}`}}
          cardId={appID}
          cardDescription={description}
        >
          <div className="apps--buttons">
            <button className="cr-button displayInline fa fa-edit" onClick={this.updateApplication}></button>
            <button className="cr-button displayInline fa fa-trash-o" onClick={this.deleteApplication}></button>
          </div>
        </CardHeader>
        <CardContent>
          <Grid container spacing={2}>
            <Grid item xs={2}>
              <CardFeatureLabel>Instances:</CardFeatureLabel>
              <span>
                {instances}
              </span>
            </Grid>
            <Grid item xs={10}>
              <CardFeatureLabel>Groups:</CardFeatureLabel>
              <span>
                {groups.length}
              </span>
              <GroupsList
                groups={groups}
                appID={this.props.application.id}
                appName={this.props.application.name} />
            </Grid>
            <Grid item xs={12}>
              <CardFeatureLabel>Channels:</CardFeatureLabel>
              <ChannelsList channels={channels} />
            </Grid>
          </Grid>
        </CardContent>
      </Card>
    )
  }
}

Item.propTypes = {
  application: PropTypes.object.isRequired,
  handleUpdateApplication: PropTypes.func.isRequired
}

export default Item
