import { useEffect, useState } from 'react';
import { TextField, type SxProps, type Theme } from '@mui/material';

type SubmitTextFieldProps = {
  label: string;
  value: string;
  onSubmit: (value: string) => void;
  multiline?: boolean;
  sx?: SxProps<Theme>;
}

export function SubmitTextField({ label, value, onSubmit, multiline, sx }: SubmitTextFieldProps) {
  const [draft, setDraft] = useState(value);

  useEffect(() => {
    setDraft(value);
  }, [value]);

  const submit = () => {
    if (draft === value) {
      return;
    }
    onSubmit(draft);
  }

  return (
    <TextField
      label={label}
      size="small"
      value={draft}
      multiline={multiline}
      sx={sx}
      onChange={(event) => setDraft(event.target.value)}
      onBlur={submit}
      onKeyDown={(event) => {
        // Multiline fields use Enter for newlines, so they only submit on blur
        if (event.key === 'Enter' && !multiline) {
          event.preventDefault();
          submit();
        }
      }}
    />
  )
}
