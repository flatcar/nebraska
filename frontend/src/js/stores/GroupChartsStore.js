export default class GroupChartsStore{
  constructor(){
    this.VersionBreakdownChartData={};
    this.StatusBreakdownChartData={};
  }
  setVersionChartData(key, data){
    if (this.VersionBreakdownChartData.hasOwnProperty(key))
    {
      return;
    }
    this.VersionBreakdownChartData[key]=data;
  }
  getVersionChartData(key){
    return this.VersionBreakdownChartData[key];
  }
  setStatusChartData(key, data){
    if (this.StatusBreakdownChartData.hasOwnProperty(key)){
      return;
    }
    this.StatusBreakdownChartData[key]=data;
  }
  getStatusChartData(key){
    return this.StatusBreakdownChartData[key];
  }
}

