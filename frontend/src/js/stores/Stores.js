import ActivityStore from './ActivityStore';
import ApplicationsStore from './ApplicationsStore';
import GroupChartsStore from './GroupChartsStore';
import InstancesStore from './InstancesStore';

export const applicationsStore = new ApplicationsStore;
export const activityStore = new ActivityStore;
export const instancesStore = new InstancesStore;
export const groupChartStore=new GroupChartsStore();
