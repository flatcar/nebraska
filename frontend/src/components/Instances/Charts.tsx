import { IconifyIcon, InlineIcon } from '@iconify/react';
import { Theme } from '@mui/material';
import Grid from '@mui/material/Grid';
import Typography from '@mui/material/Typography';
import { makeStyles, useTheme } from '@mui/styles';
import React from 'react';
import { Trans, useTranslation } from 'react-i18next';
import { Cell, Label, Pie, PieChart } from 'recharts';

import Empty from '../common/EmptyContent';
import LightTooltip from '../common/LightTooltip';
import Loader from '../common/Loader';
import { InstanceCountLabel } from './Common';
import makeStatusDefs from './StatusDefs';

interface DoughnutLabelProps {
  color: string;
  labelSize: string;
}

const useStyles = makeStyles((theme: Theme) => ({
  doughnutLabel: ({ color, labelSize }: DoughnutLabelProps) => ({
    fontSize: labelSize,
    color: color || theme.palette.text.secondary,
    display: 'inline',
    boxShadow: 'none',
  }),
}));

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
  icon: IconifyIcon;
  data: ProgressData[];
}

interface RechartsPieData {
  x: number;
  y: number;
  color: string;
  description?: string;
}

function ProgressDoughnut(props: ProgressDoughnutProps) {
  const { label, data, width = 100, height = 100, color = '#afafaf', icon } = props;
  const [hoverData, setHoverData] = React.useState<RechartsPieData | null>(null);
  const [showTooltip, setShowTooltip] = React.useState(false);
  const [activeIndex, setActiveIndex] = React.useState(-1);
  const iconSize = '1.1rem';

  const classes = useStyles({ color: color, labelSize: iconSize });
  const theme = useTheme<Theme>();

  const pieSize = width > height ? height : width;
  const radius = pieSize * 0.45;

  let totalFilled = 0;
  let valuesSum = 0;
  const dataSet: RechartsPieData[] = data.map(({ value, color, description }, i) => {
    // Ensure that the minimum value displayed is 0.5 if the original value
    // is 0, or 1.5 otherwise. This ensures the user is able to see the bits
    // related to this value in the charts.
    const percentageValue = Math.max(value * 100, value === 0 ? 0.5 : 1.5);

    totalFilled += percentageValue;
    valuesSum += value * 100;

    return {
      x: i,
      y: percentageValue,
      color: color,
      description: description,
    };
  });

  // Use a minimum of 0.5 so a little progress is seen, which helps predict how
  // the circle will be filled, and the current status.
  const percentage = Math.max(totalFilled, 0.5);

  function getTooltipText() {
    return hoverData ? hoverData.description : null;
  }
  const mainTooltipText = data.map(({ description }) => description).join('\n');

  dataSet.push({
    x: percentage,
    y: 100 - percentage,
    color: theme.palette.grey['100'],
  });

  return (
    <Grid container direction="column" justifyContent="center" alignItems="center">
      <Grid>
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
            onMouseOver={(dataum, index) => {
              if (!showTooltip) {
                setHoverData(dataum);
                setShowTooltip(true);
                // Highlight the bit on hover, if it's not
                // the remaining percentage.
                if (dataum.x !== 'remain') {
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
                  fill={entry.color}
                  stroke={activeIndex === index ? theme.palette.primary.light : '#fff'}
                  strokeWidth={activeIndex === index ? 2 : 0}
                />
              );
            })}
          </Pie>
        </PieChart>
      </Grid>
      <Grid container alignItems="center" justifyContent="center" spacing={1}>
        {icon && (
          <Grid>
            <InlineIcon icon={icon} color={color} width={iconSize} height={iconSize} />
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
              className={classes.doughnutLabel}
            >
              {label}
            </Typography>
          </LightTooltip>
        </Grid>
      </Grid>
    </Grid>
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
  const theme = useTheme<Theme>();
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
    <Grid container justifyContent="space-between" alignItems="center">
      <Grid size={4}>
        <InstanceCountLabel countText={totalInstances} href={href} />
      </Grid>
      <Grid container justifyContent="space-between" size={8}>
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
