import Avatar from '@material-ui/core/Avatar';
import makeStyles from '@material-ui/styles/makeStyles';
import React from 'react';

const useStyles = makeStyles({
  colorAvatar: props => ({
    color: props.color,
    backgroundColor: props.backgroundColor || props.color,
    width: props.size,
    height: props.size,
    display: 'inline-block'
  }),
});

export default function ChannelAvatar(props) {
  const classes = useStyles(props);

  return (
    <Avatar className={classes.colorAvatar}>{props.children}</Avatar>
  );
}
