import { Theme } from '@material-ui/core';
import Avatar from '@material-ui/core/Avatar';
import makeStyles from '@material-ui/styles/makeStyles';
import React from 'react';

interface ChannelAvatarProps {
  backgroundColor?: string;
  color?: string;
  size?: number;
  children?: React.ReactNode;
}

const useStyles = makeStyles( (theme: Theme) => ({
  colorAvatar: (props: ChannelAvatarProps) => ({
    color: '#fff',
    backgroundColor: props.backgroundColor || props.color ||
    theme.palette.secondary.main,
    width: props.size,
    height: props.size,
    display: 'inline-flex'
  }),
}));

export default function ChannelAvatar(props: ChannelAvatarProps) {
  const classes = useStyles(props);

  return (
    <Avatar className={classes.colorAvatar}>{props.children || '' }</Avatar>
  );
}
