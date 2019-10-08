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
  let {label, value, width=100, height=100, color='#afafaf', icon} = props;

  const iconSize = '1.1rem';

  const classes = useStyles({color: color, labelSize: iconSize});
  const theme = useTheme();

  const pieSize = (width > height ? height : width);
  const radius = pieSize * .45;
  // Use a minimum of 0.5 so a little progress is seen, which helps predict how
  // the circle will be filled, and the current status.
  const percentage = Math.max(value * 100, 0.5);


  let data = [{
    x: 1,
    y: percentage,
  },
  {
    x: 2,
    y: (100 - percentage),
  }];
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
            data={data}
            radius={radius}
            innerRadius={radius * .6}
            padAngle={.5}
            labels={() => null}
            style={{
              data: { fill: ({datum}) => {
                  if (datum.x === 2)
                    return theme.palette.grey['100'];
                  return color;
                }
              }
            }}
          />
          <VictoryAnimation duration={1000} data={value * 100}>
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
  const statusDefs = makeStatusDefs(theme);

  let {instanceStats} = props;
  const instanceStateCount = {
    InstanceStatusComplete: instanceStats['complete'],
    InstanceStatusDownloaded: instanceStats['downloaded'],
    InstanceStatusOnHold: instanceStats['onhold'],
    InstanceStatusInstalled: instanceStats['installed'],
    InstanceStatusDownloading: instanceStats['downloading'],
    InstanceStatusError: instanceStats['error'],
  };
  let {filter=Object.keys(instanceStateCount)} = props;
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
            {filter.map(key => {
                  return (
                    <Grid item>
                      <ProgressDoughnut
                        value={instanceStateCount[key] / instanceStats['total']}
                        width={125}
                        height={125}
                        {...statusDefs[key]}
                      />
                    </Grid>
                  );
                })
            }
            </Grid>
          </Grid>
          :
          <Empty>No instances have yet registered with this group.<br/><br/>Registration will happen automatically the first time the instance requests an update.</Empty>
        }
      </Box>
    </Paper>
  );
}