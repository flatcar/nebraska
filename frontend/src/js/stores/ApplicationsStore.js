import _ from 'underscore';
import API from '../api/API';
import Store from './BaseStore';

class ApplicationsStore extends Store {

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

  getCachedApplication(applicationID) {
    return _.findWhere(this.applications, {id: applicationID});
  }

  getApplications() {
    API.getApplications()
      .then(applications => {
        this.applications = applications;
        this.emitChange();
      })
      .catch((error) => {
        if (error.status === 404) {
          this.applications = [];
          this.emitChange();
        }
      });
  }

  getApplication(applicationID) {
    API.getApplication(applicationID)
      .then(application => {
        if (this.applications) {
          const applicationItem = application;
          const index = this.applications ?
            _.findIndex(this.applications, {id: applicationID}) : null;
          if (index >= 0) {
            this.applications[index] = applicationItem;
          } else {
            this.applications.unshift(applicationItem);
          }
          this.emitChange();
        }
      });
  }

  getCachedPackages(applicationID) {
    const app = _.findWhere(this.applications, {id: applicationID});
    const packages = app ? app.packages : [];
    return packages;
  }

  getCachedChannels(applicationID) {
    const app = _.findWhere(this.applications, {id: applicationID});
    const channels = app ? app.channels : [];
    return channels;
  }

  createApplication(data, clonedApplication) {
    return API.createApplication(data, clonedApplication)
      .then(application => {
        const applicationItem = application;
        this.applications.unshift(applicationItem);
        this.emitChange();
      });
  }

  updateApplication(applicationID, data) {
    data.id = applicationID;

    return API.updateApplication(data)
      .then(application => {
        const applicationItem = application;
        const applicationToUpdate = _.findWhere(this.applications, {id: applicationItem.id});

        applicationToUpdate.name = applicationItem.name;
        applicationToUpdate.description = applicationItem.description;
        this.emitChange();
      });
  }

  getAndUpdateApplication(applicationID) {
    API.getApplication(applicationID)
      .then(application => {
        const applicationItem = application;
        const index = _.findIndex(this.applications, {id: applicationID});
        this.applications[index] = applicationItem;
        this.emitChange();
      });
  }

  deleteApplication(applicationID) {
    API.deleteApplication(applicationID).
      then(() => {
        this.applications = _.without(this.applications,
          _.findWhere(this.applications, {id: applicationID}));
        this.emitChange();
      });
  }

  // Groups

  createGroup(data) {
    return API.createGroup(data)
      .then(group => {
        const groupItem = group;
        const applicationToUpdate = _.findWhere(this.applications, {id: groupItem.application_id});
        if (applicationToUpdate.groups) {
          applicationToUpdate.groups.unshift(groupItem);
        } else {
          applicationToUpdate.groups = [groupItem];
        }
        this.emitChange();
      });
  }

  deleteGroup(applicationID, groupID) {
    API.deleteGroup(applicationID, groupID)
      .then(() => {
        const applicationToUpdate = _.findWhere(this.applications, {id: applicationID});
        const newGroups = _.without(applicationToUpdate.groups,
          _.findWhere(applicationToUpdate.groups, {id: groupID}));
        applicationToUpdate.groups = newGroups;
        this.emitChange();
      });
  }

  updateGroup(data) {
    return API.updateGroup(data)
      .then(group => {
        const groupItem = group;
        const applicationToUpdate = _.findWhere(this.applications, {id: groupItem.application_id});
        const index = _.findIndex(applicationToUpdate.groups, {id: groupItem.id});
        applicationToUpdate.groups[index] = groupItem;
        this.emitChange();
      });
  }

  getGroup(applicationID, groupID) {
    API.getGroup(applicationID, groupID)
      .then(group => {
        const groupItem = group;
        const applicationToUpdate = _.findWhere(this.applications, {id: groupItem.application_id});
        const index = _.findIndex(applicationToUpdate.groups, {id: groupItem.id});

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

  getGroupVersionCountTimeline(applicationID, groupID) {
    // Not much value here, but we still abstract the API call in case we
    // want to e.g. cache the results in the future.
    return API.getGroupVersionCountTimeline(applicationID, groupID);
  }

  getGroupStatusCountTimeline(applicationID, groupID) {
    // Not much value here, but we still abstract the API call in case we
    // want to e.g. cache the results in the future.
    return API.getGroupStatusCountTimeline(applicationID, groupID);
  }

  // Channels

  createChannel(data) {
    return API.createChannel(data)
      .then(channel => {
        const channelItem = channel;
        this.getAndUpdateApplication(channelItem.application_id);
      });
  }

  deleteChannel(applicationID, channelID) {
    API.deleteChannel(applicationID, channelID)
      .then(() => {
        this.getAndUpdateApplication(applicationID);
      });
  }

  updateChannel(data) {
    return API.updateChannel(data)
      .then(channel => {
        const channelItem = channel;
        this.getAndUpdateApplication(channelItem.application_id);
      });
  }

  // Packages

  createPackage(data) {
    return API.createPackage(data)
      .then(packageItem => {
        const newpackage = packageItem;
        this.getAndUpdateApplication(newpackage.application_id);
      });
  }

  deletePackage(applicationID, packageID) {
    API.deletePackage(applicationID, packageID)
      .then(() => {
        this.getAndUpdateApplication(applicationID);
      });
  }

  updatePackage(data) {
    return API.updatePackage(data)
      .then(packageItem => {
        const updatedpackage = packageItem;
        this.getAndUpdateApplication(updatedpackage.application_id);
      });
  }

}

export default ApplicationsStore;
