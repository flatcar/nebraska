import { makeStyles } from '@material-ui/core/styles';
import React from 'react';

const useStyles = makeStyles(theme => ({
  label: {
    background: '#b4b4b4',
    color: '#ffffff',
    fontSize: '75%',
    textAlign: 'center',
    borderRadius: '.2em',
    padding: '.2em .6em .3em',
  },
}));

export default function Label(props) {
  const classes = useStyles();
  return (
    <span className={classes.label}>
      {props.children}
    </span>
  );
}
