import { Box } from '@material-ui/core';
import MuiList from '@material-ui/core/List';
import { makeStyles } from '@material-ui/core/styles';
import Typography from '@material-ui/core/Typography';
import PropTypes from 'prop-types';
import React from 'react';
import { makeLocaleTime } from '../../constants/helpers';
import Item from './Item.react';

const useStyles = makeStyles(theme => ({
  listTitle: {
    fontSize: '1em',
  },
}));

function List(props) {
  const classes = useStyles();
  const entries = props.entries ? props.entries : [];

  return (
    <React.Fragment>
      <Typography className={classes.listTitle}>
        <Box padding="1em">
          {makeLocaleTime(props.timestamp, {
            showTime: false,
            dateFormat: {weekday: 'long', month: 'long', day: 'numeric', year: 'numeric'}
          })}
        </Box>
      </Typography>
      <MuiList>
        {entries.map((entry, i) =>
          <Item key={i} entry={entry} />
        )}
      </MuiList>
    </React.Fragment>
  );
}

List.propTypes = {
  timestamp: PropTypes.string.isRequired,
  entries: PropTypes.array.isRequired
};

export default List;
