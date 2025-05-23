import Avatar from '@mui/material/Avatar';

export interface ChannelAvatarProps {
  backgroundColor?: string;
  color?: string;
  size?: string | number;
  children?: React.ReactNode;
}

export default function ChannelAvatar({
  backgroundColor,
  color,
  size,
  children,
}: ChannelAvatarProps) {
  return (
    <Avatar
      sx={{
        color: 'rgb(15 15 15)',
        display: 'inline-flex',
        backgroundColor: theme => backgroundColor ?? color ?? theme.palette.secondary.main,
        width: size,
        height: size,
      }}
    >
      {children || ' '}
    </Avatar>
  );
}
