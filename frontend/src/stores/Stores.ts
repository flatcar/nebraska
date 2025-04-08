import { Context, createContext } from 'react';

import ActivityStore from './ActivityStore';
import ApplicationsStore from './ApplicationsStore';
import GroupChartsStore from './GroupChartsStore';

interface Stores {
  applicationsStore: ApplicationsStore;
  activityStore: ActivityStore;
  groupChartStore: GroupChartsStore;

  applicationsStoreContext: Context<ApplicationsStore>;
  activityStoreContext: Context<ActivityStore>;
  groupChartStoreContext: Context<GroupChartsStore>;
}
let stores: Stores | undefined;

export function getStores(noRefresh?: boolean): Stores {
  if (stores === undefined) {
    const applicationsStore = new ApplicationsStore(noRefresh);
    const activityStore = new ActivityStore(noRefresh);
    const groupChartStore = new GroupChartsStore();

    const applicationsStoreContext = createContext(applicationsStore);
    const activityStoreContext = createContext(activityStore);
    const groupChartStoreContext = createContext(groupChartStore);

    stores = {
      applicationsStore,
      activityStore,
      groupChartStore,
      applicationsStoreContext,
      activityStoreContext,
      groupChartStoreContext,
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

export function groupChartStoreContext() {
  return getStores().groupChartStoreContext;
}
