import { IconifyIcon } from '@iconify/react';
import { Theme } from '@mui/material';
import Box from '@mui/material/Box';
import Chip from '@mui/material/Chip';
import Grid from '@mui/material/Grid';
import { useTheme } from '@mui/material/styles';
import Typography from '@mui/material/Typography';
import React from 'react';
import _ from 'underscore';

import { Group } from '../../../api/apiDataTypes';
import { makeLocaleTime } from '../../../i18n/dateTime';
import { groupChartStoreContext } from '../../../stores/Stores';
import { getInstanceStatus } from '../../../utils/helpers';
import Loader from '../../common/Loader/Loader';
import SimpleTable from '../../common/SimpleTable/SimpleTable';
import makeStatusDefs from '../../Instances/StatusDefs';
import TimelineChart from './TimelineChart';
import { Duration } from './TimelineChart';

export interface StatusCountTimelineProps {
  duration: Duration;
  group: Group | null;
  isAnimationActive?: boolean;
}

export default function StatusCountTimeline(props: StatusCountTimelineProps) {
  const [selectedEntry, setSelectedEntry] = React.useState(-1);
  const { duration } = props;
  const [timelineChartData, setTimelineChartData] = React.useState<{
    data: {
      index: number;
      timestamp: string;
    }[];
    keys: string[];
    colors: {
      [key: string]: string;
    };
  }>({
    data: [],
    keys: [],
    colors: {},
  });

  const [timeline, setTimeline] = React.useState<{
    timeline: {
      [key: string]: any;
    };
    lastUpdate: Date | string;
  }>({
    timeline: {},
    // A long time ago, to force the first update...
    lastUpdate: new Date(2000, 1, 1),
  });

  const ChartStoreContext = groupChartStoreContext();
  const groupChartStore = React.useContext(ChartStoreContext);

  const theme = useTheme();
  const statusDefs: {
    [key: string]: {
      label: string;
      color: string;
      icon: IconifyIcon;
      queryValue: string;
    };
  } = makeStatusDefs(theme as Theme);

  function makeChartData(groupTimeline: { [key: string]: any }) {
    const data = Object.keys(groupTimeline).map((timestamp, i) => {
      const status = groupTimeline[timestamp];
      const statusCount: {
        [key: string]: any;
      } = {};
      Object.keys(status).forEach((st: string) => {
        const values = status[st];
        const count = Object.values(values).reduce((a: any, b: any) => a + b, 0);
        statusCount[st] = count;
      });

      return {
        index: i,
        timestamp: timestamp,
        ...statusCount,
      };
    });

    const statuses = getStatusFromTimeline(groupTimeline);
    const colors = makeStatusesColors(statuses);

    setTimelineChartData({
      data: data,
      keys: statuses,
      colors: colors,
    });
  }

  function makeStatusesColors(statuses: { [key: string]: any }) {
    const colors: {
      [key: string]: string;
    } = {};

    Object.values(statuses).forEach(status => {
      const statusInfo = getInstanceStatus(status, '');
      colors[status] = statusDefs[statusInfo.type].color;
    });

    return colors;
  }

  function getStatusFromTimeline(timeline: { [key: number]: number }) {
    if (Object.keys(timeline).length === 0) {
      return [];
    }

    return Object.keys(Object.values(timeline)[0]).filter(status => parseInt(status) !== 0);
  }

  function getInstanceCount(selectedEntry: number) {
    const status_breakdown: {
      status: string;
      version: string;
      instances: number;
    }[] = [];
    const statusTimeline: {
      [key: string]: any;
    } = timeline.timeline;

    // Populate it from the selected time one.
    if (!_.isEmpty(statusTimeline) && !_.isEmpty(timelineChartData.data)) {
      const timelineIndex = selectedEntry >= 0 ? selectedEntry : timelineChartData.data.length - 1;
      if (timelineIndex < 0) return [];

      const ts = timelineChartData.data[timelineIndex].timestamp;
      // Create the version breakdown from the timeline
      const entries = statusTimeline[ts] || [];
      for (const status in entries) {
        if (parseInt(status) === 0) {
          continue;
        }

        const versions = entries[status];

        Object.keys(versions).forEach(version => {
          const versionCount = versions[version];
          status_breakdown.push({
            status: status,
            version: version,
            instances: versionCount,
          });
        });
      }
    }

    status_breakdown.forEach((entry: { status: string; version: string;[key: string]: any }) => {
      const statusInfo = getInstanceStatus(parseInt(entry.status), entry.version);
      const statusTheme = statusDefs[statusInfo.type];

      entry.color = statusTheme.color;
      entry.status = statusTheme.label;
    });

    // Sort the entries per number of instances (higher first).
    status_breakdown.sort((elem1, elem2) => {
      return -(elem1.instances - elem2.instances);
    });

    return status_breakdown;
  }

  function getSelectedTime() {
    const data = timelineChartData.data;
    if (selectedEntry < 0 || data.length === 0) {
      return '';
    }
    const timestamp = data[selectedEntry].timestamp;
    return makeLocaleTime(timestamp);
  }

  // Make the timeline data again when needed.
  React.useEffect(() => {
    async function getStatusTimeline(group: Group | null) {
      if (group) {
        setTimelineChartData({ data: [], keys: [], colors: {} });
        try {
          const statusCountTimeline = await groupChartStore.getGroupStatusCountTimeline(
            group.application_id,
            group.id,
            duration.queryValue
          );
          setTimeline({
            timeline: statusCountTimeline,
            lastUpdate: new Date().toUTCString(),
          });

          makeChartData(statusCountTimeline || []);
          setSelectedEntry(-1);
        } catch (error) {
          console.error(error);
        }
      }
    }
    setSelectedEntry(-1);
    getStatusTimeline(props.group);
  }, [props.duration]);

  return (
    <Grid container alignItems="center" spacing={2}>
      <Grid size={12}>
        {timelineChartData.data.length > 0 ? (
          <TimelineChart
            {...timelineChartData}
            interpolation="step"
            onSelect={setSelectedEntry}
            isAnimationActive={props.isAnimationActive}
          />
        ) : (
          <Loader />
        )}
      </Grid>
      <Grid container size={12}>
        <Grid size={12}>
          <Box width={500}>
            {timelineChartData.data.length > 0 ? (
              selectedEntry !== -1 ? (
                <React.Fragment>
                  <Typography component="span">Showing for:</Typography>
                  &nbsp;
                  <Chip
                    label={getSelectedTime()}
                    onDelete={() => {
                      setSelectedEntry(-1);
                    }}
                  />
                </React.Fragment>
              ) : (
                <Box sx={{
                  color: 'text.secondary',
                  fontSize: 14,
                  textAlign: 'center',
                  lineHeight: 1.5,
                }}>
                  Showing data for the last time point.
                  <br />
                  Click the chart to choose a different time point.
                </Box>
              )
            ) : null}
          </Box>
        </Grid>
        <Grid size={12}>
          {timelineChartData.data.length > 0 && (
            <SimpleTable
              emptyMessage="No data to show for this time point."
              columns={{ status: 'Status', version: 'Version', instances: 'Instances' }}
              instances={getInstanceCount(selectedEntry)}
            />
          )}
        </Grid>
      </Grid>
    </Grid >
  );
}
