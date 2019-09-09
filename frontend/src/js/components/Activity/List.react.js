import Grid from '@material-ui/core/Grid';
import MuiList from '@material-ui/core/List';
import { makeStyles } from '@material-ui/core/styles';
import Typography from '@material-ui/core/Typography';
import PropTypes from 'prop-types';
import React from 'react';
import Item from './Item.react';

const useStyles = makeStyles(theme => ({
  listTitle: {
    fontSize: '1em',
  },
}));

function List(props) {
  const classes = useStyles();
  let entries = props.entries ? props.entries : []

  return(
    <Grid
      container
      alignItems="stretch"
      direction="column"
    >
      <Typography className={classes.listTitle}>
        {props.day}
      </Typography>
      <MuiList>
        {entries.map((entry, i) =>
          <Item key={i} entry={entry} />
        )}
      </MuiList>
    </Grid>
  );
}

List.propTypes = {
  day: PropTypes.string.isRequired,
  entries: PropTypes.array.isRequired
}

export default List
