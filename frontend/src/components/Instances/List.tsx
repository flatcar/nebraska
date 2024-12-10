import infoIcon from '@iconify/icons-mdi/information-circle-outline';
import searchIcon from '@iconify/icons-mdi/search';
import Icon from '@iconify/react';
import { makeStyles, Theme } from '@material-ui/core';
import Box from '@material-ui/core/Box';
import Button from '@material-ui/core/Button';
import FormControl from '@material-ui/core/FormControl';
import Grid from '@material-ui/core/Grid';
import IconButton from '@material-ui/core/IconButton';
import Input from '@material-ui/core/Input';
import InputAdornment from '@material-ui/core/InputAdornment';
import InputLabel from '@material-ui/core/InputLabel';
import MenuItem from '@material-ui/core/MenuItem';
import Paper from '@material-ui/core/Paper';
import Select from '@material-ui/core/Select';
import TablePagination from '@material-ui/core/TablePagination';
import { useTheme } from '@material-ui/styles';
import PropTypes from 'prop-types';
import React, { ChangeEvent } from 'react';
import { useTranslation } from 'react-i18next';
import { useHistory, useLocation } from 'react-router-dom';
import _ from 'underscore';
import API from '../../api/API';
import { Application, Group, Instance, Instances } from '../../api/apiDataTypes';
import {
  getInstanceStatus,
  getKeyByValue,
  InstanceSortFilters,
  SearchFilterClassifiers,
  useGroupVersionBreakdown,
} from '../../utils/helpers';
import Empty from '../common/EmptyContent';
import LightTooltip from '../common/LightTooltip';
import ListHeader from '../common/ListHeader';
import SearchInput from '../common/ListSearch';
import Loader from '../common/Loader';
import TimeIntervalLinks from '../common/TimeIntervalLinks';
import { InstanceCountLabel } from './Common';
import makeStatusDefs from './StatusDefs';
import Table from './Table';

// The indexes for the sorting names match the backend index for that criteria as well.
const SORT_ORDERS = ['asc', 'desc'];

const useStyles = makeStyles(theme => ({
  root: {
    backgroundColor: theme.palette.lightSilverShade,
  },
}));

interface InstanceFilterProps {
  versions: any[];
  onFiltersChanged: (newFilters: any) => void;
  filter: {
    [key: string]: any;
  };
  disabled?: boolean;
}

function InstanceFilter(props: InstanceFilterProps) {
  const statusDefs = makeStatusDefs(useTheme());
  const { t } = useTranslation();
  const { onFiltersChanged, versions } = props;

  function changeFilter(filterName: string, filterValue: string) {
    if (filterValue === props.filter[filterName]) {
      return;
    }

    const filter = { ...props.filter };
    filter[filterName] = filterValue;

    onFiltersChanged(filter);
  }

  return (
    <Box pr={2}>
      <Grid container spacing={2} justify="flex-end">
        <Grid item xs={5}>
          <FormControl fullWidth disabled={props.disabled}>
            <InputLabel htmlFor="select-status" shrink>
              {t('instances|Filter Status')}
            </InputLabel>
            <Select
              onChange={(event: any) => changeFilter('status', event.target.value)}
              input={<Input id="select-status" />}
              renderValue={(selected: any) =>
                selected ? statusDefs[selected].label : t('instances|Show All')
              }
              value={props.filter.status}
              displayEmpty
            >
              <MenuItem key="" value="">
                {t('instances|Show All')}
              </MenuItem>
              {Object.keys(statusDefs).map(statusType => {
                const label = statusDefs[statusType].label;
                return (
                  <MenuItem key={statusType} value={statusType}>
                    {label}
                  </MenuItem>
                );
              })}
            </Select>
          </FormControl>
        </Grid>
        <Grid item xs={5}>
          <FormControl fullWidth disabled={props.disabled}>
            <InputLabel htmlFor="select-versions" shrink>
              {t('instances|Filter Version')}
            </InputLabel>
            <Select
              onChange={(event: ChangeEvent<{ name?: string | undefined; value: any }>) =>
                changeFilter('version', event.target.value)
              }
              input={<Input id="select-versions" />}
              renderValue={(selected: any) => (selected ? selected : t('instances|Show All'))}
              value={props.filter.version}
              displayEmpty
            >
              <MenuItem key="" value="">
                {t('instances|Show All')}
              </MenuItem>
              {(versions || []).map(({ version }) => {
                return (
                  <MenuItem key={version} value={version}>
                    {version}
                  </MenuItem>
                );
              })}
            </Select>
          </FormControl>
        </Grid>
      </Grid>
    </Box>
  );
}

function ListView(props: { application: Application; group: Group }) {
  const classes = useStyles();
  const theme = useTheme();
  const { t } = useTranslation();
  const statusDefs = makeStatusDefs(useTheme());
  const { application, group } = props;
  const versionBreakdown = useGroupVersionBreakdown(group);
  /*TODO: use the URL as the single source of truth and remove states */
  const [page, setPage] = React.useState(0);
  const [rowsPerPage, setRowsPerPage] = React.useState(10);
  const [isDescSortOrder, setIsDescSortOrder] = React.useState(false);
  const [sortQuery, setSortQuery] = React.useState(InstanceSortFilters['last-check']);
  const [filters, setFilters] = React.useState<{ [key: string]: any }>({
    status: '',
    version: '',
    sortOrder: SORT_ORDERS[1],
  });
  const [instancesObj, setInstancesObj] = React.useState<Instances>({
    instances: [],
    total: -1,
  });
  const [instanceFetchLoading, setInstanceFetchLoading] = React.useState(false);
  const [totalInstances, setTotalInstances] = React.useState(-1);
  const [searchObject, setSearchObject] = React.useState<{
    searchFilter?: string;
    searchValue?: string;
  }>({});
  const location = useLocation();
  const history = useHistory();

  function getDuration() {
    return new URLSearchParams(location.search).get('period') || '1d';
  }

  function fetchFiltersFromURL(callback: (...args: any) => void) {
    let status = '';
    const queryParams = new URLSearchParams(location.search);
    if (queryParams.has('status')) {
      const statusValue = queryParams.get('status');
      if (statusValue !== 'ShowAll') {
        for (const key in statusDefs) {
          if (statusDefs[key].label === statusValue) {
            status = key;
            break;
          }
        }
      }
    }
    const version = queryParams.get('version') || '';
    const sort = InstanceSortFilters[queryParams.get('sort') || 'last-check'];
    const pageFromURL = queryParams.get('page');
    const pageQueryParam = ((pageFromURL && parseInt(pageFromURL)) || 1) - 1;
    const perPage = parseInt(queryParams.get('perPage') as string) || 10;
    const sortOrder = SORT_ORDERS[1] === (queryParams.get('sortOrder') as string) ? 1 : 0;
    const duration = getDuration();

    callback(status, version, sort, sortOrder, pageQueryParam, perPage, duration);
  }

  function addQuery(queryObj: { [key: string]: any }) {
    const pathname = location.pathname;
    const searchParams = new URLSearchParams(location.search);
    for (const key in queryObj) {
      const value = queryObj[key];
      if (value) {
        searchParams.set(key, value);
      } else {
        searchParams.delete(key);
      }
    }

    history.push({
      pathname: pathname,
      search: searchParams.toString(),
    });
  }

  function fetchInstances(
    filters: { [key: string]: any },
    page: number,
    perPage: number,
    duration: string,
    searchObject: { searchFilter?: string; searchValue?: string }
  ) {
    setInstanceFetchLoading(true);
    const fetchFilters = { ...filters };
    if (filters.status === '') {
      fetchFilters.status = '0';
    } else {
      const statusDefinition = statusDefs[fetchFilters.status];
      fetchFilters.status = statusDefinition.queryValue;
    }
    API.getInstances(application.id, group.id, {
      ...fetchFilters,
      sortOrder: Number(isDescSortOrder),
      page: page + 1,
      perpage: perPage,
      duration,
      ...searchObject,
    })
      .then(result => {
        setInstanceFetchLoading(false);
        // Since we have retrieved the instances without a filter (i.e. all instances)
        // we update the total.
        if (!fetchFilters.status && !fetchFilters.version) {
          setTotalInstances(result.total);
        }
        if (result.instances) {
          const massagedInstances = result.instances.map((instance: Instance) => {
            instance.statusInfo = getInstanceStatus(instance.application.status);
            return instance;
          });
          setInstancesObj({ instances: massagedInstances, total: result.total });
        } else {
          setInstancesObj({ instances: [], total: result.total });
        }
      })
      .catch(() => {
        setInstanceFetchLoading(false);
      });
  }

  function handleChangePage(
    event: React.MouseEvent<HTMLButtonElement, MouseEvent> | null,
    newPage: number
  ) {
    addQuery({ page: newPage + 1 });
  }

  function handleChangeRowsPerPage(event: React.ChangeEvent<{ value: string }>) {
    addQuery({ page: 1, perPage: +event.target.value });
  }

  function onFiltersChanged(newFilters: { [key: string]: any }) {
    applyFilters(newFilters);
  }

  function applyFilters(_filters = {}) {
    const newFilters: { [key: string]: any } =
      Object.keys(_filters).length !== 0 ? _filters : { status: '', version: '' };
    const statusQueryParam = newFilters.status ? statusDefs[newFilters.status].label : '';
    addQuery({ status: statusQueryParam, version: newFilters.version });
    setFilters(newFilters);
  }

  function resetFilters() {
    applyFilters();
  }

  function handleInstanceFetch(searchObject: {
    searchFilter?: string | undefined;
    searchValue?: string | undefined;
  }) {
    fetchFiltersFromURL(
      (
        status: string,
        version: string,
        sort: string,
        isDescSortOrder: boolean,
        pageParam: number,
        perPageParam: number,
        duration: string
      ) => {
        setFilters({ status, version, sort });
        setPage(pageParam);
        setIsDescSortOrder(isDescSortOrder);
        setSortQuery(sort);
        setRowsPerPage(perPageParam);
        fetchInstances({ status, version, sort }, pageParam, perPageParam, duration, searchObject);
      }
    );
  }
  React.useEffect(() => {
    handleInstanceFetch(searchObject);
  }, [location]);

  React.useEffect(() => {
    // We want to run it only if the searchValue is empty and once for change in totalInstances.
    if (totalInstances > 0 && searchObject.searchValue) {
      return;
    }

    // We use this function without any filter to get the total number of instances
    // in the group.
    const queryParams = new URLSearchParams(window.location.search);
    const duration = queryParams.get('period');
    API.getInstancesCount(application.id, group.id, duration as string)
      .then(result => {
        setTotalInstances(result);
      })
      .catch(err => console.error('Error loading total instances in Instances/List', err));
  }, [totalInstances, searchObject]);

  function getInstanceCount() {
    const total = totalInstances > -1 ? totalInstances : '…';
    const instancesTotal = instancesObj.total > -1 ? instancesObj.total : '...';
    if (
      (!searchObject.searchValue && !filters.status && !filters.version) ||
      instancesTotal === total
    ) {
      return total;
    }
    return `${instancesTotal}/${total}`;
  }

  function isFiltered() {
    return filters.status || filters.version;
  }

  function sortHandler(isDescSortOrder: boolean, sortQuery: string) {
    setIsDescSortOrder(isDescSortOrder);
    setSortQuery(sortQuery);
    const sortAliasKey = getKeyByValue(InstanceSortFilters, sortQuery);
    addQuery({ sort: sortAliasKey, sortOrder: SORT_ORDERS[Number(isDescSortOrder)] });
  }

  function searchHandler(e: React.ChangeEvent<{ value: string }>) {
    const value = e.target.value;
    // This means user has reseted the input field, and now we need to fetch all the instances
    if (value === '') {
      handleInstanceFetch({});
    }
    // handle if a classifier is present
    const [classifierName] = value.split(':');
    const classifierVal = value.substring(value.indexOf(':') + 1);
    const classifiedFilter = SearchFilterClassifiers.find(
      classifier => classifier.name === classifierName
    );
    if (classifierVal !== undefined && classifiedFilter) {
      setSearchObject({
        searchValue: classifierVal,
        searchFilter: classifiedFilter.queryValue,
      });
      return;
    }
    if (!classifiedFilter) {
      // this means user is trying to search without any classifier
      setSearchObject({
        searchFilter: 'All',
        searchValue: value,
      });
    }
  }

  function handleSearchSubmit(e: any) {
    if (e.key === 'Enter' && Object.keys(searchObject).length !== 0) {
      const debouncedFetch = _.debounce(handleInstanceFetch, 100);
      debouncedFetch(searchObject);
    }
  }

  function getSearchTooltipText() {
    return t(`instances|You can search by typing and pressing enter.
     The search will show matches for the instances id, alias, and ip fields, in this order. 
     It is also possible to match only one field by using its classifier, for example: 
     id:0001
     alias:"My instance" 
     ip:256.0.0.1`);
  }

  const searchInputRef = React.createRef<HTMLInputElement>();

  return (
    <>
      <ListHeader title={t('instances|Instance List')} />
      <Paper>
        <Box padding="1em">
          <Grid container spacing={1}>
            <Grid item container justify="space-between" alignItems="stretch">
              <Grid item>
                <Box
                  mb={2}
                  color={(theme as Theme).palette.greyShadeColor}
                  fontSize={30}
                  fontWeight={700}
                >
                  {group.name}
                </Box>
              </Grid>
              <Grid item>
                <InputLabel htmlFor="instance-search-filter" shrink>
                  {t('frequent|Search')}
                </InputLabel>
                <SearchInput
                  id="instance-search-filter"
                  startAdornment={
                    <InputAdornment position="start">
                      <IconButton
                        onClick={() => searchInputRef.current?.focus()}
                        title="Search Icon"
                      >
                        <Icon icon={searchIcon} width="15" height="15" />
                      </IconButton>
                    </InputAdornment>
                  }
                  endAdornment={
                    <InputAdornment position="end">
                      <LightTooltip title={getSearchTooltipText()}>
                        <IconButton>
                          <Icon icon={infoIcon} width="20" height="20" />
                        </IconButton>
                      </LightTooltip>
                    </InputAdornment>
                  }
                  onChange={searchHandler}
                  onKeyPress={handleSearchSubmit}
                  inputRef={searchInputRef}
                  ariaLabel="Search"
                />
              </Grid>
              <Grid item>
                <TimeIntervalLinks
                  intervalChangeHandler={duration => addQuery({ period: duration.queryValue })}
                  selectedInterval={getDuration()}
                  appID={application.id}
                  groupID={group.id}
                />
              </Grid>
            </Grid>
            <Box width="100%" borderTop={1} borderColor={'#E0E0E0'} className={classes.root}>
              <Grid item container md={12} alignItems="stretch" justify="space-between">
                <Grid item md>
                  <Box display="flex" alignItems="center">
                    <Box ml={2}>
                      <InstanceCountLabel countText={getInstanceCount()} instanceListView />
                    </Box>
                  </Box>
                </Grid>
                <Grid item md>
                  <Box mt={2}>
                    <InstanceFilter
                      versions={versionBreakdown}
                      onFiltersChanged={onFiltersChanged}
                      filter={filters}
                    />
                  </Box>
                </Grid>
              </Grid>
            </Box>
            {isFiltered() && (
              <Grid item md={12} container justify="center">
                <Grid item>
                  <Button variant="outlined" color="secondary" onClick={resetFilters}>
                    {t('instances|Reset filters')}
                  </Button>
                </Grid>
              </Grid>
            )}
            <Grid item md={12}>
              {!instanceFetchLoading ? (
                instancesObj.instances.length > 0 ? (
                  <React.Fragment>
                    <Table
                      channel={group.channel}
                      instances={instancesObj.instances}
                      isDescSortOrder={isDescSortOrder}
                      sortQuery={sortQuery}
                      sortHandler={sortHandler}
                    />
                    <TablePagination
                      rowsPerPageOptions={[10, 25, 50, 100]}
                      component="div"
                      count={instancesObj.total}
                      rowsPerPage={rowsPerPage}
                      page={page}
                      backIconButtonProps={{
                        'aria-label': t('frequent|previous page'),
                      }}
                      nextIconButtonProps={{
                        'aria-label': t('frequent|next page'),
                      }}
                      onPageChange={handleChangePage}
                      onRowsPerPageChange={handleChangeRowsPerPage}
                    />
                  </React.Fragment>
                ) : (
                  <Empty>{t('frequent|No instances.')}</Empty>
                )
              ) : (
                <Loader />
              )}
            </Grid>
          </Grid>
        </Box>
      </Paper>
    </>
  );
}

ListView.propTypes = {
  application: PropTypes.object.isRequired,
  group: PropTypes.object.isRequired,
};

export default ListView;
