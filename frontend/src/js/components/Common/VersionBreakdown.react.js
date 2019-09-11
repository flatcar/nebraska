import Box from '@material-ui/core/Box';
import Grid from '@material-ui/core/Grid';
import { makeStyles } from '@material-ui/core/styles';
import Typography from '@material-ui/core/Typography';
import { ResponsiveBar } from '@nivo/bar';
import PropTypes from 'prop-types';
import React from "react";
import semver from 'semver';
import { cleanSemverVersion } from '../../constants/helpers';
import { CardFeatureLabel } from '../Common/Card';

const useStyles = makeStyles(theme => ({
  subtitle: {
    fontSize: '0.9em',
    color: theme.palette.text.secondary,
    textTransform: 'lowercase',
    display: 'inline',
  }
}));

function ProgressBar({keys, data}) {
  function setUpFill() {
    return ['info', 'warning', 'danger', 'success'].map(level => {
      return {
        match: ({data}) => data.data[data.id + 'Level'] === level,
        id: level
      };
    });
  }

  return (
    <ResponsiveBar
      data={data}
      keys={keys}
      // This is only because it needs an index, even though we're not
      // adding more than one bar.
      indexBy="dummy"
      margin={{ top: 50, right: 20, bottom: 50, left: 20 }}
      padding={0.3}
      layout="horizontal"
      colors={{ scheme: 'nivo' }}
      label={({id, data}) => data[id + 'Label']}
      tooltip={({ id, color }) => (
        <strong style={{ color }}>{id}</strong>
      )}
      theme={{
        tooltip: {
          container: {
            background: '#333',
          },
        },
      }}
      defs={[
        {
          id: 'success',
          type: 'patternDots',
          background: '#61cdbb',
          color: '#38bcb2',
          size: 4,
          padding: 1,
          stagger: true
        },
        {
          id: 'warning',
          type: 'patternLines',
          background: '#f1e15b',
          color: '#eed312',
          rotation: -45,
          lineWidth: 6,
          spacing: 10
        },
        {
          id: 'danger',
          type: 'patternLines',
          background: '#f47560',
          color: '#ff5600',
          rotation: -45,
          lineWidth: 6,
          spacing: 10
        },
        {
          id: 'info',
          type: 'patternLines',
          background: '#a4a4a4',
          color: '#737373',
          rotation: -45,
          lineWidth: 6,
          spacing: 10
        }
      ]}
      fill={setUpFill()}
      borderColor={{ from: 'color', modifiers: [ [ 'darker', 1.6 ] ] }}
      axisTop={null}
      axisRight={null}
      axisBottom={null}
      axisLeft={null}
      labelSkipWidth={12}
      labelSkipHeight={12}
      labelTextColor='black'
      legends={[]}
      animate={true}
      motionStiffness={90}
      motionDamping={15}
    />
  );
}

function VersionBreakdown(props) {
  const classes = useStyles();
  let versions = props.version_breakdown || [];
  let lastVersionChannel = '';
  let data = [];
  let channel = props.channel || {};
  let legendVersion = null;
  let versionsValues = versions.map(version =>
    cleanSemverVersion(version.version)
  ).sort(semver.rcompare);

  // Early return if we don't have a progress bar to show
  if (!versionsValues)
    return (<Typography variant='h7'>No instances available!</Typography>);

  versions.forEach(version => {
    let barStyle = 'default';
    let cleanVersion = cleanSemverVersion(version.version);
    let labelLegend = cleanVersion;

    if (channel) {
      lastVersionChannel = channel.package ? cleanSemverVersion(channel.package.version) : '';
      let currentVersionIndex = versionsValues.indexOf(lastVersionChannel);

      if (lastVersionChannel) {
        if (cleanSemverVersion(version.version) == lastVersionChannel) {
          barStyle = 'success';
          labelLegend = cleanSemverVersion(version.version) + "*"
          legendVersion = <Typography className={classes.subtitle}>{"*Current channel version"}</Typography>
        } else if (semver.gt(cleanSemverVersion(version.version), lastVersionChannel)) {
          barStyle = 'info';
        } else {
          let indexDiff = versionsValues.indexOf(cleanSemverVersion(version.version)) - currentVersionIndex;
          barStyle = indexDiff == 1 ? 'warning' : 'danger';
        }
      } else {
        legendVersion = <Typography className={classes.subtitle}>{"No colors available as channel is not pointing to any package"}</Typography>
      }
    }

    data[cleanVersion] = version.percentage;
    data[cleanVersion + 'Label'] = labelLegend;
    data[cleanVersion + 'Level'] = barStyle;
  });

  return (
    <Grid container>
      <Grid item xs={12} container justify="space-between">
        <Grid item>
          <CardFeatureLabel>
            Version breakdown
          </CardFeatureLabel>
        </Grid>
        {legendVersion &&
        <Grid item>
          {legendVersion}
        </Grid>
        }
      </Grid>
      <Grid item xs={12}>
        <Box height={30}>
          <ProgressBar keys={versionsValues} data={[data]} />
        </Box>
      </Grid>
    </Grid>
  );
}

VersionBreakdown.propTypes = {
  version_breakdown: PropTypes.array.isRequired,
  channel: PropTypes.object.isRequired
}

export default VersionBreakdown
