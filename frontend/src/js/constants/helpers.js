import amber from '@material-ui/core/colors/amber';
import deepOrange from '@material-ui/core/colors/deepOrange';
import lime from '@material-ui/core/colors/lime';
import orange from '@material-ui/core/colors/orange';
import red from '@material-ui/core/colors/red';
import React from 'react';
import API from '../api/API';

// Indexes/keys for the architectures need to match the ones in
// pkg/api/arches.go.
export const ARCHES = {
  0: 'ALL',
  1: 'AMD64',
  2: 'ARM64',
  3: 'X86',
};

const colors = makeColors();

function makeColors() {
  const colors = [];

  const colorScheme = [lime, amber, orange, deepOrange, red];

  // Create a color list for versions to pick from.
  colorScheme.forEach(color => {
    // We choose the shades beyond 300 because they should not be too
    // light (in order to improve contrast).
    for (let i = 3; i <= 9; i += 2) {
      colors.push(color[i * 100]);
    }
  });

  return colors;
}

export function cleanSemverVersion(version) {
  let shortVersion = version;
  if (version.includes('+')) {
    shortVersion = version.split('+')[0];
  }
  return shortVersion;
}

export function getMinuteDifference(date1, date2) {
  return (date1 - date2) / 1000 / 60;
}

export function makeLocaleTime(timestamp, opts = {}) {
  const {useDate = true, showTime = true, dateFormat = {weekday: 'short', day: 'numeric'}} = opts;
  const date = new Date(timestamp);
  const formattedDate = date.toLocaleDateString('default', dateFormat);
  const timeFormat = date.toLocaleString('default', { hour: '2-digit', minute: '2-digit' });

  if (useDate && showTime) {
    return `${formattedDate} ${timeFormat}`;
  }
  if (useDate) {
    return formattedDate;
  }
  return timeFormat;
}

export function makeColorsForVersions(theme, versions, channel = null) {
  const versionColors = {};
  let colorIndex = 0;
  let latestVersion = null;

  if (channel && channel.package) {
    latestVersion = cleanSemverVersion(channel.package.version);
  }

  for (let i = versions.length - 1; i >= 0; i--) {
    const version = versions[i];
    const cleanVersion = cleanSemverVersion(version);

    if (cleanVersion === latestVersion) {
      versionColors[cleanVersion] = theme.palette.success.main;
    } else {
      versionColors[cleanVersion] = colors[colorIndex++ % colors.length];
    }
  }

  return versionColors;
}

export function getInstanceStatus(statusID, version) {
  const status = {
    1: {
      type: 'InstanceStatusUndefined',
      className: '',
      spinning: false,
      icon: '',
      description: '',
      status: 'Undefined',
      explanation: ''
    },
    2: {
      type: 'InstanceStatusUpdateGranted',
      className: 'warning',
      spinning: true,
      icon: '',
      description: 'Updating: granted',
      status: 'Granted',
      explanation: 'The instance has received an update package -version ' + version + '- and the update process is about to start'
    },
    3: {
      type: 'InstanceStatusError',
      className: 'danger',
      spinning: false,
      icon: 'glyphicon glyphicon-remove',
      description: 'Error updating',
      status: 'Error',
      explanation: 'The instance reported an error while updating to version ' + version
    },
    4: {
      type: 'InstanceStatusComplete',
      className: 'success',
      spinning: false,
      icon: 'glyphicon glyphicon-ok',
      description: 'Update completed',
      status: 'Completed',
      explanation: 'The instance has been updated successfully and is now running version ' + version
    },
    5: {
      type: 'InstanceStatusInstalled',
      className: 'warning',
      spinning: true,
      icon: '',
      description: 'Updating: installed',
      status: 'Installed',
      explanation: 'The instance has installed the update package -version ' + version + '- but it isn’t using it yet'
    },
    6: {
      type: 'InstanceStatusDownloaded',
      className: 'warning',
      spinning: true,
      icon: '',
      description: 'Updating: downloaded',
      status: 'Downloaded',
      explanation: 'The instance has downloaded the update package -version ' + version + '- and will install it now'
    },
    7: {
      type: 'InstanceStatusDownloading',
      className: 'warning',
      spinning: true,
      icon: '',
      description: 'Updating: downloading',
      status: 'Downloading',
      explanation: 'The instance has just started downloading the update package -version ' + version + '-'
    },
    8: {
      type: 'InstanceStatusOnHold',
      className: 'default',
      spinning: false,
      icon: '',
      description: 'Waiting...',
      status: 'On hold',
      explanation: 'There was an update pending for the instance but it was put on hold because of the rollout policy'
    }
  };

  const statusDetails = statusID ? status[statusID] : status[1];

  return statusDetails;
}

export function useGroupVersionBreakdown(group) {
  const [versionBreakdown, setVersionBreakdown] = React.useState([]);

  React.useEffect(() => {
    if (!group) {
      return;
    }

    const version_breakdown = API.getGroupVersionBreakdown(group.application_id, group.id)
      .then(version_breakdown => {
        setVersionBreakdown(version_breakdown || []);
      })
      .catch(err => {
        console.error('Error getting version breakdown for group', group.id, '\nError:', err);
      });
  },
  [group]);

  return versionBreakdown;
}
