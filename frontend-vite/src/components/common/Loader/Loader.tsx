import Box from '@mui/material/Box';
import CircularProgress from '@mui/material/CircularProgress';
import makeStyles from '@mui/styles/makeStyles';

const useStyles = makeStyles({
  loaderContainer: {
    margin: '30px auto',
    textAlign: 'center',
  },
});

export default function Loader(props: { noContainer?: boolean }) {
  const classes = useStyles();
  const { noContainer = false, ...other } = props;
  const progress = <CircularProgress aria-label="Loading" {...other} />;

  if (noContainer) return progress;

  return (
    <Box className={classes.loaderContainer} data-testid="loader-container">
      {progress}
    </Box>
  );
}
