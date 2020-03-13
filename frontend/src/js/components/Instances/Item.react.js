import menuDown from '@iconify/icons-mdi/menu-down';
import menuUp from '@iconify/icons-mdi/menu-up';
import { InlineIcon } from '@iconify/react';
import Button from '@material-ui/core/Button';
import Collapse from '@material-ui/core/Collapse';
import Link from '@material-ui/core/Link';
import TableCell from '@material-ui/core/TableCell';
import TableRow from '@material-ui/core/TableRow';
import { styled } from '@material-ui/styles';
import PropTypes from 'prop-types';
import React from "react";
import { Link as RouterLink } from 'react-router-dom';
import semver from "semver";
import _ from "underscore";
import LoadingGif from "../../../img/mini_loading.gif"
import { cleanSemverVersion, makeLocaleTime } from "../../constants/helpers";
import { instancesStore } from "../../stores/Stores";
import Label from '../Common/Label';
import StatusHistoryContainer from "./StatusHistoryContainer.react";

const TableLabel = styled(Label)({
  lineHeight: '45px',
});

function Item(props) {
  let date = props.instance.application.last_check_for_updates;
  let downloadingIcon = props.instance.statusInfo.spinning ? <img src={LoadingGif} /> : '';
  let statusIcon = props.instance.statusInfo.icon ? <i className={props.instance.statusInfo.icon}></i> : '';
  let instanceLabel = props.instance.statusInfo.className ? <TableLabel>{statusIcon} {downloadingIcon} {props.instance.statusInfo.description}</TableLabel> : <div>&nbsp;</div>;
  let version = cleanSemverVersion(props.instance.application.version);
  let currentVersionIndex = props.lastVersionChannel ? _.indexOf(props.versionNumbers, props.lastVersionChannel) : null;
  let versionStyle = 'default';
  let appID = props.instance.application.application_id;
  let groupID = props.instance.application.group_id;
  let instanceID = props.instance.id;
  const [statusHistory, setStatusHistory] = React.useState(props.instance.statusHistory || []);

  function fetchStatusHistoryFromStore() {
    const selected = props.selected;

    if (!selected) {
      instancesStore.getInstanceStatusHistory(appID, groupID, instanceID)
        .done(() => {
          let cachedInstances = instancesStore.getCachedInstances(appID, groupID) || [];
          let instance = cachedInstances.find(({id}) => id == props.instance.id);
          if (instance)
            setStatusHistory(instance.statusHistory);
          props.onToggle(instanceID);
        })
        .fail((error) => {
          if (error.status === 404) {
            props.onToggle(instanceID);
          }
        })
    } else {
      props.onToggle(instanceID);
    }
  }

  function onToggle() {
    fetchStatusHistoryFromStore();
  }

  if (!_.isEmpty(props.lastVersionChannel)) {
    if (version == props.lastVersionChannel) {
      versionStyle = 'success';
    } else if (semver.gt(version, props.lastVersionChannel)) {
      versionStyle = 'info';
    } else {
      let indexDiff = _.indexOf(props.versionNumbers, version) - currentVersionIndex
      if (indexDiff == 1)
        versionStyle = 'warning';
      else
        versionStyle = 'danger';
    }
  }

  let instancePath = `/apps/${appID}/groups/${groupID}/instances/${instanceID}`;

  return(
    <React.Fragment>
      <TableRow>
        <TableCell>
          <Button size="small" onClick={onToggle}>
            {props.instance.ip}&nbsp;
            <InlineIcon icon={ props.selected ? menuUp : menuDown } />
          </Button>
        </TableCell>
        <TableCell>
          <Link to={instancePath} component={RouterLink}>{props.instance.id}</Link>
        </TableCell>
        <TableCell>
          {instanceLabel}
        </TableCell>
        <TableCell>
          <span className={"box--" + versionStyle}>{version}</span>
        </TableCell>
        <TableCell>
          {makeLocaleTime(date)}
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
}

export default Item
