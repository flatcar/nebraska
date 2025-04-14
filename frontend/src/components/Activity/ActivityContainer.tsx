import Box from '@mui/material/Box';
import Grid from '@mui/material/Grid';
import Paper from '@mui/material/Paper';
import TablePagination from '@mui/material/TablePagination';
import makeStyles from '@mui/styles/makeStyles';
import React from 'react';
import { Trans, useTranslation } from 'react-i18next';
import _ from 'underscore';

import { Activity } from '../../api/apiDataTypes';
import { activityStore } from '../../stores/Stores';
import Empty from '../common/EmptyContent';
import ListHeader from '../common/ListHeader';
import Loader from '../common/Loader';
import ActivityList from './ActivityList';

const useStyles = makeStyles({
  toolbar: {
    padding: 0,
  },
  select: {
    fontSize: '.85em',
  },
  selectLabel: {
    fontSize: '.85em',
  },
  displayedRows: {
    fontSize: '.85em',
  },
});

export type ActivityContainerProps = any;

function Container() {
  const classes = useStyles();
  const { t } = useTranslation();

  const [activity, setActivity] = React.useState(getActivityEntries());
  const rowsOptions = [5, 10, 50];
  const [page, setPage] = React.useState(0);
  const [rowsPerPage, setRowsPerPage] = React.useState(rowsOptions[0]);

  React.useEffect(() => {
    activityStore().addChangeListener(onChange);

    return function cleanup() {
      activityStore().removeChangeListener(onChange);
    };
  }, [activity]);

  function onChange() {
    setActivity(getActivityEntries());
    setPage(0);
  }

  function handleChangePage(
    _: React.MouseEvent<HTMLButtonElement, MouseEvent> | null,
    newPage: number
  ) {
    setPage(newPage);
  }

  function handleChangeRowsPerPage(
    event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>
  ) {
    setRowsPerPage(+event.target.value);
    setPage(0);
  }

  function getPagedActivity() {
    const entriesPerTime: { [key: string]: Activity[] } = {};
    let timestamp = null;
    if (!activity) {
      return entriesPerTime;
    }

    for (
      let i = page * rowsPerPage;
      i < Math.min(activity.length, page * rowsPerPage + rowsPerPage);
      ++i
    ) {
      const entry = activity[i];
      const date = new Date(entry.created_ts);
      if (!timestamp || date.getDay() !== new Date(timestamp).getDay()) {
        timestamp = date.toUTCString();
        entriesPerTime[timestamp] = [];
      }

      entriesPerTime[timestamp] = entriesPerTime[timestamp].concat(entry);
    }
    return entriesPerTime;
  }

  function getActivityEntries() {
    const activityObj = activityStore().getCachedActivity();
    if (_.isNull(activityObj)) {
      return null;
    }

    let entries: Activity[] = [];

    Object.values(activityObj).forEach(value => {
      entries = entries.concat(value);
    });

    return entries;
  }

  return (
    <>
      <ListHeader title={t('activity|activity')} />
      <Paper>
        <Box padding="1em">
          {_.isNull(activity) ? (
            <Loader />
          ) : _.isEmpty(activity) ? (
            <Empty>
              <Trans t={t} ns="activity" i18nKey="no_activity">
                No activity found for the last week.
                <br />
                <br />
                You will see here important events related to the rollout of your updates. Stay
                tuned!
              </Trans>
            </Empty>
          ) : (
            <Grid container direction="column">
              <Grid item>
                {Object.values(
                  _.mapObject(getPagedActivity(), (entries, timestamp) => {
                    return <ActivityList timestamp={timestamp} entries={entries} key={timestamp} />;
                  })
                )}
              </Grid>
              <Grid item>
                <TablePagination
                  classes={classes}
                  rowsPerPageOptions={rowsOptions}
                  component="div"
                  count={activity.length}
                  rowsPerPage={rowsPerPage}
                  page={page}
                  backIconButtonProps={{
                    'aria-label': t('activity|previous_page'),
                  }}
                  nextIconButtonProps={{
                    'aria-label': t('activity|next_page'),
                  }}
                  onPageChange={handleChangePage}
                  onRowsPerPageChange={handleChangeRowsPerPage}
                />
              </Grid>
            </Grid>
          )}
        </Box>
      </Paper>
    </>
  );
}

export default Container;
