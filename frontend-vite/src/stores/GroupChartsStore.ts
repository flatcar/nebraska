import { applicationsStore } from './Stores';
export default class GroupChartsStore {
  versionBreakdownChartData: {
    [key: string]: {
      [key: string]: {
        [key: string]: string;
      };
    };
  };
  statusBreakdownChartData: {
    [key: string]: {
      [key: string]: {
        [key: string]: string;
      };
    };
  };

  constructor() {
    this.versionBreakdownChartData = {};
    this.statusBreakdownChartData = {};
  }

  async getGroupVersionCountTimeline(appID: string, groupID: string, duration: string) {
    let versionCountTimeline;
    versionCountTimeline = this.getVersionChartData(groupID, duration);
    if (!versionCountTimeline) {
      try {
        versionCountTimeline = await applicationsStore().getGroupVersionCountTimeline(
          appID,
          groupID,
          duration
        );
        this.setVersionChartData(groupID, duration, versionCountTimeline);
      } catch (error) {
        throw new Error(`Error getting version count timeline ${error}`);
      }
    }
    return versionCountTimeline;
  }

  setVersionChartData(
    groupID: string,
    duration: string,
    data: {
      [key: string]: string;
    }
  ) {
    if (
      duration === '1h' ||
      (Object.prototype.hasOwnProperty.call(this.versionBreakdownChartData, groupID) &&
        Object.prototype.hasOwnProperty.call(this.versionBreakdownChartData[groupID], duration))
    ) {
      return;
    }
    if (!this.versionBreakdownChartData[groupID]) {
      this.versionBreakdownChartData[groupID] = {};
    }
    this.versionBreakdownChartData[groupID][duration] = data;
  }

  getVersionChartData(groupID: string, duration: string) {
    if (!this.versionBreakdownChartData[groupID]) {
      return null;
    }
    return this.versionBreakdownChartData[groupID][duration];
  }

  async getGroupStatusCountTimeline(appID: string, groupID: string, duration: string) {
    let statusCountTimeline;
    statusCountTimeline = this.getStatusChartData(groupID, duration);
    if (!statusCountTimeline) {
      try {
        statusCountTimeline = await applicationsStore().getGroupStatusCountTimeline(
          appID,
          groupID,
          duration
        );
        this.setStatusChartData(groupID, duration, statusCountTimeline);
      } catch (error) {
        throw new Error(`Error getting status count timeline ${error}`);
      }
    }
    return statusCountTimeline;
  }

  setStatusChartData(
    groupID: string,
    duration: string,
    data: {
      [key: string]: string;
    }
  ) {
    if (
      duration === '1h' ||
      (Object.prototype.hasOwnProperty.call(this.statusBreakdownChartData, groupID) &&
        Object.prototype.hasOwnProperty.call(this.statusBreakdownChartData[groupID], duration))
    ) {
      return;
    }
    if (!this.statusBreakdownChartData[groupID]) {
      this.statusBreakdownChartData[groupID] = {};
    }

    this.statusBreakdownChartData[groupID][duration] = data;
  }

  getStatusChartData(groupID: string, duration: string) {
    if (!this.statusBreakdownChartData[groupID]) {
      return null;
    }
    return this.statusBreakdownChartData[groupID][duration];
  }
}
