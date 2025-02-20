import Box from '@mui/material/Box';
import Paper from '@mui/material/Paper';
import Typography from '@mui/material/Typography';
import { Area, AreaChart, AreaProps, CartesianGrid, Tooltip, XAxis, YAxis } from 'recharts';
import { getMinuteDifference, makeLocaleTime } from '../../../i18n/dateTime';

export interface Duration {
  displayValue: string;
  queryValue: string;
  disabled: boolean;
}

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

export interface TimelineChartProps {
  width?: number;
  height?: number;
  interpolation?: AreaProps['type'];
  data: any;
  onSelect: (activeLabel: any) => void;
  colors: any;
  keys: string[];
  isAnimationActive?: boolean;
}

export default function TimelineChart(props: TimelineChartProps) {
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
      onClick={(obj: any) => obj && props.onSelect(obj.activeLabel)}
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
        stroke={'#000'}
      />
      <YAxis stroke="#000" />
      {props.keys.map((key: string, i: number) => (
        <Area
          isAnimationActive={props.isAnimationActive}
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
