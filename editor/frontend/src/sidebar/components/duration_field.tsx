import { Stack, Typography } from '@mui/material';
import { ticksToMinutes, minutesToTicks } from '../../store';
import { SubmitNumberField } from './submit_number_field';

type DurationFieldProps = {
  label: string;
  ticks: number;
  onSubmit: (ticks: number) => void;
}

// Durations are stored in world ticks but edited in whole minutes
export function DurationField({ label, ticks, onSubmit }: DurationFieldProps) {
  return (
    <Stack direction="row" spacing={1} sx={{ alignItems: 'center' }}>
      <Typography>{label}:</Typography>
      <SubmitNumberField
        min={0}
        value={Math.floor(ticksToMinutes(ticks))}
        onSubmit={(minutes) => {
          onSubmit(minutesToTicks(minutes));
        }}
      />
      <Typography>minutes</Typography>
    </Stack>
  );
}
