import { makeStyles } from '@material-ui/core/styles';
import PropTypes from 'prop-types';
import React from "react";
import _ from "underscore";
import StatusHistoryList from "./StatusHistoryList.react";
import Empty from '../Common/EmptyContent';

const useStyles = makeStyles({
  historyBox: {
    paddingLeft: '2em',
    paddingRight: '2em',
    maxHeight: '400px',
    overflow: 'auto',
  },
});

function StatusHistoryContainer(props) {
  const classes = useStyles();
  let entries = '';

  if (_.isEmpty(props.instance.statusHistory)) {
    entries = <Empty>This instance hasn’t reported any events yet in the context of this application/group.</Empty>;
  } else {
    entries = <StatusHistoryList entries={props.instance.statusHistory} />;
  }

  return(
    <div className={classes.historyBox}>
      {entries}
    </div>
  );
}

StatusHistoryContainer.propTypes = {
  key: PropTypes.string.isRequired,
  instance: PropTypes.object.isRequired
}

export default StatusHistoryContainer
