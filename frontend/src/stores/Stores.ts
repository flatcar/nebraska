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

export function getStores(): Stores {
  if (stores === undefined) {
    const applicationsStore = new ApplicationsStore();
    const activityStore = new ActivityStore();
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

export function applicationsStore() {
  return getStores().applicationsStore;
}

export function activityStore() {
  return getStores().activityStore;
}

export function groupChartStore() {
  return getStores().groupChartStore;
}

export function groupChartStoreContext() {
  return getStores().groupChartStoreContext;
}
