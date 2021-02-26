import green from '@material-ui/core/colors/green';
import { createMuiTheme } from '@material-ui/core/styles';

export const theme = createMuiTheme({
  palette: {
    primary: {
      contrastText: '#fff',
      main: process.env.REACT_APP_PRIMARY_COLOR,
    },
    success: {
      main: green['500'],
      ...green,
    },
  },
  typography: {
    fontFamily: 'Overpass, sans-serif',
  },
});
