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
import PropTypes from 'prop-types';
import React from 'react';
import { FixedSizeList } from 'react-window';

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
    />
  );
}

renderInput.propTypes = {
  classes: PropTypes.object.isRequired,
  InputProps: PropTypes.object,
};

function renderSuggestion(suggestionProps) {
  const { suggestion, style, itemProps, selectedItem, getSecondaryLabel} = suggestionProps;
  const isSelected = (selectedItem || '').indexOf(suggestion.primary) > -1;

  return (
    <ListItem
      {...itemProps}
      button
      key={suggestion.primary}
      selected={isSelected}
      style={style}
    >
      <ListItemText
        primary={suggestion.primary}
        secondary={suggestion.secondary}
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

function getSuggestions(value, selectedItem, suggestions) {
  const inputValue = value.toLowerCase();

  if (value === selectedItem)
    return suggestions;

  return inputValue.length === 0 ? suggestions
    : suggestions.filter(suggestion => {
      return suggestion.primary.toLowerCase().includes(inputValue);
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
                             itemProps: getItemProps({item: suggestion.primary}),
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

export default function AutoCompletePicker(props) {
  const [showPicker, setShowPicker] = React.useState(false);
  const [selectedValue, setSelectedValue] = React.useState(props.defaultValue);
  const suggestions = props.getSuggestions;

  const classes = useStyles();
  function onInputActivate(event) {
    setShowPicker(true);
  }

  function handleClose(event) {
    setShowPicker(false);
  }

  function handleSelect(event) {
    setShowPicker(false);
    props.onSelect(selectedValue);
  }

  return (
    <div>
      <FormControl fullWidth>
        <InputLabel shrink>{props.label}</InputLabel>
        <Input
          onClick={onInputActivate}
          inputProps={{
            className: classes.pickerButtonInput
          }}
          value={selectedValue}
          placeholder={props.placeholder}
          readOnly
        />
      </FormControl>
      <Dialog open={showPicker}>
        <DialogTitle>
          {props.dialogTitle}
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
              setSelectedValue(selectedItem);

              const { onBlur, onChange, onFocus, ...inputProps } = getInputProps();

              return (
                <div className={classes.container}>
                  {renderInput({
                    fullWidth: true,
                    autoFocus: true,
                    classes,
                    label: props.label,
                    placeholder: props.pickerPlaceholder,
                    InputLabelProps: getLabelProps({ shrink: true }),
                    InputProps: { onBlur, onChange, onFocus },
                    inputProps,
                  })}
                  <LazyList
                    options={getSuggestions(inputValue, selectedItem, suggestions)}
                    itemData={{
                      getItemProps,
                      highlightedIndex,
                      selectedItem
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
