import Box from '@mui/material/Box';
import Grid from '@mui/material/Grid';
import Stack from '@mui/material/Stack';
import { useTheme } from '@mui/material/styles';
import Typography from '@mui/material/Typography';
import React from 'react';
import { Trans, useTranslation } from 'react-i18next';
import { Cell, Label, Pie, PieChart } from 'recharts';

import Empty from '../common/EmptyContent';
import LightTooltip from '../common/LightTooltip';
import Loader from '../common/Loader';
import { InstanceCountLabel } from './Common';
import makeStatusDefs from './StatusDefs';

interface ProgressData {
  value: number;
  color: string;
  description: string;
}

interface ProgressDoughnutProps {
  label: string;
  width: number;
  height: number;
  color: string;
  icon: React.ElementType;
  data: ProgressData[];
}

interface PieDataEntry {
  x: number;
  y: number;
  color: string;
  description?: string;
}

function ProgressDoughnut(props: ProgressDoughnutProps) {
  const { label, data, width = 100, height = 100, color = '#afafaf', icon: StatusIcon } = props;
  const [hoverData, setHoverData] = React.useState<PieDataEntry | null>(null);
  const [showTooltip, setShowTooltip] = React.useState(false);
  const [activeIndex, setActiveIndex] = React.useState(-1);
  const iconSize = '1.1rem';

  const theme = useTheme();

  const pieSize = width > height ? height : width;
  const radius = pieSize * 0.45;

  // Ensure that the minimum value displayed is 0.5 if the original value
  // is 0, or 1.5 otherwise. This ensures the user is able to see the bits
  // related to this value in the charts.
  const percentageValue = (value: number) => Math.max(value * 100, value === 0 ? 0.5 : 1.5);

  const totalFilled = data.reduce((acc, { value }) => acc + percentageValue(value), 0);
  const valuesSum = data.reduce((acc, { value }) => acc + value * 100, 0);
  const dataSet: PieDataEntry[] = data.map(({ value, color, description }, i) => {
    return {
      x: i,
      y: percentageValue(value),
      color: color,
      description: description,
    };
  });

  // Use a minimum of 0.5 so a little progress is seen, which helps predict how
  // the circle will be filled, and the current status.
  const percentage = Math.max(totalFilled, 0.5);

  function getTooltipText() {
    return hoverData?.description ?? null;
  }
  const mainTooltipText = data.map(({ description }) => description).join('\n');

  dataSet.push({
    x: percentage,
    y: 100 - percentage,
    color: theme.palette.grey['100'],
  });

  return (
    <Stack
      direction="column"
      sx={{
        justifyContent: 'center',
        alignItems: 'center',
      }}
    >
      <Box>
        <PieChart width={width} height={height}>
          <Pie
            data={dataSet}
            dataKey="y"
            nameKey="x"
            paddingAngle={0.5}
            outerRadius={radius}
            isAnimationActive
            startAngle={90}
            endAngle={-270}
            innerRadius={radius * 0.8}
            animationDuration={1000}
            animationEasing={'ease-in-out'}
            onMouseOver={(_datum, index) => {
              if (!showTooltip) {
                setHoverData(dataSet[index]);
                setShowTooltip(true);
                // Highlight the bit on hover, if it's not
                // the remaining percentage.
                if (index < dataSet.length - 1) {
                  setActiveIndex(index);
                }
              }
            }}
            onMouseOut={() => {
              setActiveIndex(-1);
              setShowTooltip(false);
              setHoverData(null);
            }}
          >
            <Label position="center" value={`${valuesSum.toFixed(1)}%`} />
            {dataSet.map((entry, index) => {
              return (
                <Cell
                  key={`cell-${index}`}
                  fill={entry.color as string}
                  stroke={activeIndex === index ? theme.palette.primary.light : '#fff'}
                  strokeWidth={activeIndex === index ? 2 : 0}
                />
              );
            })}
          </Pie>
        </PieChart>
      </Box>
      <Grid
        container
        spacing={1}
        sx={{
          alignItems: 'center',
          justifyContent: 'center',
        }}
      >
        {StatusIcon && (
          <Grid>
            <StatusIcon sx={{ color, fontSize: iconSize }} />
          </Grid>
        )}
        <Grid>
          <LightTooltip title={getTooltipText() || mainTooltipText} open={showTooltip}>
            <Typography
              onMouseOver={() => {
                setShowTooltip(true);
              }}
              onMouseOut={() => {
                setShowTooltip(false);
              }}
              sx={{
                fontSize: iconSize,
                color: theme => color || theme.palette.text.secondary,
                display: 'inline',
                boxShadow: 'none',
              }}
            >
              {label}
            </Typography>
          </LightTooltip>
        </Grid>
      </Grid>
    </Stack>
  );
}

interface InstanceStats {
  [key: string]: number;
  total: number;
}

interface InstanceStatusAreaProps {
  instanceStats: InstanceStats | null;
  href?: object;
  period: string;
}

interface InstanceStatusCount {
  status: string;
  count: {
    key: string;
    label?: string;
  }[];
}

export default function InstanceStatusArea(props: InstanceStatusAreaProps) {
  const theme = useTheme();
  const statusDefs = makeStatusDefs(theme);
  const { t } = useTranslation();

  const { instanceStats, href, period } = props;
  const instanceStateCount: InstanceStatusCount[] = [
    {
      status: 'InstanceStatusComplete',
      count: [{ key: 'complete' }],
    },
    {
      status: 'InstanceStatusDownloaded',
      count: [{ key: 'downloaded' }],
    },
    {
      status: 'InstanceStatusOther',
      count: [
        { key: 'onhold', label: t('instances|instance_status_on_hold') },
        { key: 'undefined', label: t('instances|instance_status_undefined') },
      ],
    },
    {
      status: 'InstanceStatusInstalled',
      count: [{ key: 'installed' }],
    },
    {
      status: 'InstanceStatusDownloading',
      count: [
        { key: 'downloading', label: t('instances|instance_status_downloading') },
        { key: 'update_granted', label: t('instances|instance_status_update_granted') },
      ],
    },
    {
      status: 'InstanceStatusError',
      count: [{ key: 'error' }],
    },
  ];

  statusDefs['InstanceStatusOther'] = { ...statusDefs['InstanceStatusUndefined'] };
  statusDefs['InstanceStatusOther'].label = t('instances|other');

  const totalInstances = instanceStats ? instanceStats.total : 0;

  if (!instanceStats) {
    return <Loader />;
  }

  return totalInstances > 0 ? (
    <Grid
      container
      sx={{
        justifyContent: 'space-between',
        alignItems: 'center',
      }}
    >
      <Grid size={4}>
        <InstanceCountLabel countText={totalInstances} href={href} />
      </Grid>
      <Grid
        container
        size={8}
        sx={{
          justifyContent: 'space-between',
        }}
      >
        {instanceStateCount.map(({ status, count }, i) => {
          // Sort the data entries so the smaller amounts are shown first.
          count.sort((obj1, obj2) => {
            const stats1 = instanceStats[obj1.key];
            const stats2 = instanceStats[obj2.key];
            if (stats1 === stats2) return 0;
            if (stats1 < stats2) return -1;
            return 1;
          });

          return (
            <Grid key={i}>
              <ProgressDoughnut
                data={count.map(({ key, label = status }) => {
                  const statusLabel = statusDefs[label].label;
                  return {
                    value: instanceStats[key] / instanceStats.total,
                    color: statusDefs[label].color,
                    description: t('{{statusLabel}}: {{stat, number}} instances', {
                      statusLabel: statusLabel,
                      stat: instanceStats[key],
                    }),
                  };
                })}
                width={140}
                height={140}
                {...statusDefs[status]}
              />
            </Grid>
          );
        })}
      </Grid>
    </Grid>
  ) : (
    <Empty>
      <Trans t={t} ns="instances" i18nKey="noinstances">
        No instances have registered with this group for the past {period}.
        <br />
        <br />
        Instances will be shown here automatically the next time they request an update.
      </Trans>
    </Empty>
  );
}
