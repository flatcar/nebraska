import Avatar from '@material-ui/core/Avatar';
import makeStyles from '@material-ui/styles/makeStyles';
import React from 'react';

const useStyles = makeStyles( theme => ({
  colorAvatar: props => ({
    color: '#fff',
    backgroundColor: props.backgroundColor || props.color || theme.palette.secondary.main,
    width: props.size,
    height: props.size,
    display: 'inline-flex'
  }),
}));

export default function ChannelAvatar(props) {
  const classes = useStyles(props);

  return (
    <Avatar className={classes.colorAvatar}>{props.children}</Avatar>
  );
}
