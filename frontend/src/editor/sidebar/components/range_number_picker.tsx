import { Stack, Typography } from '@mui/material';
import NumberField from './number_field';
import * as mud from '../../../mud/types';

type RangeNumberPickerProps = {
  min?: number;
  max?: number;
  value: mud.Int32Range;
  onValueChange: (value: mud.Int32Range) => void;
}

export function RangeNumberPicker({ min, max, value, onValueChange }: RangeNumberPickerProps) {
  return (
    <Stack direction="row" spacing={1} sx={{ alignItems: 'center' }}>
      <Typography>Min:</Typography>
      <NumberField
        min={min}
        max={value.Max}
        step={1}
        value={value.Min}
        size="small"
        onValueChange={(newMin) =>
          onValueChange({
            Min: Math.floor(newMin),
            Max: value.Max,
          })}
      />

      <Typography>Max:</Typography>
      <NumberField
        min={value.Min}
        max={max}
        step={1}
        value={value.Max}
        size="small"
        onValueChange={(newMax) =>
          onValueChange({
            Min: Math.floor(value.Min),
            Max: newMax,
          })}
      />
    </Stack>
  );
}
