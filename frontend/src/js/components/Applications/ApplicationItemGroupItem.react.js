import PropTypes from 'prop-types';
import React from 'react';
import Link from '@material-ui/core/Link';
import { Link as RouterLink } from 'react-router-dom';

function ApplicationItemGroupItem(props) {
  const instances_total = props.group.instances_stats.total ? '(' + props.group.instances_stats.total + ')' : '';

  return(
    <Link
      to={{pathname: `/apps/${props.group.application_id}/groups/${props.group.id}`}}
      component={RouterLink}
    >
      {props.group.name} {instances_total}
    </Link>
  )
}

ApplicationItemGroupItem.propTypes = {
  group: PropTypes.object.isRequired,
  appName: PropTypes.string.isRequired
};

export default ApplicationItemGroupItem;
