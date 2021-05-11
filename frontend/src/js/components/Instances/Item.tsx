import chevronDown from '@iconify/icons-mdi/chevron-down';
import chevronUp from '@iconify/icons-mdi/chevron-up';
import { InlineIcon } from '@iconify/react';
import { Box } from '@material-ui/core';
import Collapse from '@material-ui/core/Collapse';
import IconButton from '@material-ui/core/IconButton';
import Link from '@material-ui/core/Link';
import TableCell from '@material-ui/core/TableCell';
import TableRow from '@material-ui/core/TableRow';
import React, { PropsWithChildren } from 'react';
import { Link as RouterLink } from 'react-router-dom';
import semver from 'semver';
import _ from 'underscore';
import LoadingGif from '../../../img/mini_loading.gif';
import API from '../../api/API';
import { Instance } from '../../api/apiDataTypes';
import { cleanSemverVersion, makeLocaleTime } from '../../utils/helpers';
import StatusHistoryContainer from './StatusHistoryContainer';

const TableLabel = function (props: PropsWithChildren<{bgColor?: string; textColor?: string}>) {
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
  const downloadingIcon = instance.statusInfo?.spinning ? <img src={LoadingGif} /> : '';
  const statusDescription = instance.statusInfo?.description;
  const instanceLabel = instance.statusInfo?.className ? (
    <TableLabel
      bgColor={instance.statusInfo?.bgColor}
      textColor={instance.statusInfo?.textColor}
    >
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
    <React.Fragment>
      <TableRow>
        <TableCell>
          <Link to={instancePath} component={RouterLink}>
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
            <Box
            >
              <IconButton
                onClick={onToggle}
              >
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
    </React.Fragment>
  );
}

export default Item;
