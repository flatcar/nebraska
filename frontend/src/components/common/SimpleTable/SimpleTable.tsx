import squareIcon from '@iconify/icons-mdi/square';
import { InlineIcon } from '@iconify/react';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableHead from '@mui/material/TableHead';
import TablePagination from '@mui/material/TablePagination';
import TableRow from '@mui/material/TableRow';
import React from 'react';
import { useTranslation } from 'react-i18next';
import Empty from '../EmptyContent/EmptyContent';

interface SimpleTableProps {
  columns: {
    [key: string]: string;
  };
  instances: any[];
  emptyMessage: React.ReactNode;
}

export default function SimpleTable(props: SimpleTableProps) {
  const { columns } = props;
  const [page, setPage] = React.useState(0);
  const rowsPerPageOptions = [5, 10, 50];
  const [rowsPerPage, setRowsPerPage] = React.useState(rowsPerPageOptions[0]);
  const { t } = useTranslation();

  function handleChangePage(event: any, newPage: number) {
    setPage(newPage);
  }

  function handleChangeRowsPerPage(event: any) {
    setRowsPerPage(+event.target.value);
    setPage(0);
  }

  React.useEffect(() => {
    setPage(0);
  }, [props.instances]);

  function getPagedRows() {
    const startIndex = page * rowsPerPage;
    return props.instances.slice(startIndex, startIndex + rowsPerPage);
  }

  return props.instances.length === 0 ? (
    <Empty>{props.emptyMessage ? props.emptyMessage : t('common|no_data_message')}</Empty>
  ) : (
    <React.Fragment>
      <Table>
        <TableHead>
          <TableRow>
            {Object.keys(columns).map((column, i) => (
              <TableCell key={`tabletitle_${i}`}>{columns[column]}</TableCell>
            ))}
          </TableRow>
        </TableHead>
        <TableBody>
          {props.instances &&
            getPagedRows().map((row, i) => (
              <TableRow key={i}>
                {Object.keys(columns).map((column, i) => (
                  <TableCell key={`cell_${i}`}>
                    {i === 0 && row.color && (
                      <React.Fragment>
                        <InlineIcon icon={squareIcon} color={row.color} height="15" width="15" />
                        &nbsp;
                      </React.Fragment>
                    )}
                    {row[column]}
                  </TableCell>
                ))}
              </TableRow>
            ))}
        </TableBody>
      </Table>
      {props.instances.length > rowsPerPageOptions[0] && (
        <TablePagination
          rowsPerPageOptions={rowsPerPageOptions}
          component="div"
          count={props.instances.length}
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
      )}
    </React.Fragment>
  );
}
