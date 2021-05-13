import _, { Collection, List } from 'underscore';
import API from '../api/API';
import { Application, Channel, Group, Package } from '../api/apiDataTypes';
import Store from './BaseStore';
import { setUser } from './redux/actions';
import store from './redux/store';

class ApplicationsStore extends Store {
  applications: Application[] | null;
  constructor() {
    super();
    this.applications = null;
    this.getApplications();

    setInterval(() => {
      this.getApplications();
    }, 60 * 1000);
  }

  // Applications

  getCachedApplications() {
    return this.applications;
  }

  getCachedApplication(applicationID: string) {
    const app = _.findWhere(this.applications as Collection<any>, { id: applicationID });
    if (!app) {
      return null;
    }
    return { ...app };
  }

  getApplications() {
    API.getApplications()
      .then(applications => {
        this.applications = applications;
        this.emitChange();
      })
      .catch(error => {
        switch (error.status) {
          case 404:
            this.applications = [];
            this.emitChange();
            break;
          case 401:
            store.dispatch(setUser({authenticated: false}))
            break;
          default:
            console.debug('Error', error)
        }
      });
  }

  getApplication(applicationID: string) {
    API.getApplication(applicationID).then(application => {
      if (this.applications) {
        const applicationItem = application;
        const index = _.findIndex(this.applications, { id: applicationID });
        if (index >= 0) {
          this.applications[index] = applicationItem;
        } else {
          this.applications.unshift(applicationItem);
        }
        this.emitChange();
      }
    });
  }

  getCachedPackages(applicationID: string) {
    const app = _.findWhere(this.applications as Collection<any>, { id: applicationID });
    const packages = app ? app.packages : [];
    return packages;
  }

  getCachedChannels(applicationID: string) {
    const app = _.findWhere(this.applications as Collection<any>, { id: applicationID });
    const channels = app ? app.channels : [];
    return channels;
  }

  createApplication(data: { name: string; description: string }, clonedApplication: string) {
    return API.createApplication(data, clonedApplication).then(application => {
      const applicationItem = application;
      if (this.applications) {
        this.applications.unshift(applicationItem);
        this.applications = [...this.applications];
        this.emitChange();
      }
    });
  }

  updateApplication(applicationID: string, data: any) {
    data.id = applicationID;

    return API.updateApplication(data).then(application => {
      const applicationItem = application;
      const applicationToUpdate = _.findWhere(this.applications as Collection<any>, {
        id: applicationItem.id,
      });

      applicationToUpdate.name = applicationItem.name;
      applicationToUpdate.description = applicationItem.description;
      this.emitChange();
    });
  }

  getAndUpdateApplication(applicationID: string) {
    API.getApplication(applicationID).then(application => {
      const applicationItem = application;
      const index = _.findIndex(this.applications as List<any>, { id: applicationID });
      if (this.applications) {
        this.applications[index] = applicationItem;
        this.emitChange();
      }
    });
  }

  deleteApplication(applicationID: string) {
    API.deleteApplication(applicationID).then(() => {
      this.applications = _.without(
        this.applications as List<any>,
        _.findWhere(this.applications as Collection<any>, { id: applicationID })
      );
      this.emitChange();
    });
  }

  // Groups

  createGroup(data: Group) {
    return API.createGroup(data).then(group => {
      const groupItem = group;
      const applicationToUpdate = _.findWhere(this.applications as Collection<any>, {
        id: groupItem.application_id,
      });
      if (applicationToUpdate.groups) {
        applicationToUpdate.groups.unshift(groupItem);
      } else {
        applicationToUpdate.groups = [groupItem];
      }
      this.emitChange();
    });
  }

  deleteGroup(applicationID: string, groupID: string) {
    API.deleteGroup(applicationID, groupID).then(() => {
      const applicationToUpdate = _.findWhere(this.applications as Collection<any>, {
        id: applicationID,
      });
      const newGroups = _.without(
        applicationToUpdate.groups,
        _.findWhere(applicationToUpdate.groups, { id: groupID })
      );

      applicationToUpdate.groups = [...newGroups];
      this.emitChange();
    });
  }

  updateGroup(data: Group) {
    return API.updateGroup(data).then(group => {
      const groupItem = group;
      const applicationToUpdate = _.findWhere(this.applications as Collection<any>, {
        id: groupItem.application_id,
      });
      const index = _.findIndex(applicationToUpdate.groups, { id: groupItem.id });
      applicationToUpdate.groups[index] = groupItem;
      this.emitChange();
    });
  }

  getGroup(applicationID: string, groupID: string) {
    API.getGroup(applicationID, groupID).then(group => {
      const groupItem = group;
      const applicationToUpdate = _.findWhere(this.applications as Collection<any>, {
        id: groupItem.application_id,
      });
      const index = _.findIndex(applicationToUpdate.groups, { id: groupItem.id });

      if (applicationToUpdate) {
        if (applicationToUpdate.groups) {
          if (index >= 0) {
            applicationToUpdate.groups[index] = groupItem;
          } else {
            applicationToUpdate.groups.unshift(groupItem);
          }
        } else {
          applicationToUpdate.groups = [groupItem];
        }
      }
      this.emitChange();
    });
  }

  getGroupVersionCountTimeline(applicationID: string, groupID: string, duration: string) {
    // Not much value here, but we still abstract the API call in case we
    // want to e.g. cache the results in the future.
    return API.getGroupVersionCountTimeline(applicationID, groupID, duration);
  }

  getGroupStatusCountTimeline(applicationID: string, groupID: string, duration: string) {
    // Not much value here, but we still abstract the API call in case we
    // want to e.g. cache the results in the future.
    return API.getGroupStatusCountTimeline(applicationID, groupID, duration);
  }

  // Channels

  createChannel(data: Channel) {
    return API.createChannel(data).then(channel => {
      const channelItem = channel;
      this.getAndUpdateApplication(channelItem.application_id);
    });
  }

  deleteChannel(applicationID: string, channelID: string) {
    API.deleteChannel(applicationID, channelID).then(() => {
      this.getAndUpdateApplication(applicationID);
    });
  }

  updateChannel(data: Channel) {
    return API.updateChannel(data).then(channel => {
      const channelItem = channel;
      this.getAndUpdateApplication(channelItem.application_id);
    });
  }

  // Packages

  createPackage(data: Partial<Package>) {
    return API.createPackage(data).then(packageItem => {
      const newpackage = packageItem;
      this.getAndUpdateApplication(newpackage.application_id);
    });
  }

  deletePackage(applicationID: string, packageID: string) {
    API.deletePackage(applicationID, packageID).then(() => {
      this.getAndUpdateApplication(applicationID);
    });
  }

  updatePackage(data: Partial<Package>) {
    return API.updatePackage(data).then(packageItem => {
      const updatedpackage = packageItem;
      this.getAndUpdateApplication(updatedpackage.application_id);
    });
  }
}

export default ApplicationsStore;
