import { Icon } from '@iconify/react';
import { TableContainer, Theme } from '@mui/material';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import FormControl from '@mui/material/FormControl';
import Grid from '@mui/material/Grid';
import IconButton from '@mui/material/IconButton';
import Input from '@mui/material/Input';
import InputAdornment from '@mui/material/InputAdornment';
import InputLabel from '@mui/material/InputLabel';
import MenuItem from '@mui/material/MenuItem';
import Paper from '@mui/material/Paper';
import Select, { SelectChangeEvent } from '@mui/material/Select';
import { styled } from '@mui/material/styles';
import { useTheme } from '@mui/material/styles';
import TablePagination from '@mui/material/TablePagination';
import React from 'react';
import { Trans, useTranslation } from 'react-i18next';
import { useLocation, useNavigate } from 'react-router';


import API from '../../api/API';
import { Application, Group, Instance, Instances } from '../../api/apiDataTypes';
import infoIcon from '../../icons/mdi/information-circle-outline.json';
import searchIcon from '../../icons/mdi/search.json';
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

const PREFIX = 'ListView';

const classes = {
  mainTable: `${PREFIX}-mainTable`,
};

const Root = styled('div')(({ theme }) => ({
  [`& .${classes.mainTable}`]: {
    backgroundColor: theme.palette.lightSilverShade,
  },
}));

// The indexes for the sorting names match the backend index for that criteria as well.
const SORT_ORDERS = ['asc', 'desc'];

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
      <Grid container spacing={2} justifyContent="flex-end">
        <Grid size={5}>
          <FormControl fullWidth disabled={props.disabled}>
            <InputLabel variant="standard" htmlFor="select-status" shrink>
              {t('instances|filter_status')}
            </InputLabel>
            <Select
              onChange={(event: any) => changeFilter('status', event.target.value)}
              input={<Input id="select-status" />}
              renderValue={(selected: any) =>
                selected ? statusDefs[selected].label : t('instances|show_all')
              }
              value={props.filter.status}
              displayEmpty
            >
              <MenuItem key="" value="">
                {t('instances|show_all')}
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
        <Grid size={5}>
          <FormControl fullWidth disabled={props.disabled}>
            <InputLabel variant="standard" htmlFor="select-versions" shrink>
              {t('instances|filter_version')}
            </InputLabel>
            <Select
              onChange={(event: SelectChangeEvent<string>) =>
                changeFilter('version', event.target.value)
              }
              input={<Input id="select-versions" />}
              renderValue={(selected: any) => (selected ? selected : t('instances|show_all'))}
              value={props.filter.version}
              displayEmpty
            >
              <MenuItem key="" value="">
                {t('instances|show_all')}
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

interface ListViewProps {
  application: Application;
  group: Group;
}

function ListView(props: ListViewProps) {
  const theme = useTheme();
  const { t } = useTranslation();
  const statusDefs = makeStatusDefs(useTheme());
  const { application, group } = props;
  const versionBreakdown = useGroupVersionBreakdown(group);
  const location = useLocation();
  const navigate = useNavigate();
  const queryParams = new URLSearchParams(location.search);

  const page = Math.max(0, parseInt(queryParams.get('page') || '1', 10) - 1);
  const rowsPerPage = parseInt(queryParams.get('perPage') || '10', 10);
  const isDescSortOrder = queryParams.get('sortOrder') === SORT_ORDERS[1];
  const sortQuery = InstanceSortFilters[queryParams.get('sort') || 'last-check'];
  const duration = queryParams.get('period') || '1d';

  let status = '';
  const statusValue = queryParams.get('status');
  if (statusValue && statusValue !== 'ShowAll') {
    for (const key in statusDefs) {
      if (statusDefs[key].label === statusValue) {
        status = key;
        break;
      }
    }
  }
  const version = queryParams.get('version') || '';
  const filters = {
    status,
    version,
    sortOrder: isDescSortOrder ? SORT_ORDERS[1] : SORT_ORDERS[0],
  };

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

  function getDuration() {
    return duration;
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

    navigate({
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
    _event: React.MouseEvent<HTMLButtonElement, MouseEvent> | null,
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
  }

  function resetFilters() {
    applyFilters();
  }

  React.useEffect(() => {
    fetchInstances(filters, page, rowsPerPage, duration, searchObject);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [location.search]);

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
    // eslint-disable-next-line react-hooks/exhaustive-deps
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

  function sortHandler(isDescSortOrderParam: boolean, sortQueryParam: string) {
    const sortAliasKey = getKeyByValue(InstanceSortFilters, sortQueryParam);
    addQuery({ sort: sortAliasKey, sortOrder: SORT_ORDERS[Number(isDescSortOrderParam)] });
  }

  function searchHandler(e: React.ChangeEvent<{ value: string }>) {
    const value = e.target.value;
    // This means user has reset the input field, and now we need to fetch all the instances
    if (value === '') {
      setSearchObject({});
      fetchInstances(filters, page, rowsPerPage, duration, {});
      return;
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
      fetchInstances(filters, page, rowsPerPage, duration, searchObject);
    }
  }

  function getSearchTooltipText() {
    return <Trans t={t} i18nKey="search_instruction" components={{ br: <br /> }} />;
  }

  const searchInputRef = React.createRef<HTMLInputElement>();

  return (
    <Root>
      <ListHeader title={t('instances|instance_list')} />
      <Paper>
        <Box padding="1em">
          <Grid container spacing={1}>
            <Grid
              container
              sx={{ width: '100%', display: 'flex', justifyContent: 'space-between' }}
            >
              <Grid>
                <Box
                  mb={2}
                  color={(theme as Theme).palette.greyShadeColor}
                  fontSize={30}
                  fontWeight={700}
                >
                  {group.name}
                </Box>
              </Grid>
              <Grid>
                <InputLabel variant="standard" htmlFor="instance-search-filter" shrink>
                  {t('frequent|search')}
                </InputLabel>
                <SearchInput
                  id="instance-search-filter"
                  startAdornment={
                    <InputAdornment position="start">
                      <IconButton
                        onClick={() => searchInputRef.current?.focus()}
                        title="Search Icon"
                        size="large"
                      >
                        <Icon icon={searchIcon} width="15" height="15" />
                      </IconButton>
                    </InputAdornment>
                  }
                  endAdornment={
                    <InputAdornment position="end">
                      <LightTooltip title={getSearchTooltipText()}>
                        <IconButton size="large">
                          <Icon icon={infoIcon} width="20" height="20" />
                        </IconButton>
                      </LightTooltip>
                    </InputAdornment>
                  }
                  onChange={searchHandler}
                  onKeyPress={handleSearchSubmit}
                  inputRef={searchInputRef}
                  aria-label="Search"
                />
              </Grid>
              <Grid>
                <TimeIntervalLinks
                  intervalChangeHandler={duration => addQuery({ period: duration.queryValue })}
                  selectedInterval={getDuration()}
                  appID={application.id}
                  groupID={group.id}
                />
              </Grid>
            </Grid>
            <Box width="100%" borderTop={1} borderColor={'#E0E0E0'} className={classes.mainTable}>
              <Grid
                container
                alignItems="stretch"
                justifyContent="space-between"
                size={{
                  md: 12,
                }}
              >
                <Grid
                  size={{
                    md: 'grow',
                  }}
                >
                  <Box display="flex" alignItems="center">
                    <Box ml={2}>
                      <InstanceCountLabel countText={getInstanceCount()} instanceListView />
                    </Box>
                  </Box>
                </Grid>
                <Grid
                  size={{
                    md: 'grow',
                  }}
                >
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
              <Grid
                container
                justifyContent="center"
                size={{
                  md: 12,
                }}
              >
                <Grid>
                  <Button variant="outlined" color="secondary" onClick={resetFilters}>
                    {t('instances|reset_filters')}
                  </Button>
                </Grid>
              </Grid>
            )}
            <Grid
              size={{
                md: 12,
              }}
            >
              {!instanceFetchLoading ? (
                instancesObj.instances.length > 0 ? (
                  <React.Fragment>
                    <TableContainer component={Paper}>
                      <Table
                        channel={group.channel}
                        instances={instancesObj.instances}
                        isDescSortOrder={isDescSortOrder}
                        sortQuery={sortQuery}
                        sortHandler={sortHandler}
                      />
                    </TableContainer>
                    <TablePagination
                      rowsPerPageOptions={[10, 25, 50, 100]}
                      component="div"
                      count={instancesObj.total}
                      rowsPerPage={rowsPerPage}
                      page={page}
                      backIconButtonProps={{
                        'aria-label': t('frequent|previous_page'),
                      }}
                      nextIconButtonProps={{
                        'aria-label': t('frequent|next_page'),
                      }}
                      onPageChange={handleChangePage}
                      onRowsPerPageChange={handleChangeRowsPerPage}
                    />
                  </React.Fragment>
                ) : (
                  <Empty>{t('frequent|no_instances')}</Empty>
                )
              ) : (
                <Loader />
              )}
            </Grid>
          </Grid>
        </Box>
      </Paper>
    </Root>
  );
}

export default ListView;
