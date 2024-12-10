import green from '@material-ui/core/colors/green';
import { createTheme } from '@material-ui/core/styles';

export const theme = createTheme({
  palette: {
    primary: {
      contrastText: '#fff',
      main: process.env.REACT_APP_PRIMARY_COLOR || '#000',
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
