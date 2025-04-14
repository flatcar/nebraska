import makeStyles from '@mui/styles/makeStyles';
import React from 'react';

const useStyles = makeStyles({
  label: {
    background: '#b4b4b4',
    color: '#ffffff',
    fontSize: '75%',
    textAlign: 'center',
    borderRadius: '.2em',
    padding: '.2em .6em .3em',
  },
});

export default function Label(props: { children: React.ReactNode }) {
  const classes = useStyles();
  return <span className={classes.label}>{props.children}</span>;
}
