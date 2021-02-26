import { ListItemText } from '@material-ui/core';
import Button from '@material-ui/core/Button';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogTitle from '@material-ui/core/DialogTitle';
import FormControl from '@material-ui/core/FormControl';
import Input from '@material-ui/core/Input';
import InputLabel from '@material-ui/core/InputLabel';
import ListItem from '@material-ui/core/ListItem';
import { makeStyles } from '@material-ui/core/styles';
import TextField from '@material-ui/core/TextField';
import Downshift, { GetLabelPropsOptions } from 'downshift';
import React from 'react';
import { FixedSizeList } from 'react-window';

interface RenderInputProps {
  classes: {
    inputRoot: string;
    inputInput: string;
  };
  ref?: React.Ref<any>;
  fullWidth: boolean;
  autoFocus: boolean;
  label: string;
  placeholder: string;
  InputLabelProps: (options?: GetLabelPropsOptions | undefined) => void;
  InputProps: {
    onBlur: () => void;
    onChange: () => void;
    onFocus: () => void;
  };
  inputProps: object;
  variant: 'outlined';
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
    />
  );
}

interface RenderSuggestionProps {
  highlightedIndex: null | number;
  index: number;
  itemProps: object;
  selectedItem: string;
  suggestion: {
    label: string;
    primary: string;
    secondary: string;
  };
  style?: object;
  getSecondaryLabel?: () => {};
}

function renderSuggestion(suggestionProps: RenderSuggestionProps) {
  const { suggestion, style, itemProps, selectedItem } = suggestionProps;
  const isSelected = (selectedItem || '').indexOf(suggestion.primary) > -1;

  return (
    <ListItem {...itemProps} button key={suggestion.primary} selected={isSelected} style={style}>
      <ListItemText primary={suggestion.primary} secondary={suggestion.secondary} />
    </ListItem>
  );
}

function getSuggestions(
  value: string | null,
  selectedItem: string,
  suggestions: RenderSuggestionProps['suggestion'][]
) {
  if (!value) {
    return suggestions;
  }

  const inputValue = value.toLowerCase();

  if (value === selectedItem) return suggestions;

  return inputValue.length === 0
    ? suggestions
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

interface LazyListProps {
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
      itemProps: getItemProps({ item: suggestion.primary }),
      ...data,
    });
  }

  return (
    <FixedSizeList itemCount={options.length} itemData={itemData} {...others}>
      {Row}
    </FixedSizeList>
  );
}
interface AutoCompletePickerProps {
  defaultValue: string;
  getSuggestions: RenderSuggestionProps['suggestion'][];
  onSelect: (selectedValue: string) => void;
  label: string;
  placeholder: string;
  dialogTitle: string;
  pickerPlaceholder: string;
}

export default function AutoCompletePicker(props: AutoCompletePickerProps) {
  const [showPicker, setShowPicker] = React.useState(false);
  const [selectedValue, setSelectedValue] = React.useState(props.defaultValue);
  const suggestions = props.getSuggestions;

  const classes = useStyles();
  function onInputActivate() {
    setShowPicker(true);
  }

  function handleClose() {
    setShowPicker(false);
  }

  function handleSelect() {
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
            className: classes.pickerButtonInput,
          }}
          value={selectedValue}
          placeholder={props.placeholder}
          readOnly
        />
      </FormControl>
      <Dialog open={showPicker}>
        <DialogTitle>{props.dialogTitle}</DialogTitle>
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
                    InputLabelProps: getLabelProps(),
                    InputProps: { onBlur, onChange, onFocus },
                    inputProps,
                    variant: 'outlined',
                  })}
                  <LazyList
                    options={getSuggestions(inputValue, selectedItem, suggestions)}
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
            Cancel
          </Button>
          <Button onClick={handleSelect} color="primary">
            Select
          </Button>
        </DialogActions>
      </Dialog>
    </div>
  );
}
