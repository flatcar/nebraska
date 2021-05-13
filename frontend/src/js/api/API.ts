import PubSub from 'pubsub-js';
import queryString from 'querystring';
import _ from 'underscore';
import { getToken, setToken } from '../utils/auth';
import { Application, Channel, FlatcarAction, Group, Package } from './apiDataTypes';

const MAIN_PROGRESS_BAR = 'main_progress_bar';
const BASE_URL = '/api';
type REQUEST_DATA_TYPE =
  | string
  | Blob
  | ArrayBufferView
  | ArrayBuffer
  | FormData
  | URLSearchParams
  | ReadableStream<Uint8Array>
  | null
  | undefined;

class API {
  // Applications

  static getApplications() {
    return API.getJSON(BASE_URL + '/apps');
  }

  static getApplication(applicationID: string) {
    return API.getJSON(BASE_URL + '/apps/' + applicationID);
  }

  static deleteApplication(applicationID: string) {
    const url = BASE_URL + '/apps/' + applicationID;

    return API.doRequest('DELETE', url, '');
  }

  static createApplication(
    applicationData: { name: string; description: string },
    clonedFromAppID: string
  ) {
    let url = BASE_URL + '/apps';
    if (clonedFromAppID) {
      url += '?clone_from=' + clonedFromAppID;
    }

    return API.doRequest('POST', url, JSON.stringify(applicationData));
  }

  static updateApplication(applicationData: Application) {
    const url = BASE_URL + '/apps/' + applicationData.id;

    return API.doRequest('PUT', url, JSON.stringify(applicationData));
  }

  // Groups

  static getGroup(applicationID: string, groupID: string) {
    return API.getJSON(BASE_URL + '/apps/' + applicationID + '/groups/' + groupID);
  }

  static deleteGroup(applicationID: string, groupID: string) {
    const url = BASE_URL + '/apps/' + applicationID + '/groups/' + groupID;

    return API.doRequest('DELETE', url, '');
  }

  static createGroup(groupData: Group) {
    const url = BASE_URL + '/apps/' + groupData.application_id + '/groups';

    return API.doRequest('POST', url, JSON.stringify(groupData));
  }

  static updateGroup(groupData: Group) {
    const url = BASE_URL + '/apps/' + groupData.application_id + '/groups/' + groupData.id;

    return API.doRequest('PUT', url, JSON.stringify(groupData));
  }

  static getGroupVersionCountTimeline(applicationID: string, groupID: string, duration: string) {
    return API.getJSON(
      `${BASE_URL}/apps/${applicationID}/groups/${groupID}/version_timeline?duration=${duration}`
    );
  }

  static getGroupStatusCountTimeline(applicationID: string, groupID: string, duration: string) {
    return API.getJSON(
      `${BASE_URL}/apps/${applicationID}/groups/${groupID}/status_timeline?duration=${duration}`
    );
  }

  static getGroupInstancesStats(applicationID: string, groupID: string, duration: string) {
    return API.getJSON(
      `${BASE_URL}/apps/${applicationID}/groups/${groupID}/instances_stats?duration=${duration}`
    );
  }

  static getGroupVersionBreakdown(applicationID: string, groupID: string) {
    return API.getJSON(
      BASE_URL + '/apps/' + applicationID + '/groups/' + groupID + '/version_breakdown'
    );
  }

  // Channels

  static deleteChannel(applicationID: string, channelID: string) {
    const url = BASE_URL + '/apps/' + applicationID + '/channels/' + channelID;

    return API.doRequest('DELETE', url, '');
  }

  static createChannel(channelData: Channel) {
    const url = BASE_URL + '/apps/' + channelData.application_id + '/channels';

    return API.doRequest('POST', url, JSON.stringify(channelData));
  }

  static updateChannel(channelData: Channel) {
    const keysToRemove = ['id', 'created_ts', 'package'];
    const processedChannel = API.removeKeysFromObject(channelData, keysToRemove);
    const url = BASE_URL + '/apps/' + channelData.application_id + '/channels/' + channelData.id;

    return API.doRequest('PUT', url, JSON.stringify(processedChannel));
  }

  // Packages
  static getPackages(applicationID: string) {
    const url = BASE_URL + '/apps/' + applicationID + '/packages/';

    return API.doRequest('GET', url);
  }

  static deletePackage(applicationID: string, packageID: string) {
    const url = BASE_URL + '/apps/' + applicationID + '/packages/' + packageID;

    return API.doRequest('DELETE', url, '');
  }

  static createPackage(packageData: Partial<Package>) {
    const url = BASE_URL + '/apps/' + packageData.application_id + '/packages';

    return API.doRequest('POST', url, JSON.stringify(packageData));
  }

  static updatePackage(packageData: Partial<Package>) {
    const keysToRemove = ['id', 'created_ts', 'package'];
    const processedPackage = API.removeKeysFromObject(packageData, keysToRemove);
    const url = BASE_URL + '/apps/' + packageData.application_id + '/packages/' + packageData.id;

    return API.doRequest('PUT', url, JSON.stringify(processedPackage));
  }

  // Instances

  static getInstances(applicationID: string, groupID: string, queryOptions = {}) {
    let url = BASE_URL + '/apps/' + applicationID + '/groups/' + groupID + '/instances';

    if (!_.isEmpty(queryOptions)) {
      url += '?' + queryString.stringify(queryOptions);
    }

    return API.getJSON(url);
  }

  static getInstancesCount(applicationID: string, groupID: string, duration: string) {
    const url = `${BASE_URL}/apps/${applicationID}/groups/${groupID}/instancescount?duration=${duration}`;

    return API.getJSON(url);
  }

  static getInstance(applicationID: string, groupID: string, instanceID: string) {
    const url =
      BASE_URL + '/apps/' + applicationID + '/groups/' + groupID + '/instances/' + instanceID;

    return API.getJSON(url);
  }

  static getInstanceStatusHistory(applicationID: string, groupID: string, instanceID: string) {
    const url =
      BASE_URL +
      '/apps/' +
      applicationID +
      '/groups/' +
      groupID +
      '/instances/' +
      instanceID +
      '/status_history';

    return API.getJSON(url);
  }

  static updateInstance(instanceID: string, alias: REQUEST_DATA_TYPE) {
    const url = BASE_URL + '/instances/' + instanceID;
    const params = JSON.stringify({ alias });
    return API.doRequest('PUT', url, params);
  }

  // Activity

  static getActivity() {
    const currentDate = new Date();
    const now = currentDate.toISOString();
    currentDate.setDate(currentDate.getDate() - 7);
    const weekAgo = currentDate.toISOString();
    const query = '?start=' + weekAgo + '&end=' + now;
    const url = BASE_URL + '/activity' + query;

    return API.getJSON(url);
  }

  // Config

  static getConfig() {
    return API.doRequest('GET', '/config');
  }

  // Helpers

  static removeKeysFromObject(
    data: Package | Application | Group | FlatcarAction | Channel | Partial<Package>,
    valuesToRemove: string[]
  ) {
    return _.omit(data, valuesToRemove);
  }

  // Wrappers

  static getJSON(url: string) {
    PubSub.publish(MAIN_PROGRESS_BAR, 'add');
    const token = getToken();

    return fetch(url
      ,{
        headers:{
          Authorization: `Bearer ${token}`,
        },
      })
      .then(response => {
        if (!response.ok) {
          throw response;
        }

        // The token has been renewed, let's store it.
        const newIdToken = response.headers.get('id_token');
        if (!!newIdToken && getToken() !== newIdToken) {
          console.debug('Refreshed token')
          setToken(newIdToken);
        }
        return response.json();
      })
      .finally(() => PubSub.publish(MAIN_PROGRESS_BAR, 'done'));
  }

  static doRequest(method: 'GET', url: string): Promise<any>;
  static doRequest(
    method: 'POST' | 'PUT' | 'PATCH' | 'DELETE',
    url: string,
    data: REQUEST_DATA_TYPE
  ): Promise<any>;

  static doRequest(method: string, url: string, data: REQUEST_DATA_TYPE = '') {
    const token = getToken();
    PubSub.publish(MAIN_PROGRESS_BAR, 'add');
    let fetchConfigObject: {
      method: string;
      body?: REQUEST_DATA_TYPE;
      headers?: {
        [prop: string]: any;
      };
    } = { method: 'GET' };

    const headers = {
      Authorization: `Bearer ${token}`,
    };

    if (method === 'DELETE') {
      fetchConfigObject = {
        method,
        headers
      };
      return fetch(url, fetchConfigObject).finally(() => PubSub.publish(MAIN_PROGRESS_BAR, 'done'));
    } else {
      if (method !== 'GET') {
        fetchConfigObject = {
          method,
          body: data,
          headers,
        };
      }
    }

    return fetch(url, fetchConfigObject)
      .then(response => {
        if (!response.ok) {
          throw response;
        }
        return response.json();
      })
      .finally(() => PubSub.publish(MAIN_PROGRESS_BAR, 'done'));
  }
}

export default API;
