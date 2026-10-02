import { useEffect, useState } from 'react';
import { TextField } from '@mui/material';

type SubmitTextFieldProps = {
  label: string;
  value: string;
  onSubmit: (value: string) => void;
}

export function SubmitTextField({ label, value, onSubmit }: SubmitTextFieldProps) {
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
      onChange={(event) => setDraft(event.target.value)}
      onBlur={submit}
      onKeyDown={(event) => {
        if (event.key === 'Enter') {
          event.preventDefault();
          submit();
        }
      }}
    />
  )
}
