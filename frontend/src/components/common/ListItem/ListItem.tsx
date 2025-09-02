import MuiListItem from '@mui/material/ListItem';
import { styled } from '@mui/material/styles';
import React from 'react';

const PREFIX = 'ListItem';

const classes = {
  divider: `${PREFIX}-divider`,
};

const StyledMuiListItem = styled(MuiListItem)({
  [`&.${classes.divider}`]: {
    borderBottom: '2px solid rgba(0, 0, 0, 0.12)',
  },
});

export default function ListItem(props: { children: React.ReactNode; [key: string]: any }) {
  return (
    <StyledMuiListItem
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
