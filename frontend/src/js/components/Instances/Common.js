import { Box, Link } from '@material-ui/core';
import Grid from '@material-ui/core/Grid';
import Typography from '@material-ui/core/Typography';
import { makeStyles } from '@material-ui/styles';
import React from 'react';
import { Link as RouterLink} from 'react-router-dom';

const useInstanceCountStyles = makeStyles(theme => ({
  instancesCount: {
    fontSize: '2rem;',
    fontWeight: 700,
    paddingBottom: '0.5rem',
    color: theme.palette.greyShadeColor
  },
  instancesLabel: {
    color: theme.palette.text.secondary,
    fontSize: '1rem;',
    paddingTop: '0.5rem'
  },
  instanceLink: {
    fontSize: '1.2rem'
  }
}));

export function InstanceCountLabel(props) {
  const classes = useInstanceCountStyles();
  const {countText, href, instanceListView = false} = props;

  return (
    <Grid
      container
      direction="column"
    >
      <Grid item>
        <Typography className={classes.instancesLabel}>INSTANCES</Typography>
      </Grid>
      <Grid item>
        <Typography className={classes.instancesCount}>{countText}</Typography>
      </Grid>
      <Grid item>
        {!instanceListView && countText > 0 ?
          <Box>
            {!props.loading &&
            <Link
              to={{...href}}
              component={RouterLink}
            >
              <Typography className={classes.instancesLink}>See all instances</Typography>
            </Link>
            }
          </Box>
          :
          []
        }
      </Grid>
    </Grid>
  );
}
