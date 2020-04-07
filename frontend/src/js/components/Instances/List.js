import Box from '@material-ui/core/Box';
import Button from '@material-ui/core/Button';
import FormControl from '@material-ui/core/FormControl';
import Grid from '@material-ui/core/Grid';
import Input from '@material-ui/core/Input';
import InputLabel from '@material-ui/core/InputLabel';
import MenuItem from '@material-ui/core/MenuItem';
import Paper from '@material-ui/core/Paper';
import Select from '@material-ui/core/Select';
import TablePagination from '@material-ui/core/TablePagination';
import { useTheme } from '@material-ui/styles';
import PropTypes from 'prop-types';
import React from 'react';
import { cleanSemverVersion, getMinuteDifference } from '../../constants/helpers';
import { instancesStore } from '../../stores/Stores';
import Empty from '../Common/EmptyContent';
import ListHeader from '../Common/ListHeader';
import Loader from '../Common/Loader';
import { InstanceCountLabel } from './Common';
import makeStatusDefs from './StatusDefs';
import Table from './Table';

const CHECKS_TIMEOUT = 60; // secs

function InstanceFilter(props) {
  const statusDefs = makeStatusDefs(useTheme());
  const {onFiltersChanged, versions} = props;

  function changeFilter(filterName, filterValue) {
    if (filterValue === props.filter[filterName]) {
      return;
    }

    const filter = props.filter;
    filter[filterName] = filterValue;

    onFiltersChanged(filter);
  }

  return (
    <Grid container spacing={2}>
      <Grid item xs={6}>
        <FormControl
          fullWidth
          disabled={props.disabled}
        >
          <InputLabel htmlFor="select-status" shrink>Filter Status</InputLabel>
          <Select
            onChange={event => changeFilter('status', event.target.value) }
            input={<Input id="select-status" />}
            renderValue={selected =>
              selected ? statusDefs[selected].label : 'Show All'
            }
            value={props.filter.status}
            displayEmpty
          >
            <MenuItem key="" value="">Show All</MenuItem>
            {
              Object.keys(statusDefs).map(statusType => {
                const label = statusDefs[statusType].label;
                return <MenuItem key={statusType} value={statusType}>{label}</MenuItem>;
              })
            }
          </Select>
        </FormControl>
      </Grid>
      <Grid item xs={6}>
        <FormControl
          fullWidth
          disabled={props.disabled}
        >
          <InputLabel htmlFor="select-versions" shrink>Filter Version</InputLabel>
          <Select
            onChange={event => changeFilter('version', event.target.value) }
            input={<Input id="select-versions" />}
            renderValue={selected =>
              selected ? selected : 'Show All'
            }
            value={props.filter.version}
            displayEmpty
          >
            <MenuItem key="" value="">Show All</MenuItem>
            {
              (versions || []).map(({version}) => {
                return <MenuItem key={version} value={version}>{version}</MenuItem>;
              })
            }
          </Select>
        </FormControl>
      </Grid>
    </Grid>
  );
}

function ListView(props) {
  const {application, group} = props;
  const [page, setPage] = React.useState(0);
  const [rowsPerPage, setRowsPerPage] = React.useState(10);
  const [instances, setInstances] = React.useState(null);
  const [filteredInstances, setFilteredInstances] = React.useState([]);
  const [lastCheck, setLastCheck] = React.useState(new Date(0, 0)); // Long long time ago.
  const [filters, setFilters] = React.useState({status: '', version: ''});

  function handleChangePage(event, newPage) {
    setPage(newPage);
  }

  function handleChangeRowsPerPage(event) {
    setRowsPerPage(+event.target.value);
    setPage(0);
  }

  function onFiltersChanged(newFilters) {
    applyFilters(newFilters);
  }

  function applyFilters(_filters = {}) {
    const newFilters = _filters || {status: '', version: ''};
    setFilters(newFilters);

    const filterInstances = instances.filter(instance => {
      if (newFilters.version &&
          newFilters.version !== cleanSemverVersion(instance.application.version)) {
        return false;
      }
      if (newFilters.status &&
          instance.statusInfo.type !== newFilters.status) {
        return false;
      }
      return true;
    });

    setPage(0);
    setFilteredInstances(filterInstances);
  }

  function resetFilters() {
    applyFilters();
  }

  function onChangeInstances() {
    const cachedInstances = instancesStore.getCachedInstances(application.id, group.id) || [];
    if (!instances || instances.length === 0 || filteredInstances.length === 0) {
      setInstances(cachedInstances);
      setFilteredInstances(cachedInstances);
      return;
    }

    if (cachedInstances.length > 0 && instances && instances.length > 0) {
      // Update instances state only when needed.
      if (cachedInstances.length !== instances.length ||
          cachedInstances[0].id !== instances[0].id ||
          cachedInstances[cachedInstances.length - 1].id !== instances[instances.length - 1].id) {
        // @todo: Allow to manually refresh list from UI.
        setInstances(cachedInstances);
      }
    }
  }

  React.useEffect(() => {
    instancesStore.addChangeListener(onChangeInstances);
    // @todo: This check avoids multiple unnecessary fetches, but we should
    // use a smarter refresh in the background that updates the list when needed.
    const now = new Date();
    // get seconds difference of now and lastCheck
    if (getMinuteDifference(now, lastCheck) / 60 > CHECKS_TIMEOUT) {
      setLastCheck(now);
      instancesStore.getInstances(application.id, group.id, null,
        {
          perpage: group.instances_stats.total
        });
    }

    return function cleanup() {
      instancesStore.removeChangeListener(onChangeInstances);
    };
  },
  [lastCheck, instances, filteredInstances]);

  function getInstanceCount() {
    if (!instances || instances.length === 0)
      return group.instances_stats.total;
    if (filteredInstances.length === instances.length)
      return filteredInstances.length;
    return `${filteredInstances.length}/${instances.length}`;
  }

  function isFiltered() {
    return filters.status || filters.version;
  }

  return (
    <Paper>
      <ListHeader
        title="Instance List"
      />
      <Box padding="1em">
        <Grid
          container
          spacing={1}
        >
          <Grid
            item
            container
            md={12}
            alignItems="stretch"
          >
            <Grid item md>
              <InstanceCountLabel
                countText={getInstanceCount()}
              />
            </Grid>
            <Grid item md>
              <InstanceFilter
                versions={group.version_breakdown}
                onFiltersChanged={onFiltersChanged}
                filter={filters}
                disabled={!instances || instances.length === 0}
              />
            </Grid>
          </Grid>
          {isFiltered() &&
            <Grid item md={12} container justify="center">
              <Grid item>
                <Button
                  variant="outlined"
                  color="secondary"
                  onClick={resetFilters}
                >
                  Reset filters
                </Button>
              </Grid>
            </Grid>
          }
          <Grid item md={12}>
            { instances ?
              (filteredInstances.length > 0 ?
                <React.Fragment>
                  <Table
                    group={group}
                    channel={group.channel}
                    instances={
                      filteredInstances
                        .slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage)
                    }
                  />
                  <TablePagination
                    rowsPerPageOptions={[10, 25, 50, 100]}
                    component="div"
                    count={filteredInstances.length}
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
                </React.Fragment>
                :
                <Empty>No instances.</Empty>
              )
              :
                <Loader />
            }
          </Grid>
        </Grid>
      </Box>
    </Paper>
  );
}

ListView.propTypes = {
  application: PropTypes.object.isRequired,
  group: PropTypes.object.isRequired,
};

export default ListView;
