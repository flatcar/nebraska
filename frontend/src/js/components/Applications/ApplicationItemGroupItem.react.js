import Link from '@material-ui/core/Link';
import { makeStyles } from '@material-ui/core/styles';
import PropTypes from 'prop-types';
import React from 'react';
import { Link as RouterLink } from 'react-router-dom';

const useStyles = makeStyles(theme => ({
  groupLink: {
    fontSize: '1rem'
  },
}));

function ApplicationItemGroupItem(props) {
  const classes = useStyles();
  const instances_total = props.group.instances_stats.total ? '(' + props.group.instances_stats.total + ')' : '';

  return (
    <Link
      className={classes.groupLink}
      to={{pathname: `/apps/${props.group.application_id}/groups/${props.group.id}`}}
      component={RouterLink}
    >
      {props.group.name} {instances_total}
    </Link>
  );
}

ApplicationItemGroupItem.propTypes = {
  group: PropTypes.object.isRequired,
  appName: PropTypes.string.isRequired
};

export default ApplicationItemGroupItem;
