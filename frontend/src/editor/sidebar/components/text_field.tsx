import { TextField as MuiTextField, type SxProps, type Theme } from '@mui/material';

type TextFieldProps = {
  label: string;
  value: string;
  onChange: (event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement, Element>) => void;
  onSubmit: () => void;
  multiline?: boolean
  sx?: SxProps<Theme>;
}

export function TextField({ label, value, onChange, onSubmit, multiline, sx }: TextFieldProps) {
  const onTextFieldKeydown = (event) => {
    if (event.key === 'Enter') {
      onSubmit();
    }
  };

  return (
    <MuiTextField
        label={label}
        value={value}
        onChange={onChange}
        onKeyDown={onTextFieldKeydown}
        onBlur={onSubmit}
        multiline={multiline}
        sx={sx}
        variant="standard"
    />
  )
}
