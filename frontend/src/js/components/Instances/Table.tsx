import menuDown from '@iconify/icons-mdi/menu-down';
import menuUp from '@iconify/icons-mdi/menu-up';
import Icon from '@iconify/react';
import { IconButton } from '@material-ui/core';
import MuiTable from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableHead from '@material-ui/core/TableHead';
import TableRow from '@material-ui/core/TableRow';
import React from 'react';
import { useTranslation } from 'react-i18next';
import semver from 'semver';
import _ from 'underscore';
import { Channel, Instance } from '../../api/apiDataTypes';
import { cleanSemverVersion, InstanceSortFilters } from '../../utils/helpers';
import Item from './Item';

function TableCellWithSortButtons(props: {
  sortQuery: string;
  clickHandler: (sortOrder: boolean, sortKey: string) => void;
  children: React.ReactNode;
  isDefault: boolean;
  defaultSortOrder: boolean;
}) {
  const { sortQuery, clickHandler, defaultSortOrder, isDefault } = props;
  //false denotes a increasing sort order and true a decreasing sort order
  const [sortOrder, setSortOrder] = React.useState(isDefault ? defaultSortOrder : false);
  return (
    <TableCell>
      {props.children}
      <IconButton
        size="small"
        onClick={() => {
          setSortOrder(!sortOrder);
          clickHandler(!sortOrder, sortQuery);
        }}
      >
        <Icon icon={sortOrder ? menuUp : menuDown} />
      </IconButton>
    </TableCell>
  );
}

function Table(props: {
  version_breakdown?: any;
  channel: Channel;
  instances: Instance[];
  sortQuery: string;
  sortOrder: boolean;
  sortHandler: (sortOrder: boolean, sortKey: string) => void;
}) {
  const { sortHandler, sortQuery, sortOrder } = props;
  const [selectedInstance, setSelectedInstance] = React.useState<string | null>(null);
  const { t } = useTranslation();
  const versions = props.version_breakdown || [];
  const lastVersionChannel =
    props.channel && props.channel.package ? cleanSemverVersion(props.channel.package.version) : '';
  const versionNumbers = _.map(versions, version => {
    return cleanSemverVersion(version.version);
  }).sort(semver.rcompare);

  function onItemToggle(id: string | null) {
    if (selectedInstance !== id) {
      setSelectedInstance(id);
    } else {
      setSelectedInstance(null);
    }
  }

  return (
    <MuiTable>
      <TableHead>
        <TableRow>
          <TableCellWithSortButtons
            sortQuery={InstanceSortFilters['id']}
            clickHandler={sortHandler}
            isDefault={sortQuery === InstanceSortFilters['id']}
            defaultSortOrder={sortOrder}
          >
            {t('instances|Instance')}
          </TableCellWithSortButtons>
          <TableCellWithSortButtons
            clickHandler={sortHandler}
            sortQuery={InstanceSortFilters['ip']}
            defaultSortOrder={sortOrder}
            isDefault={sortQuery === InstanceSortFilters['ip']}
          >
            {t('instances|IP')}
          </TableCellWithSortButtons>
          <TableCell>{t('instances|Current Status')}</TableCell>
          <TableCell>{t('instances|Version')}</TableCell>
          <TableCellWithSortButtons
            clickHandler={sortHandler}
            sortQuery={InstanceSortFilters['last-check']}
            defaultSortOrder={sortOrder}
            isDefault={sortQuery === InstanceSortFilters['last-check']}
          >
            {t('instances|Last Check')}
          </TableCellWithSortButtons>
        </TableRow>
      </TableHead>
      <TableBody>
        {props.instances.map((instance, i) => (
          <Item
            key={'instance_' + i}
            instance={instance}
            lastVersionChannel={lastVersionChannel}
            versionNumbers={versionNumbers}
            selected={selectedInstance === instance.id}
            onToggle={onItemToggle}
          />
        ))}
      </TableBody>
    </MuiTable>
  );
}

export default Table;
