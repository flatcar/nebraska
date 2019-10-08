import { InlineIcon } from '@iconify/react';
import Box from '@material-ui/core/Box';
import Grid from '@material-ui/core/Grid';
import Paper from '@material-ui/core/Paper';
import Typography from '@material-ui/core/Typography';
import { makeStyles, useTheme } from '@material-ui/styles';
import React from 'react';
import { VictoryAnimation, VictoryLabel, VictoryPie } from 'victory';
import Empty from '../Common/EmptyContent';
import ListHeader from '../Common/ListHeader';
import { InstanceCountLabel } from './Common';
import makeStatusDefs from './StatusDefs';

const useStyles = makeStyles(theme => ({
  doughnutLabel: ({color, labelSize}) => ({
    fontSize: labelSize,
    color: color || theme.palette.text.secondary,
    display: 'inline',
  }),
}));

const useInstanceSectionStyles = makeStyles({
  instancesChartPaper: {
    height: '100%',
  },
});

function ProgressDoughnut(props) {
  let {label, data, width=100, height=100, color='#afafaf', icon} = props;

  const iconSize = '1.1rem';

  const classes = useStyles({color: color, labelSize: iconSize});
  const theme = useTheme();

  const pieSize = (width > height ? height : width);
  const radius = pieSize * .45;

  let totalFilled = 0;
  let valuesSum = 0;
  let dataSet = data.map(({value, color, description}, i) => {
    // Ensure that the minimum value displayed is 0.5 if the original value
    // is 0, or 1.5 otherwise. This ensures the user is able to see the bits
    // related to this value in the charts.
    const percentageValue = Math.max(value * 100, value == 0 ? 0.5 : 1.5);

    totalFilled += percentageValue;
    valuesSum += value * 100;

    return {
      x: i,
      y: percentageValue,
      color: color,
    };
  });

  // Use a minimum of 0.5 so a little progress is seen, which helps predict how
  // the circle will be filled, and the current status.
  const percentage = Math.max(totalFilled, 0.5);


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
            width={pieSize} height={pieSize}
            data={dataSet}
            radius={radius}
            innerRadius={radius * .6}
            padAngle={.5}
            labels={() => null}
            style={{
              data: { fill: ({datum}) => datum.color }
            }}
          />
          <VictoryAnimation duration={1000} data={valuesSum}>
            {(value) => {
              return (
                <VictoryLabel
                  textAnchor="middle" verticalAnchor="middle"
                  x={width / 2} y={height / 2}
                  text={`${Math.round(value)}%`}
                  style={{
                    fontSize: `${radius * .25}px`,
                    fontFamily: theme.typography.fontFamily,
                  }}
                  class={classes.innerLabelFontSize}
                />
              );
            }}
          </VictoryAnimation>
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
          <Typography className={classes.doughnutLabel}>{label}</Typography>
        </Grid>
      </Grid>
    </Grid>
  );
}

export default function InstanceChartSection(props) {
  const classes = useInstanceSectionStyles();
  const theme = useTheme();
  let statusDefs = makeStatusDefs(theme);

  let {instanceStats} = props;
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

  let totalInstances = instanceStats ? instanceStats.total : 0;

  return (
    <Paper className={classes.instancesChartPaper}>
      <ListHeader title="Update Progress" />
      <Box padding="1em">
        { totalInstances > 0 ?
          <Grid
            container
            justify="space-between"
            alignItems="center"
          >
            <Grid item xs={4}>
              <InstanceCountLabel countText={totalInstances}/>
            </Grid>
            <Grid
              item
              container
              justify="space-between"
              xs={8}
            >
              {instanceStateCount.map(({status, count}) => {
                // Sort the data entries so the smaller amounts are shown first.
                count.sort((obj1, obj2) => {
                  const stats1 = instanceStats[obj1.key];
                  const stats2 = instanceStats[obj2.key];
                  if (stats1 == stats2)
                    return 0;
                  if (stats1 < stats2)
                    return -1;
                  return 1;
                });

                return (
                  <Grid item>
                    <ProgressDoughnut
                      data={count.map(({key, label=status}) => {
                        return {
                          value: instanceStats[key] / instanceStats['total'],
                          color: statusDefs[label].color,
                        };
                      })}
                      width={125}
                      height={125}
                      {...statusDefs[status]}
                    />
                  </Grid>
                );
              })}
            </Grid>
          </Grid>
          :
          <Empty>No instances have yet registered with this group.<br/><br/>Registration will happen automatically the first time the instance requests an update.</Empty>
        }
      </Box>
    </Paper>
  );
}