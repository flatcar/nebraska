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

function makeEmptyTimelineChartData() {
  return { data: [], keys: [], colors: {} };
}

function getStatusFromTimeline(timeline: { [key: number]: number }) {
  if (Object.keys(timeline).length === 0) {
    return [];
  }

  return Object.keys(Object.values(timeline)[0]).filter(status => parseInt(status) !== 0);
}

function makeStatusesColors(
  statuses: { [key: string]: any },
  statusDefs: {
    [key: string]: {
      label: string;
      color: string;
      icon: IconifyIcon;
      queryValue: string;
    };
  }
) {
  const colors: {
    [key: string]: string;
  } = {};

  Object.values(statuses).forEach(status => {
    const statusInfo = getInstanceStatus(status, '');
    colors[status] = statusDefs[statusInfo.type].color;
  });

  return colors;
}

function makeTimelineChartData(
  groupTimeline: { [key: string]: any },
  statusDefs: {
    [key: string]: {
      label: string;
      color: string;
      icon: IconifyIcon;
      queryValue: string;
    };
  }
) {
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
  const colors = makeStatusesColors(statuses, statusDefs);

  return {
    data: data,
    keys: statuses,
    colors: colors,
  };
}

export default function StatusCountTimeline(props: StatusCountTimelineProps) {
  const [selectedEntry, setSelectedEntry] = React.useState(-1);
  const { duration, group } = props;
  const applicationID = group?.application_id;
  const groupID = group?.id;
  const durationQueryValue = duration.queryValue;
  const [timelineChartData, setTimelineChartData] = React.useState<{
    data: {
      index: number;
      timestamp: string;
    }[];
    keys: string[];
    colors: {
      [key: string]: string;
    };
  }>(makeEmptyTimelineChartData);

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
  const statusDefsRef = React.useRef(statusDefs);

  React.useEffect(() => {
    statusDefsRef.current = statusDefs;
  });

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

    status_breakdown.forEach((entry: { status: string; version: string; [key: string]: any }) => {
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
    let canceled = false;

    setSelectedEntry(-1);
    setTimelineChartData(makeEmptyTimelineChartData());

    if (!applicationID || !groupID) {
      setTimeline({
        timeline: {},
        lastUpdate: new Date().toUTCString(),
      });
      return () => {
        canceled = true;
      };
    }

    const selectedApplicationID = applicationID;
    const selectedGroupID = groupID;

    async function getStatusTimeline() {
      try {
        const statusCountTimeline = await groupChartStore.getGroupStatusCountTimeline(
          selectedApplicationID,
          selectedGroupID,
          durationQueryValue
        );
        if (canceled) {
          return;
        }

        const safeTimeline = statusCountTimeline || {};
        setTimeline({
          timeline: safeTimeline,
          lastUpdate: new Date().toUTCString(),
        });

        setTimelineChartData(makeTimelineChartData(safeTimeline, statusDefsRef.current));
      } catch (error) {
        if (!canceled) {
          console.error(error);
        }
      }
    }

    getStatusTimeline();

    return () => {
      canceled = true;
    };
  }, [applicationID, durationQueryValue, groupChartStore, groupID]);

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
                <Box
                  sx={{
                    color: 'text.secondary',
                    fontSize: 14,
                    textAlign: 'center',
                    lineHeight: 1.5,
                  }}
                >
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
    </Grid>
  );
}
