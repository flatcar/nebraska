import { ListItemText } from '@mui/material';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import FormControl from '@mui/material/FormControl';
import Input from '@mui/material/Input';
import InputLabel, { InputLabelProps } from '@mui/material/InputLabel';
import ListItemButton from '@mui/material/ListItemButton';
import TextField from '@mui/material/TextField';
import Downshift from 'downshift';
import React from 'react';
import { useTranslation } from 'react-i18next';
import { List } from 'react-window';

interface RenderInputProps {
  ref?: React.Ref<any>;
  fullWidth: boolean;
  autoFocus: boolean;
  label: string;
  placeholder: string;
  InputLabelProps: InputLabelProps;
  InputProps: {
    onBlur?: React.FocusEventHandler<HTMLInputElement | HTMLTextAreaElement>;
    onChange?: React.ChangeEventHandler<HTMLTextAreaElement | HTMLInputElement>;
    onFocus?: React.FocusEventHandler<HTMLInputElement | HTMLTextAreaElement>;
  };
  inputProps: object;
  variant: 'outlined';
  onKeyDown: React.KeyboardEventHandler<HTMLDivElement>;
}

function renderInput(inputProps: RenderInputProps) {
  const { InputProps, ref, onKeyDown, ...other } = inputProps;

  return (
    <TextField
      sx={{
        marginTop: '0.6em',
      }}
      slotProps={{
        input: {
          inputRef: ref,
          sx: {
            flexWrap: 'wrap',
            '& input': {
              width: 'auto',
              flexGrow: 1,
            },
          },
          ...InputProps,
        },
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
    primary: string;
    secondary: string;
  };
  style?: object;
  getSecondaryLabel?: () => object;
}

function renderSuggestion(suggestionProps: RenderSuggestionProps) {
  const { suggestion, style, itemProps, selectedItem } = suggestionProps;
  const isSelected = selectedItem === suggestion.primary;

  return (
    <ListItemButton {...itemProps} key={suggestion.primary} selected={isSelected} style={style}>
      <ListItemText primary={suggestion.primary} secondary={suggestion.secondary} />
    </ListItemButton>
  );
}

/**
 * Filters suggestions to those which match value.
 * @param value to search for.
 * @param selectedItem if the value is already selected, return unfiltered suggestions.
 * @param suggestions to search through.
 * @returns filtered suggestions that match value.
 */
function filterSuggestions(
  value: string | null,
  selectedItem: string,
  suggestions: RenderSuggestionProps['suggestion'][]
) {
  if (!value || value === selectedItem) {
    return suggestions;
  }

  const inputValue = value.toLowerCase();

  return suggestions.filter(suggestion => {
    return suggestion.primary.toLowerCase().includes(inputValue);
  });
}

interface LazyListProps {
  options: RenderSuggestionProps['suggestion'][];
  itemData: any;
  height: number;
  itemSize: number;
  width: number;
  onBottomScrolled?: () => void;
}

function RowComponent(props: {
  index: number;
  style: React.CSSProperties;
  suggestions: any[];
  getItemProps: any;
  highlightedIndex: any;
  selectedItem: any;
}) {
  const { index, style, suggestions, getItemProps, highlightedIndex, selectedItem } = props;
  const suggestion = suggestions[index];
  return renderSuggestion({
    suggestion,
    style,
    itemProps: getItemProps({ item: suggestion.primary }),
    index,
    highlightedIndex,
    selectedItem,
  });
}

function createOnItemsRendered(onBottomScrolled?: () => void) {
  return function onItemsRendered(args: {
    visibleStartIndex: number;
    visibleStopIndex: number;
    overscanStartIndex: number;
    overscanStopIndex: number;
  }) {
    const { overscanStopIndex, visibleStopIndex } = args;
    if (onBottomScrolled && overscanStopIndex === visibleStopIndex) {
      onBottomScrolled();
    }
  };
}

function LazyList(props: LazyListProps) {
  const { options, itemData, height, itemSize, width, onBottomScrolled } = props;

  const rowProps = { ...itemData, suggestions: options };
  const onItemsRendered = createOnItemsRendered(onBottomScrolled);

  return (
    <List
      rowCount={options.length}
      rowHeight={itemSize}
      rowComponent={RowComponent}
      rowProps={rowProps}
      onRowsRendered={(visibleRows, allRows) => {
        onItemsRendered({
          visibleStartIndex: visibleRows.startIndex,
          visibleStopIndex: visibleRows.stopIndex,
          overscanStartIndex: allRows.startIndex,
          overscanStopIndex: allRows.stopIndex,
        });
      }}
      style={{ height, width }}
      role="listbox"
      aria-label="Autocomplete options"
    />
  );
}

export interface AutoCompletePickerProps {
  /** The default value. Use when the component is not controlled. */
  defaultValue: string;
  /** Suggestions that can be picked. */
  suggestions: RenderSuggestionProps['suggestion'][];
  /** Callback fired when the value is selected. */
  onSelect: (selectedValue: string) => void;
  /** The label content. */
  label: string;
  /** The short hint displayed in the input before the user enters a value. */
  placeholder: string;
  /** Title shown when the picker is being displayed. */
  dialogTitle: string;
  /** A separate placeholder for the picker. */
  pickerPlaceholder: string;
  onValueChanged: (value?: string | null) => void;
  /**  */
  onBottomScrolled?: () => void;
  /** Should the color picker be displayed initially? */
  initialOpen?: boolean;
}

export default function AutoCompletePicker(props: AutoCompletePickerProps) {
  const [showPicker, setShowPicker] = React.useState(props.initialOpen);
  const [selectedValue, setSelectedValue] = React.useState(props.defaultValue);
  const [currentValue, setCurrentValue] = React.useState(props.defaultValue);
  const { onBottomScrolled } = props;
  const { t } = useTranslation();

  function onInputChange(event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) {
    props.onValueChanged(event.target.value);
  }

  function handleEscape(event: React.KeyboardEvent<HTMLDivElement>) {
    if (event.key === 'Escape') {
      props.onValueChanged('');
    }
  }

  return (
    <Box>
      <FormControl fullWidth>
        <InputLabel variant="standard" shrink>
          {props.label}
        </InputLabel>
        <Input
          onClick={() => {
            setShowPicker(true);
          }}
          slotProps={{
            input: {
              sx: {
                cursor: 'pointer',
              },
            },
          }}
          value={currentValue}
          placeholder={props.placeholder}
          readOnly
        />
      </FormControl>
      <Dialog open={showPicker || false}>
        <DialogTitle>{props.dialogTitle}</DialogTitle>
        <DialogContent>
          <Downshift
            id="downshift-options"
            onChange={selectedItem => {
              if (selectedItem) {
                setSelectedValue(selectedItem);
              }
            }}
          >
            {({
              getInputProps,
              getItemProps,
              getLabelProps,
              getRootProps,
              highlightedIndex,
              inputValue,
              selectedItem,
            }) => {
              const { onBlur, ...inputProps } = getInputProps();

              return (
                <Box
                  {...getRootProps()}
                  sx={{
                    flexGrow: 1,
                    position: 'relative',
                  }}
                >
                  {renderInput({
                    fullWidth: true,
                    autoFocus: true,
                    label: props.label,
                    placeholder: props.pickerPlaceholder,
                    InputLabelProps: getLabelProps(),
                    InputProps: { onBlur, onChange: onInputChange },
                    inputProps,
                    variant: 'outlined',
                    onKeyDown: handleEscape,
                  })}
                  <LazyList
                    options={filterSuggestions(inputValue, selectedItem, props.suggestions)}
                    itemData={{
                      getItemProps,
                      highlightedIndex,
                      selectedItem,
                    }}
                    height={400}
                    width={400}
                    itemSize={50}
                    onBottomScrolled={onBottomScrolled}
                  />
                </Box>
              );
            }}
          </Downshift>
        </DialogContent>
        <DialogActions>
          <Button
            onClick={() => {
              setShowPicker(false);
              // It's important to send this value as a way to tell that no value is
              // selected any longer.
              props.onValueChanged(null);
            }}
            color="primary"
          >
            {t('frequent|cancel')}
          </Button>
          <Button
            onClick={() => {
              setShowPicker(false);
              setCurrentValue(selectedValue);
              props.onSelect(selectedValue);
            }}
            color="primary"
          >
            {t('frequent|select')}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
