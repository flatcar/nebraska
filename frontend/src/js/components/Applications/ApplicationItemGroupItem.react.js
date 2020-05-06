import Link from '@material-ui/core/Link';
import { makeStyles } from '@material-ui/core/styles';
import PropTypes from 'prop-types';
import React from 'react';
import { Link as RouterLink } from 'react-router-dom';
import API from '../../api/API';

const useStyles = makeStyles(theme => ({
  groupLink: {
    fontSize: '1rem'
  },
}));

function ApplicationItemGroupItem(props) {
  const classes = useStyles();
  const {group} = props;
  const [totalInstances, setTotalInstances] = React.useState(-1);

  React.useEffect(() => {
    // We use this function without any filter to get the total number of instances
    // in the group.
    API.getInstancesCount(group.application_id, group.id, '1d')
      .then(result => {
        setTotalInstances(result);
      })
      .catch(err => console.error('Error loading total instances in Instances/List', err));
  },
  [group]);

  return (
    <Link
      className={classes.groupLink}
      to={{pathname: `/apps/${props.group.application_id}/groups/${props.group.id}`}}
      component={RouterLink}
    >
      {props.group.name} {(totalInstances > 0) && `(${totalInstances})`}
    </Link>
  );
}

ApplicationItemGroupItem.propTypes = {
  group: PropTypes.object.isRequired,
  appName: PropTypes.string.isRequired
};

export default ApplicationItemGroupItem;
