import Button from '@material-ui/core/Button';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogTitle from '@material-ui/core/DialogTitle';
import FormControl from '@material-ui/core/FormControl';
import Input from '@material-ui/core/Input';
import InputLabel from '@material-ui/core/InputLabel';
import ListItem from '@material-ui/core/ListItem';
import ListItemText from '@material-ui/core/ListItemText';
import { makeStyles } from '@material-ui/core/styles';
import TextField from '@material-ui/core/TextField';
import Downshift, { GetLabelPropsOptions } from 'downshift';
import moment from 'moment-timezone';
import React from 'react';
import { useTranslation } from 'react-i18next';
import { FixedSizeList } from 'react-window';

const suggestions = moment.tz.names().map(timezone => {
  return {
    label: timezone,
    utcDiff: moment.tz(moment.utc(), timezone).utcOffset() / 60, // Hours from/to UTC
  };
});

interface RenderInputProps {
  classes: {
    inputRoot: string;
    inputInput: string;
  };
  ref?: React.Ref<any>;
  InputProps: {
    onBlur: () => void;
    onChange: () => void;
    onFocus: () => void;
  };
  fullWidth: boolean;
  autoFocus: boolean;
  label: string;
  placeholder: string;
  InputLabelProps: (options?: GetLabelPropsOptions | undefined) => void;
  variant: 'outlined';
  inputProps: object;
}

function renderInput(inputProps: RenderInputProps) {
  const { InputProps, classes, ref, ...other } = inputProps;

  return (
    <TextField
      InputProps={{
        inputRef: ref,
        classes: {
          root: classes.inputRoot,
          input: classes.inputInput,
        },
        ...InputProps,
      }}
      {...other}
      data-testid="timezone-input"
    />
  );
}

interface RenderSuggestionProps {
  highlightedIndex: null | number;
  index: number;
  itemProps: object;
  selectedItem: string;
  suggestion: {
    utcDiff: number;
    label: string;
  };
  style?: object;
}

function renderSuggestion(suggestionProps: RenderSuggestionProps) {
  const { suggestion, style, itemProps, selectedItem } = suggestionProps;
  const isSelected = (selectedItem || '').indexOf(suggestion.label) > -1;

  function getUtcLabel(utcDiff: number) {
    return 'UTC ' + (utcDiff >= 0 ? '+' : '-') + Math.abs(utcDiff);
  }

  return (
    <ListItem {...itemProps} button key={suggestion.label} selected={isSelected} style={style}>
      <ListItemText primary={suggestion.label} secondary={getUtcLabel(suggestion.utcDiff)} />
    </ListItem>
  );
}

function getSuggestions(value: string | null, selectedItem: string) {
  if (!value) {
    return suggestions;
  }
  const inputValue = value.toLowerCase();

  if (value === selectedItem) return suggestions;

  return inputValue.length === 0
    ? suggestions
    : suggestions.filter(suggestion => {
        return suggestion.label.toLowerCase().includes(inputValue);
      });
}

const useStyles = makeStyles(theme => ({
  container: {
    flexGrow: 1,
    position: 'relative',
  },
  inputRoot: {
    flexWrap: 'wrap',
  },
  inputInput: {
    width: 'auto',
    flexGrow: 1,
  },
  pickerButtonInput: {
    cursor: 'pointer',
  },
}));

interface LazyListProps {
  options: string[];
  itemData: any;
}

interface LazyListProps {
  //@ts-ignore
  options: RenderSuggestionProps['suggestion'][];
  itemData: any;
  height: number;
  itemSize: number;
  width: number;
}

function LazyList(props: LazyListProps) {
  const { options, itemData, ...others } = props;

  itemData['suggestions'] = options;

  function Row(props: { index: number; style: object; data: any }) {
    const { index, style, data } = props;
    const suggestion = data.suggestions[index];
    const getItemProps = data.getItemProps;
    data['index'] = index;
    return renderSuggestion({
      suggestion,
      style,
      itemProps: getItemProps({ item: suggestion.label }),
      ...data,
    });
  }

  return (
    <FixedSizeList itemCount={options.length} itemData={itemData} {...others}>
      {Row}
    </FixedSizeList>
  );
}

export const DEFAULT_TIMEZONE = moment.tz.guess(true);

export default function TimzonePicker(props: {
  value: string;
  onSelect: (selectedTimezone: string) => void;
}) {
  const [showPicker, setShowPicker] = React.useState(false);
  const [selectedTimezone, setSelectedTimezone] = React.useState(
    props.value ? props.value : DEFAULT_TIMEZONE
  );
  const classes = useStyles();
  const { t } = useTranslation();

  function onInputActivate() {
    setShowPicker(true);
  }

  function handleClose() {
    setShowPicker(false);
  }

  function handleSelect() {
    setShowPicker(false);
    props.onSelect(selectedTimezone);
  }

  return (
    <div>
      <FormControl fullWidth>
        <InputLabel shrink>Timezone</InputLabel>
        <Input
          onClick={onInputActivate}
          value={selectedTimezone}
          inputProps={{
            className: classes.pickerButtonInput,
          }}
          placeholder={t('common|Pick a timezone')}
          readOnly
          data-testid="timezone-readonly-input"
        />
      </FormControl>
      <Dialog open={showPicker}>
        <DialogTitle>Choose a Timezone</DialogTitle>
        <DialogContent>
          <Downshift id="downshift-options">
            {({
              getInputProps,
              getItemProps,
              getLabelProps,
              highlightedIndex,
              inputValue,
              selectedItem,
            }) => {
              setSelectedTimezone(selectedItem);

              const { onBlur, onChange, onFocus, ...inputProps } = getInputProps();

              return (
                <div className={classes.container}>
                  {renderInput({
                    fullWidth: true,
                    autoFocus: true,
                    classes,
                    label: t('common|Timezone'),
                    placeholder: t('common|Start typing to search a timezone'),
                    InputLabelProps: getLabelProps(),
                    InputProps: { onBlur, onChange, onFocus },
                    inputProps,
                    variant: 'outlined',
                  })}
                  <LazyList
                    //@todo add better types
                    //@ts-ignore
                    options={getSuggestions(inputValue, selectedItem)}
                    itemData={{
                      getItemProps,
                      highlightedIndex,
                      selectedItem,
                    }}
                    height={400}
                    width={400}
                    itemSize={50}
                  />
                </div>
              );
            }}
          </Downshift>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleClose} color="primary">
            {t('frequent|Cancel')}
          </Button>
          <Button onClick={handleSelect} color="primary">
            {t('frequent|Select')}
          </Button>
        </DialogActions>
      </Dialog>
    </div>
  );
}
