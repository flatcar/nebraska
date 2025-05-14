import chevronDown from '@iconify/icons-mdi/chevron-down';
import chevronUp from '@iconify/icons-mdi/chevron-up';
import { InlineIcon } from '@iconify/react';
import { Box } from '@mui/material';
import Collapse from '@mui/material/Collapse';
import IconButton from '@mui/material/IconButton';
import Link from '@mui/material/Link';
import { styled } from '@mui/material/styles';
import TableCell from '@mui/material/TableCell';
import TableRow from '@mui/material/TableRow';
import React, { PropsWithChildren } from 'react';
import { Link as RouterLink } from 'react-router-dom';
import semver from 'semver';
import _ from 'underscore';

import API from '../../api/API';
import { Instance } from '../../api/apiDataTypes';
import { makeLocaleTime } from '../../i18n/dateTime';
import { cleanSemverVersion } from '../../utils/helpers';
import StatusHistoryContainer from './StatusHistoryContainer';

const PREFIX = 'Item';

const classes = {
  link: `${PREFIX}-link`
};

const Root = styled('div')({
  [`& .${classes.link}`]: {
    color: '#1b5c91',
  },
});

const TableLabel = function (props: PropsWithChildren<{ bgColor?: string; textColor?: string }>) {
  return (
    <Box bgcolor={props.bgColor} color={props.textColor} display="inline-block" py={1} px={2}>
      {props.children}
    </Box>
  );
};

interface ItemProps {
  instance: Instance;
  lastVersionChannel: string;
  versionNumbers: string[];
  selected: boolean;
  onToggle: (instanceId: string | null) => void;
}

function Item(props: ItemProps) {
  const { instance, selected, lastVersionChannel, versionNumbers } = props;

  const date = instance.application.last_check_for_updates;
  const statusDescription = instance.statusInfo?.description;
  const instanceLabel = instance.statusInfo?.className ? (
    <TableLabel bgColor={instance.statusInfo?.bgColor} textColor={instance.statusInfo?.textColor}>
      {statusDescription}
    </TableLabel>
  ) : (
    <div>&nbsp;</div>
  );
  const version = cleanSemverVersion(instance.application.version);
  const currentVersionIndex = lastVersionChannel
    ? _.indexOf(versionNumbers, lastVersionChannel)
    : 0;
  let versionStyle = 'default';
  const appID = instance.application.application_id;
  const groupID = instance.application.group_id;
  const instanceID = instance.id;
  const [statusHistory, setStatusHistory] = React.useState(instance.statusHistory || []);

  function fetchStatusHistoryFromStore() {
    const isSelected = selected;

    if (!isSelected) {
      API.getInstanceStatusHistory(appID, groupID, instanceID)
        .then(statusHistory => {
          setStatusHistory(statusHistory);
          props.onToggle(instanceID);
        })
        .catch(error => {
          if (error.status === 404) {
            props.onToggle(instanceID);
            setStatusHistory([]);
          }
        });
    } else {
      props.onToggle(instanceID);
    }
  }

  function onToggle() {
    fetchStatusHistoryFromStore();
  }

  if (!_.isEmpty(props.lastVersionChannel)) {
    if (version === props.lastVersionChannel) {
      versionStyle = 'success';
    } else if (semver.valid(version) && semver.gt(version, props.lastVersionChannel)) {
      versionStyle = 'info';
    } else {
      const indexDiff = _.indexOf(props.versionNumbers, version) - currentVersionIndex;
      if (indexDiff === 1) versionStyle = 'warning';
      else versionStyle = 'danger';
    }
  }
  const searchParams = new URLSearchParams(window.location.search).toString();
  const instancePath = `/apps/${appID}/groups/${groupID}/instances/${instanceID}?${searchParams}`;
  const instanceName = props.instance.alias || props.instance.id;

  return (
    <Root>
      <TableRow>
        <TableCell>
          <Link to={instancePath} component={RouterLink} className={classes.link} underline="hover">
            {instanceName}
          </Link>
        </TableCell>
        <TableCell>{props.instance.ip}</TableCell>
        <TableCell>{instanceLabel}</TableCell>
        <TableCell>
          <span className={'box--' + versionStyle}>{version}</span>
        </TableCell>
        <TableCell>
          <Box display="flex" justifyContent="space-between">
            <Box>{makeLocaleTime(date)}</Box>
            <Box>
              <IconButton onClick={onToggle} size="large">
                <InlineIcon
                  icon={props.selected ? chevronUp : chevronDown}
                  height="25"
                  width="25"
                  color="#808080"
                  style={{ cursor: 'pointer' }}
                />
              </IconButton>
            </Box>
          </Box>
        </TableCell>
      </TableRow>
      <TableRow>
        <TableCell padding="none" colSpan={5}>
          <Collapse in={props.selected}>
            <StatusHistoryContainer statusHistory={statusHistory} />
          </Collapse>
        </TableCell>
      </TableRow>
    </Root>
  );
}

export default Item;
