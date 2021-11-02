import ActivityStore from './ActivityStore';
import ApplicationsStore from './ApplicationsStore';
import GroupChartsStore from './GroupChartsStore';

interface Stores {
  applicationsStore: ApplicationsStore;
  activityStore: ActivityStore;
  groupChartStore: GroupChartsStore;
}
let stores: Stores | undefined;

export function getStores(noRefresh?: boolean): Stores {
  if (stores === undefined) {
    stores = {
      applicationsStore: new ApplicationsStore(noRefresh),
      activityStore: new ActivityStore(noRefresh),
      groupChartStore: new GroupChartsStore(),
    };
  }
  return stores;
}

export function applicationsStore(noRefresh?: boolean) {
  return getStores(noRefresh).applicationsStore;
}

export function activityStore(noRefresh?: boolean) {
  return getStores(noRefresh).activityStore;
}

export function groupChartStore(noRefresh?: boolean) {
  return getStores(noRefresh).groupChartStore;
}
