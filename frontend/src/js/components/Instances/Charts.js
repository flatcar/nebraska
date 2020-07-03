import { InlineIcon } from '@iconify/react';
import Grid from '@material-ui/core/Grid';
import Tooltip from '@material-ui/core/Tooltip';
import Typography from '@material-ui/core/Typography';
import { makeStyles, useTheme, withStyles } from '@material-ui/styles';
import React from 'react';
import { VictoryLabel, VictoryPie } from 'victory';
import Empty from '../Common/EmptyContent';
import { InstanceCountLabel } from './Common';
import makeStatusDefs from './StatusDefs';

const useStyles = makeStyles(theme => ({
  doughnutLabel: ({color, labelSize}) => ({
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
    whiteSpace: 'pre-line'
  },
}))(Tooltip);

function ProgressDoughnut(props) {
  const {label, data, width = 100, height = 100, color = '#afafaf', icon} = props;
  const [hoverData, setHoverData] = React.useState(null);
  const [showTooltip, setShowTooltip] = React.useState(false);

  const iconSize = '1.1rem';

  const classes = useStyles({color: color, labelSize: iconSize});
  const theme = useTheme();

  const pieSize = (width > height ? height : width);
  const radius = pieSize * .45;

  let totalFilled = 0;
  let valuesSum = 0;
  const dataSet = data.map(({value, color, description}, i) => {
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
      description: description
    };
  });

  // Use a minimum of 0.5 so a little progress is seen, which helps predict how
  // the circle will be filled, and the current status.
  const percentage = Math.max(totalFilled, 0.5);

  function getTooltipText() {
    return hoverData ? hoverData.description : null;
  }
  const mainTooltipText = data.map(({description}) => description).join('\n');

  dataSet.push({
    x: 'remain',
    y: (100 - percentage),
    color: theme.palette.grey['100'],
  });

  return (
    <Grid
      container
      direction="column"
      justify="center"
      alignItems="center"
    >
      <Grid item>
        <svg viewBox={`0 0 ${width} ${height}`} width={width} height={height}>
          <VictoryPie
            standalone={false}
            animate={{ duration: 1000 }}
            width={pieSize}
            height={pieSize}
            data={dataSet}
            radius={radius}
            innerRadius={radius * .8}
            padAngle={.5}
            labels={() => null}
            style={{
              data: { fill: ({datum}) => datum.color }
            }}
            events={[{
              target: 'data',
              eventHandlers: {
                onMouseOver: () => {
                  return [
                    {
                      target: 'data',
                      mutation: ({datum, style}) => {
                        // Set what to show in the tooltip on hover.
                        setHoverData(datum);
                        setShowTooltip(true);

                        // Highlight the bit on hover, if it's not
                        // the remaining percentage.
                        if (datum.x !== 'remain') {
                          return {style: {...style,
                                          stroke: theme.palette.primary.light,
                                          strokeWidth: 2}};
                        }
                      }
                    }
                  ];
                },
                onMouseOut: () => {
                  return [
                    {
                      target: 'data',
                      mutation: () => {
                        // Reset tooltip previously set on hover.
                        setHoverData(null);
                        setShowTooltip(false);
                      }
                    },
                  ];
                }
              }
            }]}
          />
          <VictoryLabel
            textAnchor="middle"
            verticalAnchor="middle"
            x={width / 2}
            y={height / 2}
            text={`${valuesSum.toFixed(1)}%`}
            style={{
              fontSize: `${radius * .25}px`,
              fontFamily: theme.typography.fontFamily,
            }}
            class={classes.innerLabelFontSize}
          />
        </svg>
      </Grid>
      <Grid item
        container
        alignItems="center"
        justify="center"
        spacing={1}
      >
        { icon &&
          <Grid item>
            <InlineIcon icon={icon} color={color} width={iconSize} height={iconSize} />
          </Grid>
        }
        <Grid item>
          <LightTooltip
            title={getTooltipText() || mainTooltipText}
            open={showTooltip}
          >
            <Typography
              onMouseOver={() => { setShowTooltip(true); }}
              onMouseOut={() => { setShowTooltip(false); }}
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

export default function InstanceStatusArea(props) {
  const theme = useTheme();
  const statusDefs = makeStatusDefs(theme);

  const {instanceStats, href, period} = props;
  const instanceStateCount = [
    {
      status: 'InstanceStatusComplete',
      count: [{key: 'complete'}],
    },
    {
      status: 'InstanceStatusDownloaded',
      count: [{key: 'downloaded'}],
    },
    {
      status: 'InstanceStatusOther',
      count: [
        {key: 'onhold', label: 'InstanceStatusOnHold'},
        {key: 'undefined', label: 'InstanceStatusUndefined'},
      ],
    },
    {
      status: 'InstanceStatusInstalled',
      count: [{key: 'installed'}],
    },
    {
      status: 'InstanceStatusDownloading',
      count: [
        {key: 'downloading', label: 'InstanceStatusDownloading'},
        {key: 'update_granted', label: 'InstanceStatusUpdateGranted'},
      ],
    },
    {
      status: 'InstanceStatusError',
      count: [{key: 'error'}],
    },
  ];

  statusDefs['InstanceStatusOther'] = {...statusDefs['InstanceStatusUndefined']};
  statusDefs['InstanceStatusOther'].label = 'Other';

  const totalInstances = instanceStats ? instanceStats.total : 0;
  return (
    totalInstances > 0 ?
      <Grid
        container
        justify="space-between"
        alignItems="center"
      >
        <Grid item xs={4}>
          <InstanceCountLabel countText={totalInstances}
            href={href}
          />
        </Grid>
        <Grid
          item
          container
          justify="space-between"
          xs={8}
        >
          {instanceStateCount.map(({status, count}, i) => {
            // Sort the data entries so the smaller amounts are shown first.
            count.sort((obj1, obj2) => {
              const stats1 = instanceStats[obj1.key];
              const stats2 = instanceStats[obj2.key];
              if (stats1 === stats2)
                return 0;
              if (stats1 < stats2)
                return -1;
              return 1;
            });

            return (
              <Grid item key={i}>
                <ProgressDoughnut
                  data={count.map(({key, label = status}) => {
                    const statusLabel = statusDefs[label].label;
                    return {
                      value: instanceStats[key] / instanceStats['total'],
                      color: statusDefs[label].color,
                      description: `${statusLabel}: ${instanceStats[key]} instances.`,
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
      :
      <Empty>
        No instances have registered with this group for the past {period}.
        <br/><br/>
        Instances will be shown here automatically the next time they request an update.
      </Empty>
  );
}
