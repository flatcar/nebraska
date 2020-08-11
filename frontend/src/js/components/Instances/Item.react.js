import chevronDown from '@iconify/icons-mdi/chevron-down';
import chevronUp from '@iconify/icons-mdi/chevron-up';
import { InlineIcon } from '@iconify/react';
import { Box } from '@material-ui/core';
import Collapse from '@material-ui/core/Collapse';
import Link from '@material-ui/core/Link';
import TableCell from '@material-ui/core/TableCell';
import TableRow from '@material-ui/core/TableRow';
import PropTypes from 'prop-types';
import React from 'react';
import { Link as RouterLink } from 'react-router-dom';
import semver from 'semver';
import _ from 'underscore';
import LoadingGif from '../../../img/mini_loading.gif';
import API from '../../api/API';
import { cleanSemverVersion, makeLocaleTime } from '../../constants/helpers';
import StatusHistoryContainer from './StatusHistoryContainer.react';

const TableLabel = function(props){
  return (<Box bgcolor={props.bgColor} color={props.textColor} display="inline-block" py={1} px={2}>

    {props.children}
  </Box>);
};
function Item(props) {
  const date = props.instance.application.last_check_for_updates;
  const downloadingIcon = props.instance.statusInfo.spinning ? <img src={LoadingGif} /> : '';
  const statusDescription = props.instance.statusInfo.description;
  const instanceLabel = props.instance.statusInfo.className ?
    <TableLabel bgColor={props.instance.statusInfo.bgColor}
      textColor={props.instance.statusInfo.textColor}
    >
      {statusDescription}
    </TableLabel> : <div>&nbsp;</div>;
  const version = cleanSemverVersion(props.instance.application.version);
  const currentVersionIndex = props.lastVersionChannel ?
    _.indexOf(props.versionNumbers, props.lastVersionChannel) : null;
  let versionStyle = 'default';
  const appID = props.instance.application.application_id;
  const groupID = props.instance.application.group_id;
  const instanceID = props.instance.id;
  const [statusHistory, setStatusHistory] = React.useState(props.instance.statusHistory || []);

  function fetchStatusHistoryFromStore() {
    const selected = props.selected;

    if (!selected) {

      API.getInstanceStatusHistory(appID, groupID, instanceID)
        .then((statusHistory) => {
          setStatusHistory(statusHistory);
          props.onToggle(instanceID);
        })
        .catch((error) => {
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
      if (indexDiff === 1)
        versionStyle = 'warning';
      else
        versionStyle = 'danger';
    }
  }
  const searchParams = new URLSearchParams(window.location.search).toString();
  const instancePath = `/apps/${appID}/groups/${groupID}/instances/${instanceID}?${searchParams}`;

  return (
    <React.Fragment>
      <TableRow>
        <TableCell>
          <Link to={instancePath} component={RouterLink}>{props.instance.id}</Link>
        </TableCell>
        <TableCell>
          {props.instance.ip}
        </TableCell>
        <TableCell>
          {instanceLabel}
        </TableCell>
        <TableCell>
          <span className={'box--' + versionStyle}>{version}</span>
        </TableCell>
        <TableCell>
          <Box display="flex" justifyContent="space-between">
            <Box>
              {makeLocaleTime(date)}
            </Box>
            <Box>
              <InlineIcon icon={ props.selected ? chevronUp : chevronDown }
                onClick={onToggle}
                height="25"
                width="25"
                color="#808080"
                style={{cursor: 'pointer'}}
              />
            </Box>
          </Box>
        </TableCell>
      </TableRow>
      <TableRow>
        <TableCell padding="none" colSpan={5}>
          <Collapse
            hidden={!props.selected}
            in={props.selected}
          >
            <StatusHistoryContainer statusHistory={statusHistory} />
          </Collapse>
        </TableCell>
      </TableRow>
    </React.Fragment>
  );
}

Item.propTypes = {
  instance: PropTypes.object.isRequired,
  selected: PropTypes.bool,
  versionNumbers: PropTypes.array,
  lastVersionChannel: PropTypes.string
};

export default Item;
