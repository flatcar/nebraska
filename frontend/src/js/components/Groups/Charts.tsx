import { IconifyIcon } from '@iconify/react';
import { Theme } from '@material-ui/core';
import Box from '@material-ui/core/Box';
import Chip from '@material-ui/core/Chip';
import Grid from '@material-ui/core/Grid';
import Paper from '@material-ui/core/Paper';
import Typography from '@material-ui/core/Typography';
import { useTheme } from '@material-ui/styles';
import React from 'react';
import { Area, AreaChart, CartesianGrid, LineType, Tooltip, XAxis, YAxis } from 'recharts';
import semver from 'semver';
import _ from 'underscore';
import { Group } from '../../api/apiDataTypes';
import {
  cleanSemverVersion,
  getInstanceStatus,
  getMinuteDifference,
  makeColorsForVersions,
  makeLocaleTime
} from '../../utils/helpers';
import Loader from '../Common/Loader';
import SimpleTable from '../Common/SimpleTable';
import makeStatusDefs from '../Instances/StatusDefs';

function TimelineTooltip(props: { label?: string; data: any }) {
  const { label, data } = props;
  return (
    <div className="custom-tooltip">
      <Paper>
        <Box padding={1}>
          <Typography>{label && data[label] && makeLocaleTime(data[label].timestamp)}</Typography>
        </Box>
      </Paper>
    </div>
  );
}

function TimelineChart(props: {
  width?: number;
  height?: number;
  interpolation?: LineType;
  data: any;
  onSelect: (activeLabel: any) => void;
  colors: any;
  keys: string[];
}) {
  const { width = 500, height = 400, interpolation = 'monotone' } = props;
  let ticks: {
    [key: string]: string;
  } = {};

  function getTickValues() {
    const DAY = 24 * 60;
    let tickCount = 4;
    let dateFormat: {
      useDate?: boolean;
      showTime?: boolean;
      dateFormat?: Intl.DateTimeFormatOptions;
    } = { useDate: false };
    const startTs = new Date(props.data[0].timestamp);
    const endTs = new Date(props.data[props.data.length - 1].timestamp);
    const lengthMinutes = getMinuteDifference(endTs.valueOf(), startTs.valueOf());
    // We remove 1 element since that's "0 hours"
    const dimension = props.data.length - 1;

    // Reset the ticks for the chart
    ticks = {};

    if (lengthMinutes === 7 * DAY) {
      tickCount = 7;
      dateFormat = { dateFormat: { month: 'short', day: 'numeric' }, showTime: false };
    }
    if (lengthMinutes === 60) {
      for (let i = 0; i < 4; i++) {
        const minuteValue = (lengthMinutes / 4) * i;
        startTs.setMinutes(new Date(props.data[0].timestamp).getMinutes() + minuteValue);
        ticks[i] = makeLocaleTime(startTs, { useDate: false });
      }
      return ticks;
    }

    if (lengthMinutes === 30 * DAY) {
      for (let i = 0; i < props.data.length; i += 2) {
        const tickDate = new Date(props.data[i].timestamp);
        ticks[i] = makeLocaleTime(tickDate, {
          showTime: false,
          dateFormat: { month: 'short', day: 'numeric' },
        });
      }
      return ticks;
    }
    // Set up a tick marking the 0 hours of the day contained in the range
    const nextDay = new Date(startTs);
    nextDay.setHours(24, 0, 0, 0);
    const midnightDay = new Date(nextDay);
    const nextDayMinuteDiff = getMinuteDifference(nextDay.valueOf(), startTs.valueOf());
    const midnightTick = (nextDayMinuteDiff * dimension) / lengthMinutes;

    // Set up the remaining ticks according to the desired amount, separated
    // evenly.
    const tickOffsetMinutes = lengthMinutes / tickCount;

    // Set the ticks around midnight.
    for (const i of [-1, 1]) {
      const tickDate = new Date(nextDay);

      while (true) {
        tickDate.setMinutes(nextDay.getMinutes() + tickOffsetMinutes * i);
        // Stop if this tick falls outside of the times being charted

        if (tickDate < startTs || tickDate > endTs) {
          break;
        }

        const tick =
          (getMinuteDifference(tickDate.valueOf(), startTs.valueOf()) * dimension) / lengthMinutes;
        // Show only the time.
        ticks[tick] = makeLocaleTime(tickDate, dateFormat);
      }
    }
    // The midnight tick just gets the date, not the hours (since they're zero)
    ticks[midnightTick] = makeLocaleTime(midnightDay, {
      dateFormat: { month: 'short', day: 'numeric' },
      showTime: false,
    });
    return ticks;
  }

  return (
    <AreaChart
      width={width}
      height={height}
      data={props.data}
      margin={{
        top: 10,
        right: 30,
        left: 0,
        bottom: 0,
      }}
      onClick={obj => obj && props.onSelect(obj.activeLabel)}
    >
      <CartesianGrid strokeDasharray="3 3" />
      <Tooltip content={<TimelineTooltip data={props.data} />} />
      <XAxis
        dataKey="index"
        type="number"
        interval={0}
        domain={[0, 'dataMax']}
        ticks={Object.keys(getTickValues())}
        tickFormatter={(index: string) => {
          return ticks[index];
        }}
      />
      <YAxis />
      {props.keys.map((key: string, i: number) => (
        <Area
          type={interpolation}
          key={i}
          dataKey={key}
          stackId="1"
          stroke={props.colors[key]}
          cursor="pointer"
          fill={props.colors[key]}
        />
      ))}
    </AreaChart>
  );
}

export function VersionCountTimeline(props: {
  group: Group | null;
  duration: {
    [key: string]: any;
  };
}) {
  const [selectedEntry, setSelectedEntry] = React.useState(-1);
  const { duration } = props;
  const [timelineChartData, setTimelineChartData] = React.useState<{
    data: any[];
    keys: any[];
    colors: any;
  }>({
    data: [],
    keys: [],
    colors: [],
  });
  const [timeline, setTimeline] = React.useState({
    timeline: {},
    // A long time ago, to force the first update...
    lastUpdate: new Date(2000, 1, 1).toUTCString(),
  });

  const theme = useTheme();

  function makeChartData(group: Group, groupTimeline: { [key: string]: any }) {
    const data = Object.keys(groupTimeline).map((timestamp, i) => {
      const versions = groupTimeline[timestamp];
      return {
        index: i,
        timestamp: timestamp,
        ...versions,
      };
    });

    const versions = getVersionsFromTimeline(groupTimeline);
    const versionColors: {
      [key: string]: string;
    } = makeColorsForVersions(theme as Theme, versions, group.channel);

    setTimelineChartData({
      data: data,
      keys: versions,
      colors: versionColors,
    });
  }

  function getVersionsFromTimeline(timeline: { [key: string]: any }) {
    if (Object.keys(timeline).length === 0) {
      return [];
    }

    const versions: string[] = [];

    Object.keys(Object.values(timeline)[0]).forEach(version => {
      const cleanedVersion = cleanSemverVersion(version);
      // Discard any invalid versions (empty strings, etc.)
      if (semver.valid(cleanedVersion)) {
        versions.push(cleanedVersion);
      }
    });

    // Sort versions (earliest first)
    versions.sort((version1, version2) => {
      return semver.compare(version1, version2);
    });

    return versions;
  }

  function getInstanceCount(selectedEntry: number) {
    const version_breakdown = [];
    let selectedEntryPoint = selectedEntry;

    // If there is no timeline or no specific time is selected,
    // show the timeline for the last time point.
    if (selectedEntry === -1) {
      selectedEntryPoint = timelineChartData.data.length - 1;
    }

    let total = 0;

    // If we're not using the default group version breakdown,
    // let's populate it from the selected time one.
    if (version_breakdown.length === 0 && selectedEntryPoint > -1) {
      // Create the version breakdown from the timeline
      const entries = timelineChartData.data[selectedEntryPoint] || [];

      for (const version of timelineChartData.keys) {
        const versionCount = entries[version];

        total += versionCount;

        version_breakdown.push({
          version: version,
          instances: versionCount,
          percentage: 0,
        });
      }
    }

    version_breakdown.forEach((entry: { [key: string]: any }) => {
      entry.color = timelineChartData.colors[entry.version];

      // Calculate the percentage if needed.
      if (total > 0) {
        entry.percentage = (entry.instances * 100.0) / total;
      }

      entry.percentage = parseFloat(entry.percentage).toFixed(1);
    });

    // Sort the entries per number of instances (higher first).
    version_breakdown.sort((elem1, elem2) => {
      return -(elem1.instances - elem2.instances);
    });

    return version_breakdown;
  }

  function getSelectedTime() {
    const data = timelineChartData.data;
    if (selectedEntry < 0 || data.length === 0) {
      return '';
    }
    const timestamp = data[selectedEntry] ? data[selectedEntry].timestamp : '';
    return makeLocaleTime(timestamp);
  }

  // Make the timeline data again when needed.
  React.useEffect(() => {
    let canceled = false;
    async function getVersionTimeline(group: Group | null) {
      if (group) {
        // Check if we should update the timeline or it's too early.
        const lastUpdate = new Date(timeline.lastUpdate);
        setTimelineChartData({ data: [], keys: [], colors: [] });
        try {
          const versionCountTimeline = await groupChartStore.getGroupVersionCountTimeline(
            group.application_id,
            group.id,
            duration.queryValue
          );
          if (!canceled) {
            setTimeline({
              timeline: versionCountTimeline,
              lastUpdate: lastUpdate.toUTCString(),
            });
          }
          makeChartData(group, versionCountTimeline || []);
          setSelectedEntry(-1);
        } catch (error) {
          console.error(error);
        }
      }
    }
    getVersionTimeline(props.group);
    return () => {
      canceled = true;
    };
  }, [duration]);

  return (
    <Grid container alignItems="center" spacing={2}>
      <Grid item xs={12}>
        {timelineChartData.data.length > 0 ? (
          <TimelineChart {...timelineChartData} onSelect={setSelectedEntry} />
        ) : (
          <Loader />
        )}
      </Grid>
      <Grid item xs={12} container>
        <Grid item xs={12}>
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
                <Box color="text.secondary" fontSize={14} textAlign="center" lineHeight={1.5}>
                  Showing data for the last time point.
                  <br />
                  Click the chart to choose a different time point.
                </Box>
              )
            ) : null}
          </Box>
        </Grid>
        <Grid item xs={11}>
          {timelineChartData.data.length > 0 && (
            <SimpleTable
              emptyMessage="No data to show for this time point."
              columns={{ version: 'Version', instances: 'Count', percentage: 'Percentage' }}
              instances={getInstanceCount(selectedEntry)}
            />
          )}
        </Grid>
      </Grid>
    </Grid>
  );
}

export function StatusCountTimeline(props: {
  duration: {
    [key: string]: any;
  };
  group: Group | null;
}) {
  const [selectedEntry, setSelectedEntry] = React.useState(-1);
  const { duration } = props;
  const [timelineChartData, setTimelineChartData] = React.useState<{
    data: {
      index: number;
      timestamp: string;
    }[];
    keys: string[];
    colors: {
      [key: string]: any;
    };
  }>({
    data: [],
    keys: [],
    colors: [],
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

  const theme = useTheme();
  const statusDefs: {
    [key: string]: {
      label: string;
      color: any;
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
      [key: string]: any;
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
    async function getStatusTimeline(group: Group | null) {
      if (group) {
        setTimelineChartData({ data: [], keys: [], colors: [] });
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
      <Grid item xs={12}>
        {timelineChartData.data.length > 0 ? (
          <TimelineChart {...timelineChartData} interpolation="step" onSelect={setSelectedEntry} />
        ) : (
          <Loader />
        )}
      </Grid>
      <Grid item xs={12} container>
        <Grid item xs={12}>
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
                <Box color="text.secondary" fontSize={14} textAlign="center" lineHeight={1.5}>
                  Showing data for the last time point.
                  <br />
                  Click the chart to choose a different time point.
                </Box>
              )
            ) : null}
          </Box>
        </Grid>
        <Grid item xs={12}>
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
