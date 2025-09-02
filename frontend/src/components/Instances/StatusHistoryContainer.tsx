import { styled } from '@mui/material/styles';
import React from 'react';
import { useTranslation } from 'react-i18next';
import _ from 'underscore';

import { InstanceStatusHistory } from '../../api/apiDataTypes';
import Empty from '../common/EmptyContent';
import StatusHistoryList from './StatusHistoryList';

const PREFIX = 'StatusHistoryContainer';

const classes = {
  historyBox: `${PREFIX}-historyBox`,
};

const Root = styled('div')({
  [`&.${classes.historyBox}`]: {
    paddingLeft: '2em',
    paddingRight: '2em',
    maxHeight: '400px',
    overflow: 'auto',
  },
});

function StatusHistoryContainer(props: { statusHistory: InstanceStatusHistory[] }) {
  const { t } = useTranslation();
  let entries: React.ReactElement;

  if (_.isEmpty(props.statusHistory)) {
    entries = (
      <Empty>
        {t(
          'instances|This instance hasn’t reported any events yet in the context of this application/group.'
        )}
      </Empty>
    );
  } else {
    entries = <StatusHistoryList entries={props.statusHistory} />;
  }

  return <Root className={classes.historyBox}>{entries}</Root>;
}

export default StatusHistoryContainer;
