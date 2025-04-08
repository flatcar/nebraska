import MuiListItem from '@mui/material/ListItem';
import makeStyles from '@mui/styles/makeStyles';
import React from 'react';

const useStyles = makeStyles({
  divider: {
    borderBottom: '2px solid rgba(0, 0, 0, 0.12)',
  },
});

export default function ListItem(props: { children: React.ReactNode; [key: string]: any }) {
  const classes = useStyles();

  return (
    <MuiListItem
      classes={{
        divider: classes.divider,
      }}
      divider
      disableGutters
      {...props}
      data-testid="list-item"
    />
  );
}
