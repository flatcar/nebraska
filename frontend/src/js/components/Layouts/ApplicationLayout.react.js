import { applicationsStore } from "../../stores/Stores"
import React from "react"
import _ from "underscore"
import { Link as RouterLink } from "react-router-dom"
import Grid from '@material-ui/core/Grid';
import ApplicationsList from "../Applications/List.react"
import Breadcrumbs from '@material-ui/core/Breadcrumbs';
import GroupsList from "../Groups/List.react"
import ChannelsList from "../Channels/List.react"
import Link from '@material-ui/core/Link';
import PackagesList from "../Packages/List.react"
import Typography from '@material-ui/core/Typography';

class ApplicationLayout extends React.Component {

  constructor(props) {
    super(props);
    this.onChange = this.onChange.bind(this);

    let appID = props.match.params.appID
    this.state = {
      appID: appID,
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
    let appName = ""
    let applications = this.state.applications ? this.state.applications : [],
        application = _.findWhere(applications, {id: this.state.appID})

    if (application) {
      appName = application.name
    }

    return(
      <div className="container">
        <Breadcrumbs aria-label="breadcrumb">
          <Link variant="h4" to="/apps" component={RouterLink}>Applications</Link>
          <Typography variant="h4">{appName}</Typography>
        </Breadcrumbs>
        <Grid
          container
          spacing={1}
          justify="space-between">
          <Grid item xs={8}>
            <GroupsList appID={this.state.appID} />
          </Grid>
          <Grid item xs={4}>
            <Grid
              container
              direction="column"
              alignItems="stretch"
              spacing={2}>
              <Grid item xs={12}>
                <ChannelsList appID={this.state.appID} />
              </Grid>
              <Grid item xs={12}>
                <PackagesList appID={this.state.appID} />
              </Grid>
            </Grid>
          </Grid>
        </Grid>
      </div>
    )
  }

}

export default ApplicationLayout
