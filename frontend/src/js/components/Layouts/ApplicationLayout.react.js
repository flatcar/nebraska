import Grid from '@material-ui/core/Grid';
import React from 'react';
import _ from 'underscore';
import { applicationsStore } from '../../stores/Stores';
import ChannelsList from '../Channels/List.react';
import SectionHeader from '../Common/SectionHeader';
import GroupsList from '../Groups/List.react';
import PackagesList from '../Packages/List.react';

class ApplicationLayout extends React.Component {

  constructor(props) {
    super(props);
    this.onChange = this.onChange.bind(this);

    const appID = props.match.params.appID;
    this.state = {
      appID: appID,
      applications: applicationsStore.getCachedApplications()
    };
  }

  componentWillMount() {
    applicationsStore.getApplication(this.props.match.params.appID);
  }

  componentDidMount() {
    applicationsStore.addChangeListener(this.onChange);
  }

  componentWillUnmount() {
    applicationsStore.removeChangeListener(this.onChange);
  }

  onChange() {
    this.setState({
      applications: applicationsStore.getCachedApplications()
    });
  }

  render() {
    let appName = '';
    const applications = this.state.applications ? this.state.applications : [];
    const application = _.findWhere(applications, {id: this.state.appID});

    if (application) {
      appName = application.name;
    }

    return (
      <div>
        <SectionHeader
          title={appName}
          breadcrumbs={[
            {
              path: '/apps',
              label: 'Applications'
            }
          ]}
        />
        <Grid
          container
          spacing={1}
          justify="space-between"
        >
          <Grid item xs={8}>
            <GroupsList appID={this.state.appID} />
          </Grid>
          <Grid item xs={4}>
            <Grid
              container
              direction="column"
              alignItems="stretch"
              spacing={2}
            >
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
    );
  }

}

export default ApplicationLayout;
