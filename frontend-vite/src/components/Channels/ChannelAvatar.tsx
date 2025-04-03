import { Theme } from '@mui/material';
import Avatar from '@mui/material/Avatar';
import makeStyles from '@mui/styles/makeStyles';

export interface ChannelAvatarProps {
  backgroundColor?: string;
  color?: string;
  size?: string | number;
  children?: React.ReactNode;
}

const useStyles = makeStyles((theme: Theme) => ({
  colorAvatar: (props: ChannelAvatarProps) => ({
    color: 'rgb(15 15 15)',
    backgroundColor: props.backgroundColor || props.color || theme.palette.secondary.main,
    width: props.size,
    height: props.size,
    display: 'inline-flex',
  }),
}));

export default function ChannelAvatar(props: ChannelAvatarProps) {
  const classes = useStyles(props);

  return <Avatar className={classes.colorAvatar}>{props.children || ' '}</Avatar>;
}
