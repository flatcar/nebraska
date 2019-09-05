import menuDown from '@iconify/icons-mdi/menu-down';
import menuUp from '@iconify/icons-mdi/menu-up';
import { InlineIcon } from '@iconify/react';
import Button from '@material-ui/core/Button';
import Collapse from '@material-ui/core/Collapse';
import TableCell from '@material-ui/core/TableCell';
import TableRow from '@material-ui/core/TableRow';
import { styled } from '@material-ui/styles';
import moment from "moment";
import PropTypes from 'prop-types';
import React from "react";
import semver from "semver";
import _ from "underscore";
import { cleanSemverVersion } from "../../constants/helpers";
import { instancesStore } from "../../stores/Stores";
import Label from '../Common/Label';
import StatusHistoryContainer from "./StatusHistoryContainer.react";

const TableLabel = styled(Label)({
  lineHeight: '45px',
});

function Item(props) {
  let date = moment.utc(props.instance.application.last_check_for_updates).local().format('DD/MM/YYYY, hh:mma');
  let downloadingIcon = props.instance.statusInfo.spinning ? <img src='img/mini_loading.gif' /> : '';
  let statusIcon = props.instance.statusInfo.icon ? <i className={props.instance.statusInfo.icon}></i> : '';
  let instanceLabel = props.instance.statusInfo.className ? <TableLabel>{statusIcon} {downloadingIcon} {props.instance.statusInfo.description}</TableLabel> : <div>&nbsp;</div>;
  let version = cleanSemverVersion(props.instance.application.version);
  let currentVersionIndex = props.lastVersionChannel ? _.indexOf(props.versionNumbers, props.lastVersionChannel) : null;
  let versionStyle = 'default';

  function fetchStatusHistoryFromStore() {
    let appID = props.instance.application.application_id;
    let groupID = props.instance.application.group_id;
    let instanceID = props.instance.id;
    const selected = props.selected;

    if (!selected) {
      instancesStore.getInstanceStatusHistory(appID, groupID, instanceID)
        .done(() => {
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
          {props.instance.id}
        </TableCell>
        <TableCell>
          {instanceLabel}
        </TableCell>
        <TableCell>
          <span className={"box--" + versionStyle}>{version}</span>
        </TableCell>
        <TableCell>
          {date}
        </TableCell>
      </TableRow>
      <TableRow>
        <TableCell padding="none" colSpan={5}>
          <Collapse
            hidden={!props.selected}
            in={props.selected}
          >
            <StatusHistoryContainer instance={props.instance} key={props.instance.id} />
          </Collapse>
        </TableCell>
      </TableRow>
    </React.Fragment>
  );
}

Item.propTypes = {
  instance: PropTypes.object.isRequired,
  key: PropTypes.number.isRequired,
  selected: PropTypes.bool,
  versionNumbers: PropTypes.array,
  lastVersionChannel: PropTypes.string
}

export default Item
