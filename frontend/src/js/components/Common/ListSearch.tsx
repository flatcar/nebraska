import Input from '@material-ui/core/Input';
import { makeStyles } from '@material-ui/core/styles';
import React from 'react';
import { useTranslation } from 'react-i18next';

const useStyles = makeStyles(theme => ({
  container: {
    display: 'flex',
    flexWrap: 'wrap',
  },
  input: {
    margin: theme.spacing(1),
  },
}));

export default function SearchInput(props: { [key: string]: any }) {
  const classes = useStyles();
  const { t } = useTranslation();

  return (
    <div className={classes.container}>
      <Input
        className={classes.input}
        inputProps={{
          'aria-label': t(`frequent|${props.ariaLabel}`),
        }}
        {...props}
      />
    </div>
  );
}
