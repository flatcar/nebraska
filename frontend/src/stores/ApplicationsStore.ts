import _ from 'underscore';

import API from '../api/API';
import { Application, Channel, Group, Package } from '../api/apiDataTypes';
import Store from './BaseStore';
import { setUser } from './redux/features/user';
import store from './redux/store';

class ApplicationsStore extends Store {
  applications: Application[];
  interval: null | number;
  constructor() {
    super();
    this.applications = [];
    this.interval = null;
  }

  // Applications

  getCachedApplications() {
    return this.applications;
  }

  getCachedApplication(applicationID: string) {
    const app = _.findWhere(this.applications, { id: applicationID });
    if (!app) {
      return null;
    }
    return { ...app };
  }

  getApplications() {
    // Start the refresh interval on first call if not already running
    if (!this.interval) {
      this.interval = window.setInterval(() => {
        this.getApplications();
      }, 60 * 1000);
    }

    API.getApplications()
      .then(response => {
        this.applications = response.applications;
        this.emitChange();
      })
      .catch(error => {
        switch (error.status) {
          case 404:
            this.applications = [];
            this.emitChange();
            break;
          case 401:
            store.dispatch(setUser({ authenticated: false }));
            break;
          default:
            console.debug('Error', error);
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
    const app = _.findWhere(this.applications as _.Collection<any>, { id: applicationID });
    const packages = app ? app.packages : [];
    return packages;
  }

  getCachedChannels(applicationID: string) {
    const app = _.findWhere(this.applications as _.Collection<any>, { id: applicationID });
    const channels = app ? app.channels : [];
    return channels;
  }

  async createApplication(
    data: Pick<Application, 'name' | 'description' | 'product_id'>,
    clonedApplication: string
  ) {
    const application = await API.createApplication(data, clonedApplication);
    const applicationItem = application;
    if (this.applications) {
      this.applications.unshift(applicationItem);
      this.applications = [...this.applications];
      this.emitChange();
    }
  }

  async updateApplication(applicationID: string, data: any) {
    data.id = applicationID;

    const application = await API.updateApplication(data);
    const applicationToUpdate = _.findWhere(this.applications as _.Collection<any>, {
      id: application.id,
    });
    applicationToUpdate.name = application.name;
    applicationToUpdate.description = application.description;
    applicationToUpdate.product_id = application.product_id;
    this.emitChange();
  }

  getAndUpdateApplication(applicationID: string) {
    API.getApplication(applicationID).then(application => {
      const applicationItem = application;
      const index = _.findIndex(this.applications as _.List<any>, { id: applicationID });
      if (this.applications) {
        this.applications[index] = applicationItem;
        this.emitChange();
      }
    });
  }

  deleteApplication(applicationID: string) {
    API.deleteApplication(applicationID).then(() => {
      this.applications = _.without(
        this.applications as _.List<any>,
        _.findWhere(this.applications as _.Collection<any>, { id: applicationID })
      );
      this.emitChange();
    });
  }

  // Groups

  async createGroup(data: Group) {
    const group = await API.createGroup(data);
    const groupItem = group;
    const applicationToUpdate = _.findWhere(this.applications as _.Collection<any>, {
      id: groupItem.application_id,
    });
    if (applicationToUpdate.groups) {
      applicationToUpdate.groups.unshift(groupItem);
    } else {
      applicationToUpdate.groups = [groupItem];
    }
    this.emitChange();
  }

  deleteGroup(applicationID: string, groupID: string) {
    API.deleteGroup(applicationID, groupID).then(() => {
      const applicationToUpdate = _.findWhere(this.applications as _.Collection<any>, {
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

  async updateGroup(data: Group) {
    const group = await API.updateGroup(data);
    const groupItem = group;
    const applicationToUpdate = _.findWhere(this.applications as _.Collection<any>, {
      id: groupItem.application_id,
    });
    const index = _.findIndex(applicationToUpdate.groups, { id: groupItem.id });
    applicationToUpdate.groups[index] = groupItem;
    this.emitChange();
  }

  getGroup(applicationID: string, groupID: string) {
    API.getGroup(applicationID, groupID).then(group => {
      const applicationToUpdate = _.findWhere(this.applications as _.Collection<any>, {
        id: group.application_id,
      });
      const index = _.findIndex(applicationToUpdate.groups, { id: group.id });

      if (applicationToUpdate) {
        if (applicationToUpdate.groups) {
          if (index >= 0) {
            applicationToUpdate.groups[index] = group;
          } else {
            applicationToUpdate.groups.unshift(group);
          }
        } else {
          applicationToUpdate.groups = [group];
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

  async createChannel(data: Channel): Promise<void> {
    const channel = await API.createChannel(data);
    const channelItem = channel;
    this.getAndUpdateApplication(channelItem.application_id);
  }

  async deleteChannel(applicationID: string, channelID: string): Promise<void> {
    await API.deleteChannel(applicationID, channelID);
    this.getAndUpdateApplication(applicationID);
  }

  async updateChannel(data: Channel): Promise<void> {
    const channel = await API.updateChannel(data);
    const channelItem = channel;
    this.getAndUpdateApplication(channelItem.application_id);
  }

  // Packages

  async createPackage(data: Partial<Package>): Promise<void> {
    const packageItem = await API.createPackage(data);
    const newpackage = packageItem;
    this.getAndUpdateApplication(newpackage.application_id);
  }

  async deletePackage(applicationID: string, packageID: string): Promise<void> {
    await API.deletePackage(applicationID, packageID);
    this.getAndUpdateApplication(applicationID);
  }

  async updatePackage(data: Partial<Package>): Promise<void> {
    const packageItem = await API.updatePackage(data);
    const updatedpackage = packageItem;
    this.getAndUpdateApplication(updatedpackage.application_id);
  }
}

export default ApplicationsStore;
