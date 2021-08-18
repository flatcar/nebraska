import { IconifyIcon, InlineIcon } from '@iconify/react';
import { Theme } from '@material-ui/core';
import Grid from '@material-ui/core/Grid';
import Tooltip from '@material-ui/core/Tooltip';
import Typography from '@material-ui/core/Typography';
import { makeStyles, useTheme, withStyles } from '@material-ui/styles';
import React from 'react';
import { Trans, useTranslation } from 'react-i18next';
import { Cell, Label, Pie, PieChart } from 'recharts';
import Empty from '../Common/EmptyContent';
import Loader from '../Common/Loader';
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

const LightTooltip = withStyles(theme => ({
  tooltip: {
    backgroundColor: theme.palette.common.white,
    color: 'rgba(0, 0, 0, 0.87)',
    boxShadow: theme.shadows[1],
    fontSize: '1rem',
    whiteSpace: 'pre-line',
  },
}))(Tooltip);

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
    <Grid container direction="column" justify="center" alignItems="center">
      <Grid item>
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
      <Grid item container alignItems="center" justify="center" spacing={1}>
        {icon && (
          <Grid item>
            <InlineIcon icon={icon} color={color} width={iconSize} height={iconSize} />
          </Grid>
        )}
        <Grid item>
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
  groupHasVersion: boolean;
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

  const { instanceStats, href, period, groupHasVersion } = props;
  const instanceStateCount: InstanceStatusCount[] = [
    {
      status: 'InstanceStatusComplete',
      count: [{ key: 'complete' }],
    },
    {
      status: 'InstanceStatusNotUpdating',
      count: [{ key: 'other_versions', label: t('instances|InstanceStatusOtherVersions') }],
    },
    {
      status: 'InstanceStatusOnHold',
      count: [{ key: 'onhold' }],
    },
    {
      status: 'InstanceStatusUpdating',
      count: [
        { key: 'update_granted', label: t('instances|InstanceStatusUpdateGranted') },
        { key: 'downloading', label: t('instances|InstanceStatusDownloading') },
        { key: 'downloaded', label: t('instances|InstanceStatusDownloaded') },
        { key: 'installed', label: t('instances|InstanceStatusInstalled') },
      ],
    },
    {
      status: 'InstanceStatusError',
      count: [{ key: 'error' }],
    },
    {
      status: 'InstanceStatusTimedOut',
      count: [{ key: 'timed_out' }],
    },
  ];

  statusDefs['InstanceStatusNotUpdating'] = { ...statusDefs['InstanceStatusUndefined'] };
  statusDefs['InstanceStatusNotUpdating'].label = t('instances|Not updating');

  statusDefs['InstanceStatusUpdating'] = {
    ...statusDefs['InstanceStatusDownloading'],
    label: t('instances|Updating'),
  };

  const totalInstances = instanceStats ? instanceStats.total : 0;

  if (!instanceStats) {
    return <Loader />;
  }

  return totalInstances > 0 ? (
    <Grid container justify="space-between" alignItems="center">
      <Grid item xs={4}>
        <InstanceCountLabel countText={totalInstances} href={href} />
      </Grid>
      <Grid item container justify="space-between" xs={8}>
        {!groupHasVersion ? (
          <Empty>
            <Trans ns="instances">
              It's not possible to get an accurate report as the group has no channel/version
              assigned to it.
            </Trans>
          </Empty>
        ) : (
          instanceStateCount.map(({ status, count }, i) => {
            // Sort the data entries so the smaller amounts are shown first.
            count.sort((obj1, obj2) => {
              const stats1 = instanceStats[obj1.key];
              const stats2 = instanceStats[obj2.key];
              if (stats1 === stats2) return 0;
              if (stats1 < stats2) return -1;
              return 1;
            });

            return (
              <Grid item key={i}>
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
          })
        )}
      </Grid>
    </Grid>
  ) : (
    <Empty>
      <Trans ns="instances">
        No instances have registered with this group for the past {period}.
        <br />
        <br />
        Instances will be shown here automatically the next time they request an update.
      </Trans>
    </Empty>
  );
}
