import Input from '@mui/material/Input';
import { styled } from '@mui/material/styles';
import { useTranslation } from 'react-i18next';

const PREFIX = 'ListSearch';

const classes = {
  container: `${PREFIX}-container`,
  input: `${PREFIX}-input`,
};

const Root = styled('div')(({ theme }) => ({
  [`&.${classes.container}`]: {
    display: 'flex',
    flexWrap: 'wrap',
  },

  [`& .${classes.input}`]: {
    margin: theme.spacing(1),
  },
}));

export default function SearchInput(props: { [key: string]: any }) {
  const { t } = useTranslation();

  return (
    <Root className={classes.container}>
      <Input
        className={classes.input}
        inputProps={{
          'aria-label': t(`frequent|${props.ariaLabel}`),
        }}
        {...props}
      />
    </Root>
  );
}
