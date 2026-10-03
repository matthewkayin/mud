// Based on: https://mui.com/material-ui/react-number-field/

import { useEffect, useId, useRef, useState } from 'react';
import { NumberField as BaseNumberField } from '@base-ui/react/number-field';
import {
  Box,
  IconButton,
  FormControl,
  OutlinedInput,
  InputAdornment,
  InputLabel,
} from '@mui/material';
import KeyboardArrowUpIcon from '@mui/icons-material/KeyboardArrowUp';
import KeyboardArrowDownIcon from '@mui/icons-material/KeyboardArrowDown';

type SubmitNumberFieldProps = {
  label?: string;
  value: number;
  min?: number;
  max?: number;
  step?: number;
  disabled?: boolean;
  onSubmit: (value: number) => void;
}

// A number field that keeps edits as a local draft and only submits once focus leaves
// the field (or Enter is pressed), so that stepping or typing doesn't create an action
// per change.
export function SubmitNumberField({ label, value, min, max, step = 1, disabled = false, onSubmit }: SubmitNumberFieldProps) {
  const id = useId();
  const rootRef = useRef<HTMLDivElement>(null);
  const [draft, setDraft] = useState<number | null>(value);
  // Base UI may update the value during the input's own blur handler, before our
  // blur handler runs, so read the latest draft from a ref rather than from state
  const draftRef = useRef<number | null>(value);

  useEffect(() => {
    draftRef.current = value;
    setDraft(value);
  }, [value]);

  const setDraftValue = (newValue: number | null) => {
    draftRef.current = newValue;
    setDraft(newValue);
  };

  const submit = () => {
    if (draftRef.current === null) {
      setDraftValue(value);
      return;
    }

    let newValue = Math.floor(draftRef.current);
    if (min !== undefined && newValue < min) {
      newValue = min;
    }
    if (max !== undefined && newValue > max) {
      newValue = max;
    }

    setDraftValue(newValue);
    if (newValue !== value) {
      onSubmit(newValue);
    }
  };

  return (
    <Box
      ref={rootRef}
      onBlur={(event) => {
        // Moving focus between the input and the step buttons is not leaving the field
        if (rootRef.current && rootRef.current.contains(event.relatedTarget as Node | null)) {
          return;
        }
        submit();
      }}
      onKeyDown={(event) => {
        if (event.key === 'Enter') {
          event.preventDefault();
          submit();
        }
      }}
    >
      <BaseNumberField.Root
        value={draft}
        min={min}
        max={max}
        step={step}
        disabled={disabled}
        onValueChange={setDraftValue}
        render={(props, state) => (
          <FormControl
            size="small"
            ref={props.ref}
            disabled={state.disabled}
            required={state.required}
            variant="outlined"
          >
            {props.children}
          </FormControl>
        )}
      >
        {label && <InputLabel htmlFor={id}>{label}</InputLabel>}
        <BaseNumberField.Input
          id={id}
          render={(props, state) => (
            <OutlinedInput
              label={label}
              inputRef={props.ref}
              value={state.inputValue}
              onBlur={props.onBlur}
              onChange={props.onChange}
              onKeyUp={props.onKeyUp}
              onKeyDown={props.onKeyDown}
              onFocus={props.onFocus}
              slotProps={{
                input: props,
              }}
              endAdornment={
                <InputAdornment
                  position="end"
                  sx={{
                    flexDirection: 'column',
                    maxHeight: 'unset',
                    alignSelf: 'stretch',
                    borderLeft: '1px solid',
                    borderColor: 'divider',
                    ml: 0,
                    '& button': {
                      py: 0,
                      flex: 1,
                      borderRadius: 0.5,
                    },
                  }}
                >
                  <BaseNumberField.Increment
                    render={<IconButton size="small" aria-label="Increase" />}
                  >
                    <KeyboardArrowUpIcon
                      fontSize="small"
                      sx={{ transform: 'translateY(2px)' }}
                    />
                  </BaseNumberField.Increment>

                  <BaseNumberField.Decrement
                    render={<IconButton size="small" aria-label="Decrease" />}
                  >
                    <KeyboardArrowDownIcon
                      fontSize="small"
                      sx={{ transform: 'translateY(-2px)' }}
                    />
                  </BaseNumberField.Decrement>
                </InputAdornment>
              }
              sx={{ pr: 0 }}
            />
          )}
        />
      </BaseNumberField.Root>
    </Box>
  );
}
