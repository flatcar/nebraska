import ActivityStore from './ActivityStore';
import ApplicationsStore from './ApplicationsStore';
import GroupChartsStore from './GroupChartsStore';

interface Stores {
  applicationsStore: ApplicationsStore;
  activityStore: ActivityStore;
  groupChartStore: GroupChartsStore;
}
let stores: Stores | undefined;

export function getStores(): Stores {
  if (stores === undefined) {
    stores = {
      applicationsStore: new ApplicationsStore(),
      activityStore: new ActivityStore(),
      groupChartStore: new GroupChartsStore(),
    };
  }
  return stores;
}

export function applicationsStore() {
  return getStores().applicationsStore;
}

export function activityStore() {
  return getStores().activityStore;
}

export function groupChartStore() {
  return getStores().groupChartStore;
}
