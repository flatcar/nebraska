import green from '@mui/material/colors/green';
import { createTheme } from '@mui/material/styles';

export const theme = createTheme({
  palette: {
    primary: {
      contrastText: '#fff',
      main: process.env.REACT_APP_PRIMARY_COLOR || '#000',
    },
    success: {
      main: green['800'],
      ...green,
    },
  },
  typography: {
    fontFamily: 'Overpass, sans-serif',
  },
});
