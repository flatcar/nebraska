import MuiList from '@material-ui/core/List';
import Paper from '@material-ui/core/Paper';
import React from 'react';
import _ from 'underscore';
import { applicationsStore, modalStore } from '../../stores/Stores';
import Empty from '../Common/EmptyContent';
import ListHeader from '../Common/ListHeader';
import SearchInput from '../Common/ListSearch';
import Loader from '../Common/Loader';
import ModalButton from '../Common/ModalButton.react';
import EditDialog from './EditDialog';
import Item from './Item.react';

class List extends React.Component {

  constructor(props) {
    super(props);
    this.onChange = this.onChange.bind(this);
    this.searchUpdated = this.searchUpdated.bind(this);
    this.openUpdateAppModal = this.openUpdateAppModal.bind(this);
    this.closeUpdateAppModal = this.closeUpdateAppModal.bind(this);

    this.state = {
      applications: applicationsStore.getCachedApplications(),
      searchTerm: '',
      updateAppModalVisible: false,
      updateAppIDModal: null
    };
  }

  closeUpdateAppModal() {
    this.setState({updateAppModalVisible: false});
  }

  openUpdateAppModal(appID) {
    this.setState({updateAppModalVisible: true, updateAppIDModal: appID});
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

  searchUpdated(event) {
    const {name, value} = event.currentTarget;
    this.setState({searchTerm: value.toLowerCase()});
  }

  render() {
    let applications = this.state.applications;
    let entries = '';

    if (this.state.searchTerm) {
      applications = applications.filter(app => app.name.toLowerCase().includes(this.state.searchTerm));
    }

    if (_.isNull(applications)) {
      entries = <Loader />;
    } else {
      if (_.isEmpty(applications)) {
        if (this.state.searchTerm) {
          entries = <Empty>No results found.</Empty>;
        } else {
          entries = <Empty>Ops, it looks like you have not created any application yet..<br/><br/> Now is a great time to create your first one, just click on the plus symbol above.</Empty>;
        }
      } else {
        entries = _.map(applications, (application, i) => {
          return <Item key={application.id} application={application} handleUpdateApplication={this.openUpdateAppModal} />;
        });
      }
    }

    const appToUpdate =  applications && this.state.updateAppIDModal ? _.findWhere(applications, {id: this.state.updateAppIDModal}) : null;

    return (
      <Paper>
        <ListHeader
          title="Applications"
          actions={[
            <ModalButton
              modalToOpen="AddApplicationModal"
              data={{applications: applications}}
            />
          ]}
        />
        <MuiList>
          {entries}
        </MuiList>
        {appToUpdate &&
          <EditDialog
            data={appToUpdate}
            show={this.state.updateAppModalVisible}
            onHide={this.closeUpdateAppModal}
          />
        }
      </Paper>
    );
  }

}

export default List;
