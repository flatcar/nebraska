import menuDown from '@iconify/icons-mdi/menu-down';
import menuSwap from '@iconify/icons-mdi/menu-swap';
import menuUp from '@iconify/icons-mdi/menu-up';
import Icon from '@iconify/react';
import { IconButton } from '@mui/material';
import MuiTable from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import React from 'react';
import { useTranslation } from 'react-i18next';
import semver from 'semver';
import _ from 'underscore';

import { Channel, Instance } from '../../api/apiDataTypes';
import { cleanSemverVersion, InstanceSortFilters } from '../../utils/helpers';
import Item from './Item';

function TableCellWithSortButtons(props: {
  sortQuery: string;
  clickHandler: (isDescSortOrder: boolean, sortKey: string) => void;
  children: React.ReactNode;
  isDefault: boolean;
  defaultIsDescSortOrder: boolean;
}) {
  const { sortQuery, clickHandler, defaultIsDescSortOrder, isDefault } = props;
  const [isDescSortOrder, setDescSortOrder] = React.useState(
    isDefault ? defaultIsDescSortOrder : false
  );
  return (
    <TableCell>
      {props.children}
      <IconButton
        size="small"
        onClick={() => {
          const newOrder = !isDescSortOrder;
          setDescSortOrder(newOrder);
          clickHandler(newOrder, sortQuery);
        }}
      >
        <Icon icon={!isDefault ? menuSwap : isDescSortOrder ? menuDown : menuUp} />
      </IconButton>
    </TableCell>
  );
}

function Table(props: {
  version_breakdown?: any;
  channel: Channel;
  instances: Instance[];
  sortQuery: string;
  isDescSortOrder: boolean;
  sortHandler: (isDescSortOrder: boolean, sortKey: string) => void;
}) {
  const { sortHandler, sortQuery, isDescSortOrder } = props;
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
            defaultIsDescSortOrder={isDescSortOrder}
          >
            {t('instances|instance')}
          </TableCellWithSortButtons>
          <TableCellWithSortButtons
            clickHandler={sortHandler}
            sortQuery={InstanceSortFilters['ip']}
            defaultIsDescSortOrder={isDescSortOrder}
            isDefault={sortQuery === InstanceSortFilters['ip']}
          >
            {t('instances|ip')}
          </TableCellWithSortButtons>
          <TableCell>{t('instances|current_status')}</TableCell>
          <TableCell>{t('instances|version')}</TableCell>
          <TableCellWithSortButtons
            clickHandler={sortHandler}
            sortQuery={InstanceSortFilters['last-check']}
            defaultIsDescSortOrder={isDescSortOrder}
            isDefault={sortQuery === InstanceSortFilters['last-check']}
          >
            {t('instances|last_check')}
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
