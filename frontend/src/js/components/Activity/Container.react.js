import Box from '@material-ui/core/Box';
import Grid from '@material-ui/core/Grid';
import Paper from '@material-ui/core/Paper';
import { makeStyles } from '@material-ui/core/styles';
import { useTheme } from '@material-ui/core/styles';
import TablePagination from '@material-ui/core/TablePagination';
import useMediaQuery from '@material-ui/core/useMediaQuery';
import React from 'react';
import _ from 'underscore';
import { activityStore } from '../../stores/Stores';
import Empty from '../Common/EmptyContent';
import ListHeader from '../Common/ListHeader';
import Loader from '../Common/Loader';
import List from './List.react';

const useStyles = makeStyles({
  toolbar: {
    padding: 0,
  },
  caption: {
    fontSize: '.85em',
  },
  select: {
    fontSize: '.85em',
  }
});

function Container(props) {
  const theme = useTheme();
  const isSmall = useMediaQuery(theme.breakpoints.down('sm'));
  const classes = useStyles();
  const [activity, setActivity] = React.useState(getActivityEntries());
  const rowsOptions = [5, 10, 50];
  const [page, setPage] = React.useState(0);
  const [rowsPerPage, setRowsPerPage] = React.useState(rowsOptions[0]);

  React.useEffect(() => {
    activityStore.addChangeListener(onChange);

    return function cleanup () {
      activityStore.removeChangeListener(onChange);
    };
  },
  [activity]);

  function onChange() {
    setActivity(getActivityEntries());
    setPage(0);
  }

  function handleChangePage(event, newPage) {
    setPage(newPage);
  }

  function handleChangeRowsPerPage(event) {
    setRowsPerPage(+event.target.value);
    setPage(0);
  }

  function getPagedActivity() {
    const entriesPerTime = {};
    let timestamp = null;

    for (let i = page * rowsPerPage;
      i < Math.min(activity.length, page * rowsPerPage + rowsPerPage); ++i) {
      const entry = activity[i];
      const date = new Date(entry.created_ts);
      if (!timestamp || date.getDay() !== new Date(timestamp).getDay()) {
        timestamp = date.toUTCString();;
        entriesPerTime[timestamp] = [];
      }

      entriesPerTime[timestamp] = entriesPerTime[timestamp].concat(entry);
    }
    return entriesPerTime;
  }

  function getActivityEntries() {
    const activityObj = activityStore.getCachedActivity();
    if (_.isNull(activityObj)) {
      return null;
    }

    let entries = [];

    Object.values(activityObj).forEach(value => {
      entries = entries.concat(value);
    });

    return entries;
  }

  return (
    <Paper>
      <ListHeader title="Activity" />
      { _.isNull(activity) ?
        <Loader />
        : _.isEmpty(activity) ?
          <Empty>
            No activity found for the last week.
            <br/><br/>
            You will see here important events related to the rollout of your updates. Stay tuned!
          </Empty>
          :
          <Grid
            container
            direction="column"
          >
            <Grid item xs={10} sm={12} md={12}>
              {Object.values(
                _.mapObject(getPagedActivity(), (entry, timestamp) => {
                  return <List timestamp={timestamp} entries={entry} key={timestamp} />;
                })
              )}
            </Grid>
            <Grid item>
              <Box padding={isSmall ? '0' : '1em'}>
                <TablePagination
                  classes={classes}
                  rowsPerPageOptions={rowsOptions}
                  component="div"
                  count={activity.length}
                  rowsPerPage={rowsPerPage}
                  page={page}
                  backIconButtonProps={{
                    'aria-label': 'previous page',
                  }}
                  nextIconButtonProps={{
                    'aria-label': 'next page',
                  }}
                  onChangePage={handleChangePage}
                  onChangeRowsPerPage={handleChangeRowsPerPage}
                />
              </Box>
            </Grid>

          </Grid>
      }
    </Paper>
  );
}

export default Container;
