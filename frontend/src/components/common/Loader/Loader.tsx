import Box from '@mui/material/Box';
import CircularProgress from '@mui/material/CircularProgress';
import { styled } from '@mui/material/styles';
const PREFIX = 'Loader';

const classes = {
  loaderContainer: `${PREFIX}-loaderContainer`,
};

const StyledBox = styled(Box)({
  [`&.${classes.loaderContainer}`]: {
    margin: '30px auto',
    textAlign: 'center',
  },
});

export default function Loader(props: { noContainer?: boolean }) {
  const { noContainer = false, ...other } = props;
  const progress = <CircularProgress aria-label="Loading" {...other} />;

  if (noContainer) return progress;

  return (
    <StyledBox className={classes.loaderContainer} data-testid="loader-container">
      {progress}
    </StyledBox>
  );
}
