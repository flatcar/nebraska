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
import { useTranslation } from 'react-i18next';
import { FixedSizeList, ListOnItemsRenderedProps } from 'react-window';

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
    onChange: React.ChangeEventHandler<HTMLInputElement | HTMLTextAreaElement>;
    onFocus: () => void;
  };
  inputProps: object;
  variant: 'outlined';
  onKeyDown: React.KeyboardEventHandler<HTMLDivElement>;
}

function renderInput(inputProps: RenderInputProps) {
  const { InputProps, classes, ref, onKeyDown, ...other } = inputProps;

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
      onKeyDown={onKeyDown}
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

const useStyles = makeStyles({
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
});

interface LazyListProps {
  options: RenderSuggestionProps['suggestion'][];
  itemData: any;
  height: number;
  itemSize: number;
  width: number;
  onItemsRendered: (args: ListOnItemsRenderedProps) => any;
}

function LazyList(props: LazyListProps) {
  const { options, itemData, onItemsRendered, ...others } = props;

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
    <FixedSizeList
      itemCount={options.length}
      itemData={itemData}
      onItemsRendered={onItemsRendered}
      {...others}
    >
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
  onValueChanged: (value?: string | null) => void;
  onBottomScrolled?: () => void;
}

export default function AutoCompletePicker(props: AutoCompletePickerProps) {
  const [showPicker, setShowPicker] = React.useState(false);
  const [selectedValue, setSelectedValue] = React.useState(props.defaultValue);
  const [currentValue, setCurrentValue] = React.useState(props.defaultValue);
  const suggestions = props.getSuggestions;
  const { onBottomScrolled } = props;
  const { t } = useTranslation();

  const classes = useStyles();
  function onInputActivate() {
    setShowPicker(true);
  }

  function handleClose() {
    setShowPicker(false);
    // It's important to send this value as a way to tell that no value is
    // selected any longer.
    props.onValueChanged(null);
  }

  function handleSelect() {
    setShowPicker(false);
    setCurrentValue(selectedValue);
    props.onSelect(selectedValue);
  }

  function onInputChange(event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) {
    props.onValueChanged(event.target.value);
  }

  function onItemsRendered(args: ListOnItemsRenderedProps) {
    const { overscanStopIndex, visibleStopIndex } = args;
    if (!!onBottomScrolled && overscanStopIndex === visibleStopIndex) {
      onBottomScrolled();
    }
  }

  function handleEscape(event: React.KeyboardEvent<HTMLDivElement>) {
    if (event.key === 'Escape') {
      props.onValueChanged('');
    }
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
          value={currentValue}
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

              const { onBlur, onFocus, ...inputProps } = getInputProps();

              return (
                <div className={classes.container}>
                  {renderInput({
                    fullWidth: true,
                    autoFocus: true,
                    classes,
                    label: props.label,
                    placeholder: props.pickerPlaceholder,
                    InputLabelProps: getLabelProps(),
                    InputProps: { onBlur, onChange: onInputChange, onFocus },
                    inputProps,
                    variant: 'outlined',
                    onKeyDown: handleEscape,
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
                    onItemsRendered={onItemsRendered}
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
