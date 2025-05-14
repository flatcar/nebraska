import { styled } from '@mui/material/styles';
import React from 'react';

const PREFIX = 'Label';

const classes = {
  label: `${PREFIX}-label`
};

const Root = styled('span')({
  [`&.${classes.label}`]: {
    background: '#b4b4b4',
    color: '#ffffff',
    fontSize: '75%',
    textAlign: 'center',
    borderRadius: '.2em',
    padding: '.2em .6em .3em',
  },
});

export default function Label(props: { children: React.ReactNode }) {
  return <Root className={classes.label}>{props.children}</Root>;
}
