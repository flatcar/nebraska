import { createContext } from 'react';
import _ from 'underscore';

import { CONFIG_STORAGE_KEY, NebraskaConfig } from '../stores/redux/features/config';
import { getToken } from '../utils/auth';
import {
  Activity,
  Application,
  Channel,
  FlatcarAction,
  Group,
  Instance,
  Instances,
  InstanceStatusHistory,
  Package,
  VersionBreakdownEntry,
} from './apiDataTypes';

type WithCount<T> = T & {
  count: number;
  totalCount: number;
};

const BASE_URL = '/api';
type REQUEST_DATA_TYPE = BodyInit | null | undefined;

function isNotNullUndefinedOrEmptyString(val: any) {
  return val !== undefined && val !== null && val !== '';
}

export default class API {
  // Applications

  static getApplications(): Promise<WithCount<{ applications: Application[] }>> {
    return API.getJSON(BASE_URL + '/apps');
  }

  static getApplication(applicationID: string): Promise<Application> {
    return API.getJSON(BASE_URL + '/apps/' + applicationID);
  }

  static deleteApplication(applicationID: string) {
    const url = BASE_URL + '/apps/' + applicationID;

    return API.doRequest('DELETE', url, '');
  }

  static createApplication(
    applicationData: Pick<Application, 'name' | 'description' | 'product_id'>,
    clonedFromAppID: string
  ): Promise<Application> {
    let url = BASE_URL + '/apps';
    if (clonedFromAppID) {
      url += '?clone_from=' + clonedFromAppID;
    }

    return API.doRequest('POST', url, JSON.stringify(applicationData));
  }

  static updateApplication(applicationData: Application): Promise<Application> {
    const url = BASE_URL + '/apps/' + applicationData.id;

    return API.doRequest('PUT', url, JSON.stringify(applicationData));
  }

  // Groups

  static getGroup(applicationID: string, groupID: string): Promise<Group> {
    return API.getJSON(BASE_URL + '/apps/' + applicationID + '/groups/' + groupID);
  }

  static deleteGroup(applicationID: string, groupID: string) {
    const url = BASE_URL + '/apps/' + applicationID + '/groups/' + groupID;

    return API.doRequest('DELETE', url, '');
  }

  static createGroup(groupData: Group): Promise<Group> {
    const url = BASE_URL + '/apps/' + groupData.application_id + '/groups';

    return API.doRequest('POST', url, JSON.stringify(groupData));
  }

  static updateGroup(groupData: Group): Promise<Group> {
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

  static getGroupVersionBreakdown(
    applicationID: string,
    groupID: string
  ): Promise<VersionBreakdownEntry[]> {
    return API.getJSON(
      BASE_URL + '/apps/' + applicationID + '/groups/' + groupID + '/version_breakdown'
    );
  }

  // Channels

  static deleteChannel(applicationID: string, channelID: string) {
    const url = BASE_URL + '/apps/' + applicationID + '/channels/' + channelID;

    return API.doRequest('DELETE', url, '');
  }

  static createChannel(channelData: Channel): Promise<Channel> {
    const url = BASE_URL + '/apps/' + channelData.application_id + '/channels';

    return API.doRequest('POST', url, JSON.stringify(channelData));
  }

  static updateChannel(channelData: Channel): Promise<Channel> {
    const keysToRemove = ['id', 'created_ts', 'package'];
    const processedChannel = API.removeKeysFromObject(channelData, keysToRemove);
    const url = BASE_URL + '/apps/' + channelData.application_id + '/channels/' + channelData.id;

    return API.doRequest('PUT', url, JSON.stringify(processedChannel));
  }

  // Packages
  static getPackages(
    applicationID: string,
    searchTerm?: string,
    queryOptions?: {
      [key: string]: any;
    }
  ): Promise<WithCount<{ packages: Package[]; totalCount: number }>> {
    const params = new URLSearchParams();

    if (searchTerm) {
      params.append('searchVersion', searchTerm);
    }

    if (queryOptions) {
      Object.keys(queryOptions).forEach(key => {
        if (isNotNullUndefinedOrEmptyString(queryOptions[key])) {
          params.append(key, queryOptions[key]);
        }
      });
    }

    const queryStr = params.toString();

    const url = `${BASE_URL}/apps/${applicationID}/packages${queryStr ? '?' + queryStr : ''}`;
    return API.doRequest('GET', url);
  }

  static deletePackage(applicationID: string, packageID: string) {
    const url = BASE_URL + '/apps/' + applicationID + '/packages/' + packageID;

    return API.doRequest('DELETE', url, '');
  }

  static createPackage(packageData: Partial<Package>): Promise<Package> {
    const url = BASE_URL + '/apps/' + packageData.application_id + '/packages';

    return API.doRequest('POST', url, JSON.stringify(packageData));
  }

  static updatePackage(packageData: Partial<Package>): Promise<Package> {
    const keysToRemove = ['id', 'created_ts', 'package'];
    const processedPackage = API.removeKeysFromObject(packageData, keysToRemove);
    const url = BASE_URL + '/apps/' + packageData.application_id + '/packages/' + packageData.id;

    return API.doRequest('PUT', url, JSON.stringify(processedPackage));
  }

  // Instances

  static getInstances(
    applicationID: string,
    groupID: string,
    queryOptions: {
      [key: string]: any;
    } = {}
  ): Promise<Instances> {
    let url = BASE_URL + '/apps/' + applicationID + '/groups/' + groupID + '/instances';

    const params = new URLSearchParams();

    Object.keys(queryOptions).forEach(key => {
      if (isNotNullUndefinedOrEmptyString(queryOptions[key])) {
        params.append(key, queryOptions[key]);
      }
    });

    const queryStr = params.toString();
    url += queryStr ? '?' + queryStr : '';

    return API.getJSON(url);
  }

  static getInstancesCount(
    applicationID: string,
    groupID: string,
    duration: string
  ): Promise<number> {
    const url = `${BASE_URL}/apps/${applicationID}/groups/${groupID}/instancescount?duration=${duration}`;

    return new Promise(resolve =>
      API.getJSON(url).then((response: { count: number }) => resolve(response.count))
    );
  }

  static getInstance(
    applicationID: string,
    groupID: string,
    instanceID: string
  ): Promise<Instance> {
    const url =
      BASE_URL + '/apps/' + applicationID + '/groups/' + groupID + '/instances/' + instanceID;

    return API.getJSON(url);
  }

  static getInstanceStatusHistory(
    applicationID: string,
    groupID: string,
    instanceID: string
  ): Promise<InstanceStatusHistory[]> {
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

  static updateInstance(instanceID: string, alias: REQUEST_DATA_TYPE): Promise<Instance> {
    const url = BASE_URL + '/instances/' + instanceID;
    const params = JSON.stringify({ alias });
    return API.doRequest('PUT', url, params);
  }

  // Activity

  static getActivity(): Promise<WithCount<{ activities: Activity[] }>> {
    const currentDate = new Date();
    const now = currentDate.toISOString();
    currentDate.setDate(currentDate.getDate() - 7);
    const weekAgo = currentDate.toISOString();
    const query = '?start=' + weekAgo + '&end=' + now;
    const url = BASE_URL + '/activity' + query;

    return API.getJSON(url);
  }

  // Config

  static getConfig(): Promise<NebraskaConfig> {
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

  private static getAuthHeaders(): { headers: Record<string, string> } | never {
    const token = getToken();
    const nebraska_config = localStorage.getItem(CONFIG_STORAGE_KEY) || '{}';
    const config = JSON.parse(nebraska_config) || {};

    if (Object.keys(config).length > 0 && config.auth_mode === 'oidc') {
      if (!token) {
        // Reject immediately if OIDC is enabled but no token is available
        throw { status: 401 };
      }
      return {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      };
    }
    return { headers: {} };
  }

  static getJSON(url: string) {
    try {
      const { headers } = this.getAuthHeaders();
      return fetch(url, {
        headers,
      }).then(response => {
        if (!response.ok) {
          throw response;
        }
        return response.json();
      });
    } catch (error) {
      return Promise.reject(error);
    }
  }

  static doRequest(method: 'GET', url: string): Promise<any>;
  static doRequest(
    method: 'POST' | 'PUT' | 'PATCH',
    url: string,
    data: REQUEST_DATA_TYPE
  ): Promise<any>;
  static doRequest(method: 'DELETE', url: string, data: REQUEST_DATA_TYPE): Promise<Response>;

  static doRequest(method: string, url: string, data: REQUEST_DATA_TYPE = '') {
    try {
      const { headers: authHeaders } = this.getAuthHeaders();
      const headers: { [key: string]: string } = {
        'Content-Type': 'application/json',
        ...authHeaders,
      };
      let fetchConfigObject: {
        method: string;
        body?: REQUEST_DATA_TYPE;
        headers?: {
          [prop: string]: any;
        };
      } = {
        method: 'GET',
        headers,
      };

      if (method === 'DELETE') {
        fetchConfigObject = {
          method,
          headers,
        };
        return fetch(url, fetchConfigObject);
      } else {
        if (method !== 'GET') {
          fetchConfigObject = {
            method,
            headers,
            body: data,
          };
        }
      }

      return fetch(url, fetchConfigObject).then(response => {
        if (!response.ok) {
          throw response;
        }
        return response.json();
      });
    } catch (error) {
      return Promise.reject(error);
    }
  }
}

export const APIContext = createContext(API);
