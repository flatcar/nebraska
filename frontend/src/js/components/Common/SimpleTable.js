import squareIcon from '@iconify/icons-mdi/square';
import { InlineIcon } from '@iconify/react';
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableHead from '@material-ui/core/TableHead';
import TablePagination from '@material-ui/core/TablePagination';
import TableRow from '@material-ui/core/TableRow';
import React from 'react';
import Empty from '../Common/EmptyContent';

export default function SimpleTable(props) {
  const {columns} = props;
  const [page, setPage] = React.useState(0);
  const rowsPerPageOptions = [5, 10, 50];
  const [rowsPerPage, setRowsPerPage] = React.useState(rowsPerPageOptions[0]);

  function handleChangePage(event, newPage) {
    setPage(newPage);
  }

  function handleChangeRowsPerPage(event) {
    setRowsPerPage(+event.target.value);
    setPage(0);
  }

  React.useEffect(() => {
    setPage(0);
  },
  [props.instances]);

  function getPagedRows() {
    const startIndex = page * rowsPerPage;
    return props.instances.slice(startIndex, startIndex + rowsPerPage);
  }

  return (
    props.instances.length === 0 ?
      <Empty>{props.emptyMessage ? props.emptyMessage : 'No data to be shown.'}</Empty>
      :
      <React.Fragment>
        <Table>
          <TableHead>
            <TableRow>
              {Object.keys(columns).map((column, i) =>
                <TableCell key={`tabletitle_${i}`}>{columns[column]}</TableCell>
              )}
            </TableRow>
          </TableHead>
          <TableBody>
            {props.instances &&
           getPagedRows().map((row, i) =>
             <TableRow key={i}>
               {Object.keys(columns).map((column, i) =>
                 <TableCell key={`cell_${i}`}>
                   {i === 0 && row.color &&
                   <React.Fragment>
                     <InlineIcon
                       icon={squareIcon}
                       color={row.color}
                       height="15"
                       width="15"
                     />
                     &nbsp;
                   </React.Fragment>
                   }
                   {row[column]}
                 </TableCell>
               )}
             </TableRow>
           )}
          </TableBody>
        </Table>
        {props.instances.length > rowsPerPageOptions[0] &&
          <TablePagination
            rowsPerPageOptions={rowsPerPageOptions}
            component="div"
            count={props.instances.length}
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
        }
      </React.Fragment>
  );
}
