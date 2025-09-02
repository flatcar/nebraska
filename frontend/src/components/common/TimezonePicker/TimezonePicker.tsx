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
import ListItemText from '@mui/material/ListItemText';
import TextField from '@mui/material/TextField';
import Downshift from 'downshift';
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
  ref?: React.Ref<any>;
  InputProps: {
    onBlur?: React.FocusEventHandler<HTMLInputElement | HTMLTextAreaElement>;
    onChange?: React.ChangeEventHandler<HTMLInputElement | HTMLTextAreaElement>;
    onFocus?: React.FocusEventHandler<HTMLInputElement | HTMLTextAreaElement>;
  };
  fullWidth: boolean;
  autoFocus: boolean;
  label: string;
  placeholder: string;
  InputLabelProps: InputLabelProps;
  variant: 'outlined';
  inputProps: object;
}

function renderInput(inputProps: RenderInputProps) {
  const { InputProps, ref, ...other } = inputProps;

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
    <ListItemButton {...itemProps} key={suggestion.label} selected={isSelected} style={style}>
      <ListItemText primary={suggestion.label} secondary={getUtcLabel(suggestion.utcDiff)} />
    </ListItemButton>
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

interface LazyListProps {
  options: string[];
  itemData: any;
}

interface LazyListProps {
  //@ts-expect-error as type mismatch
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
    <Box>
      <FormControl fullWidth>
        <InputLabel variant="standard" shrink>
          Timezone
        </InputLabel>
        <Input
          onClick={onInputActivate}
          value={selectedTimezone}
          slotProps={{
            input: {
              sx: {
                cursor: 'pointer',
              },
            },
          }}
          placeholder={t('common|pick_timezone')}
          readOnly
          data-testid="timezone-readonly-input"
        />
      </FormControl>
      <Dialog open={showPicker}>
        <DialogTitle>Choose a Timezone</DialogTitle>
        <DialogContent>
          <Downshift
            id="downshift-options"
            onChange={selectedItem => {
              if (selectedItem) {
                setSelectedTimezone(selectedItem);
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
              const { onBlur, onChange, ...inputProps } = getInputProps();

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
                    label: t('common|timezone_label'),
                    placeholder: t('common|search_timezone_prompt'),
                    InputLabelProps: getLabelProps(),
                    InputProps: { onBlur, onChange },
                    inputProps,
                    variant: 'outlined',
                  })}
                  <LazyList
                    //@todo add better types
                    //@ts-expect-error as type mismatch
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
                </Box>
              );
            }}
          </Downshift>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleClose} color="primary">
            {t('frequent|cancel')}
          </Button>
          <Button onClick={handleSelect} color="primary">
            {t('frequent|select')}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
