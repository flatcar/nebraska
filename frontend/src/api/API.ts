import { createContext } from 'react';
import _ from 'underscore';

import { NebraskaConfig } from '../stores/redux/features/config';
import store from '../stores/redux/store';
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
  static getApplications(): Promise<WithCount<{ applications: Application[] }>> {
    return API.getJSON(`${BASE_URL}/apps`);
  }

  static getApplication(applicationID: string): Promise<Application> {
    return API.getJSON(`${BASE_URL}/apps/${applicationID}`);
  }

  static deleteApplication(applicationID: string) {
    return API.doRequest('DELETE', `${BASE_URL}/apps/${applicationID}`, '');
  }

  static createApplication(
    applicationData: Pick<Application, 'name' | 'description' | 'product_id'>,
    clonedFromAppID: string
  ): Promise<Application> {
    const url = clonedFromAppID
      ? `${BASE_URL}/apps?clone_from=${clonedFromAppID}`
      : `${BASE_URL}/apps`;
    return API.doRequest('POST', url, JSON.stringify(applicationData));
  }

  static updateApplication(applicationData: Application): Promise<Application> {
    return API.doRequest(
      'PUT',
      `${BASE_URL}/apps/${applicationData.id}`,
      JSON.stringify(applicationData)
    );
  }

  static getGroup(applicationID: string, groupID: string): Promise<Group> {
    return API.getJSON(`${BASE_URL}/apps/${applicationID}/groups/${groupID}`);
  }

  static deleteGroup(applicationID: string, groupID: string) {
    return API.doRequest('DELETE', `${BASE_URL}/apps/${applicationID}/groups/${groupID}`, '');
  }

  static createGroup(groupData: Group): Promise<Group> {
    return API.doRequest(
      'POST',
      `${BASE_URL}/apps/${groupData.application_id}/groups`,
      JSON.stringify(groupData)
    );
  }

  static updateGroup(groupData: Group): Promise<Group> {
    return API.doRequest(
      'PUT',
      `${BASE_URL}/apps/${groupData.application_id}/groups/${groupData.id}`,
      JSON.stringify(groupData)
    );
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
    return API.getJSON(`${BASE_URL}/apps/${applicationID}/groups/${groupID}/version_breakdown`);
  }

  static deleteChannel(applicationID: string, channelID: string) {
    return API.doRequest('DELETE', `${BASE_URL}/apps/${applicationID}/channels/${channelID}`, '');
  }

  static createChannel(channelData: Channel): Promise<Channel> {
    return API.doRequest(
      'POST',
      `${BASE_URL}/apps/${channelData.application_id}/channels`,
      JSON.stringify(channelData)
    );
  }

  static updateChannel(channelData: Channel): Promise<Channel> {
    const keysToRemove = ['id', 'created_ts', 'package'];
    const processedChannel = API.removeKeysFromObject(channelData, keysToRemove);
    return API.doRequest(
      'PUT',
      `${BASE_URL}/apps/${channelData.application_id}/channels/${channelData.id}`,
      JSON.stringify(processedChannel)
    );
  }

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
    return API.doRequest('DELETE', `${BASE_URL}/apps/${applicationID}/packages/${packageID}`, '');
  }

  static createPackage(packageData: Partial<Package>): Promise<Package> {
    return API.doRequest(
      'POST',
      `${BASE_URL}/apps/${packageData.application_id}/packages`,
      JSON.stringify(packageData)
    );
  }

  static updatePackage(packageData: Partial<Package>): Promise<Package> {
    const keysToRemove = ['id', 'created_ts', 'package'];
    const processedPackage = API.removeKeysFromObject(packageData, keysToRemove);
    return API.doRequest(
      'PUT',
      `${BASE_URL}/apps/${packageData.application_id}/packages/${packageData.id}`,
      JSON.stringify(processedPackage)
    );
  }

  static async setChannelFloor(channelID: string, packageID: string, floorReason?: string) {
    const data = floorReason ? JSON.stringify({ floor_reason: floorReason }) : '{}';
    try {
      return await API.doRequest(
        'PUT',
        `${BASE_URL}/channels/${channelID}/floors/${packageID}`,
        data
      );
    } catch (error) {
      // Handle empty response (204 No Content) or other success with no body
      if (error instanceof Response && (error.status === 204 || error.status === 201)) {
        return {};
      }
      throw error;
    }
  }

  static async deleteChannelFloor(channelID: string, packageID: string) {
    return API.doRequest('DELETE', `${BASE_URL}/channels/${channelID}/floors/${packageID}`, '');
  }

  static getChannelFloors(
    channelID: string,
    queryOptions: { page?: number; perpage?: number } = {}
  ): Promise<WithCount<{ packages: Package[] }>> {
    // Backend defaults to 100 items for floor packages which should be sufficient
    const params = new URLSearchParams();
    if (queryOptions.page !== undefined) {
      params.append('page', String(queryOptions.page));
    }
    if (queryOptions.perpage !== undefined) {
      params.append('perpage', String(queryOptions.perpage));
    }
    const queryStr = params.toString();
    const url = `${BASE_URL}/channels/${channelID}/floors${queryStr ? '?' + queryStr : ''}`;
    return API.getJSON(url);
  }

  static getPackageFloorChannels(
    applicationID: string,
    packageID: string
  ): Promise<{
    channels: Array<{ channel: Channel; floor_reason: string | null }>;
    count: number;
  }> {
    return API.getJSON(`${BASE_URL}/apps/${applicationID}/packages/${packageID}/floor-channels`);
  }

  static getInstances(
    applicationID: string,
    groupID: string,
    queryOptions: {
      [key: string]: any;
    } = {}
  ): Promise<Instances> {
    const params = new URLSearchParams();
    Object.keys(queryOptions).forEach(key => {
      if (isNotNullUndefinedOrEmptyString(queryOptions[key])) {
        params.append(key, queryOptions[key]);
      }
    });
    const queryStr = params.toString();
    const url = `${BASE_URL}/apps/${applicationID}/groups/${groupID}/instances${queryStr ? '?' + queryStr : ''}`;
    return API.getJSON(url);
  }

  static getInstancesCount(
    applicationID: string,
    groupID: string,
    duration: string
  ): Promise<number> {
    return new Promise(resolve =>
      API.getJSON(
        `${BASE_URL}/apps/${applicationID}/groups/${groupID}/instancescount?duration=${duration}`
      ).then((response: { count: number }) => resolve(response.count))
    );
  }

  static getInstance(
    applicationID: string,
    groupID: string,
    instanceID: string
  ): Promise<Instance> {
    return API.getJSON(
      `${BASE_URL}/apps/${applicationID}/groups/${groupID}/instances/${instanceID}`
    );
  }

  static getInstanceStatusHistory(
    applicationID: string,
    groupID: string,
    instanceID: string
  ): Promise<InstanceStatusHistory[]> {
    return API.getJSON(
      `${BASE_URL}/apps/${applicationID}/groups/${groupID}/instances/${instanceID}/status_history`
    );
  }

  static updateInstance(instanceID: string, alias: REQUEST_DATA_TYPE): Promise<Instance> {
    return API.doRequest('PUT', `${BASE_URL}/instances/${instanceID}`, JSON.stringify({ alias }));
  }

  static getActivity(): Promise<WithCount<{ activities: Activity[] }>> {
    const currentDate = new Date();
    const now = currentDate.toISOString();
    currentDate.setDate(currentDate.getDate() - 7);
    const weekAgo = currentDate.toISOString();
    return API.getJSON(`${BASE_URL}/activity?start=${weekAgo}&end=${now}`);
  }

  static getConfig(): Promise<NebraskaConfig> {
    return API.doRequest('GET', '/config');
  }

  static removeKeysFromObject(
    data: Package | Application | Group | FlatcarAction | Channel | Partial<Package>,
    valuesToRemove: string[]
  ) {
    return _.omit(data, valuesToRemove);
  }

  private static getAuthHeaders(): { headers: Record<string, string> } {
    const token = getToken();
    const config = store.getState().config;

    if (Object.keys(config).length > 0 && config.auth_mode === 'oidc' && token) {
      return {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      };
    }
    return { headers: {} };
  }

  static async getJSON(url: string) {
    const { headers } = this.getAuthHeaders();
    const response = await fetch(url, {
      headers,
    });
    if (!response.ok) {
      throw response;
    }
    return await response.json();
  }

  static doRequest(method: 'GET', url: string): Promise<any>;
  static doRequest(
    method: 'POST' | 'PUT' | 'PATCH',
    url: string,
    data: REQUEST_DATA_TYPE
  ): Promise<any>;
  static doRequest(method: 'DELETE', url: string, data: REQUEST_DATA_TYPE): Promise<Response>;

  static doRequest(method: string, url: string, data: REQUEST_DATA_TYPE = '') {
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
  }
}

export const APIContext = createContext(API);
