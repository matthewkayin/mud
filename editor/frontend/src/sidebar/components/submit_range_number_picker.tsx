import { Stack, Typography } from '@mui/material';
import { world } from '../../../wailsjs/go/models';
import { SubmitNumberField } from './submit_number_field';

type SubmitRangeNumberPickerProps = {
  min?: number;
  max?: number;
  value: world.Int32Range;
  onSubmit: (value: world.Int32Range) => void;
}

export function SubmitRangeNumberPicker({ min, max, value, onSubmit }: SubmitRangeNumberPickerProps) {
  return (
    <Stack direction="row" spacing={1} sx={{ alignItems: 'center' }}>
      <Typography>Min:</Typography>
      <SubmitNumberField
        min={min}
        max={value.Max}
        value={value.Min}
        onSubmit={(newMin) => onSubmit({ Min: newMin, Max: value.Max })}
      />

      <Typography>Max:</Typography>
      <SubmitNumberField
        min={value.Min}
        max={max}
        value={value.Max}
        onSubmit={(newMax) => onSubmit({ Min: value.Min, Max: newMax })}
      />
    </Stack>
  );
}
