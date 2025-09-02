import FormControl from '@mui/material/FormControl';
import FormLabel from '@mui/material/FormLabel';
import MenuItem from '@mui/material/MenuItem';
import Select, { SelectChangeEvent } from '@mui/material/Select';
import { styled } from '@mui/material/styles';
import { useTheme } from '@mui/material/styles';
import { useTranslation } from 'react-i18next';

const PREFIX = 'LocaleSelect';

const classes = {
  formControl: `${PREFIX}-formControl`,
};

const StyledFormControl = styled(FormControl)(({ theme }) => ({
  [`&.${classes.formControl}`]: {
    margin: theme.spacing(2),
  },
}));

export interface LocaleSelectProps {
  showTitle?: boolean;
}

/**
 * A UI for selecting the locale with i18next
 */
export default function LocaleSelect(props: LocaleSelectProps) {
  const { t, i18n } = useTranslation('frequent');
  const theme = useTheme();

  const changeLng = (event: SelectChangeEvent<string>) => {
    const lng = event.target.value;

    i18n.changeLanguage(lng);
    document.body.dir = i18n.dir();
    theme.direction = i18n.dir();
  };

  return (
    <StyledFormControl className={classes.formControl}>
      {props.showTitle && <FormLabel component="legend">{t('Select locale')}</FormLabel>}
      <Select
        value={i18n.language ? i18n.language : 'en'}
        onChange={changeLng}
        inputProps={{ 'aria-label': t('Select locale') }}
      >
        {i18n?.options?.supportedLngs &&
          i18n.options.supportedLngs
            .filter(lng => lng !== 'cimode')
            .map(lng => <MenuItem value={lng}>{lng}</MenuItem>)}
      </Select>
    </StyledFormControl>
  );
}
