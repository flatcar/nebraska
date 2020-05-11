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
import Downshift from 'downshift';
import moment from 'moment-timezone';
import PropTypes from 'prop-types';
import React from 'react';
import { FixedSizeList } from 'react-window';

const suggestions = moment.tz.names().map(timezone => {
  return {label: timezone,
          utcDiff: moment.tz(moment.utc(), timezone).utcOffset() / 60 // Hours from/to UTC
  };
});

function renderInput(inputProps) {
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

renderInput.propTypes = {
  classes: PropTypes.object.isRequired,
  InputProps: PropTypes.object,
};

function renderSuggestion(suggestionProps) {
  const { suggestion, style, itemProps, selectedItem } = suggestionProps;
  const isSelected = (selectedItem || '').indexOf(suggestion.label) > -1;

  function getUtcLabel(utcDiff) {
    return 'UTC ' + (utcDiff >= 0 ? '+' : '-') + Math.abs(utcDiff);
  }

  return (
    <ListItem
      {...itemProps}
      button
      key={suggestion.label}
      selected={isSelected}
      style={style}
    >
      <ListItemText
        primary={suggestion.label}
        secondary={getUtcLabel(suggestion.utcDiff)}
      />
    </ListItem>
  );
}

renderSuggestion.propTypes = {
  highlightedIndex: PropTypes.oneOfType([PropTypes.oneOf([null]), PropTypes.number]).isRequired,
  index: PropTypes.number.isRequired,
  itemProps: PropTypes.object.isRequired,
  selectedItem: PropTypes.string.isRequired,
  suggestion: PropTypes.shape({
    label: PropTypes.string.isRequired,
  }).isRequired,
};

function getSuggestions(value, selectedItem) {
  const inputValue = value.toLowerCase();

  if (value === selectedItem)
    return suggestions;

  return inputValue.length === 0 ? suggestions
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

function LazyList(props) {
  const {options, itemData, ...others} = props;

  itemData['suggestions'] = options;

  function Row(props) {
    const {index, style, data} = props;
    const suggestion = data.suggestions[index];
    const getItemProps = data.getItemProps;
    data['index'] = index;
    return renderSuggestion({suggestion,
                             style,
                             itemProps: getItemProps({item: suggestion.label}),
                             ...data});
  }

  return (
    <FixedSizeList
      itemCount={options.length}
      itemData={itemData}
      {...others}
    >
      {Row}
    </FixedSizeList>
  );
}

export const DEFAULT_TIMEZONE = moment.tz.guess(true);

export default function TimzonePicker(props) {
  const [showPicker, setShowPicker] = React.useState(false);
  const [selectedTimezone, setSelectedTimezone] =
    React.useState(props.value ? props.value : DEFAULT_TIMEZONE);
  const classes = useStyles();

  function onInputActivate(event) {
    setShowPicker(true);
  }

  function handleClose(event) {
    setShowPicker(false);
  }

  function handleSelect(event) {
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
            className: classes.pickerButtonInput
          }}
          placeholder="Pick a timezone"
          readOnly
          data-testid="timezone-readonly-input"
        />
      </FormControl>
      <Dialog open={showPicker}>
        <DialogTitle>
          Choose a Timezone
        </DialogTitle>
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
                    label: 'Timezone',
                    placeholder: 'Start typing to search a timezone',
                    InputLabelProps: getLabelProps({ shrink: true }),
                    InputProps: { onBlur, onChange, onFocus },
                    inputProps,
                  })}
                  <LazyList
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
          <Button onClick={handleClose} color="primary">Cancel</Button>
          <Button onClick={handleSelect} color="primary">Select</Button>
        </DialogActions>
      </Dialog>
    </div>
  );
}
