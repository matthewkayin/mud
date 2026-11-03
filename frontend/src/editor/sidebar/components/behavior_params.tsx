import { useSyncExternalStore } from 'react';
import {
  Typography,
  Stack,
  Select,
  MenuItem,
  FormControlLabel,
  Checkbox,
} from '@mui/material';
import { world } from '../../api/models';
import { editorStore, DIRECTION_NAMES } from '../../store';
import type { BehaviorParamTypes, BehaviorParamValue, BehaviorParams } from './default_behavior_params';
import { SubmitTextField } from './submit_textfield';
import { SubmitNumberField } from './submit_number_field';

type BehaviorParamsEditorProps = {
  paramTypes: BehaviorParamTypes;
  params: BehaviorParams;
  onEdit: (params: BehaviorParams) => void;
}

export function BehaviorParamsEditor({ paramTypes, params, onEdit }: BehaviorParamsEditorProps) {
  const itemData = useSyncExternalStore(editorStore.subscribe, editorStore.getItemData);

  const editParam = (name: string, value: BehaviorParamValue) => {
    const editedParams = structuredClone(params);
    editedParams[name] = value;
    onEdit(editedParams);
  };

  // Sorted so that the params keep the same order between renders
  const paramEntries = Object.entries(paramTypes).sort(([a], [b]) => a.localeCompare(b));
  if (paramEntries.length === 0) {
    return null;
  }

  return (
    <Stack spacing={2}>
      <Typography>-- Behavior Params --</Typography>
      {paramEntries.map(([name, paramType]) => {
        const value = params[name];
        switch (paramType) {
          case world.BehaviorParamType.STRING:
            return (
              <SubmitTextField
                key={name}
                label={name}
                value={value as string}
                onSubmit={(newValue) => editParam(name, newValue)}
              />
            );

          case world.BehaviorParamType.NUMBER:
            return (
              <SubmitNumberField
                key={name}
                label={name}
                value={value as number}
                onSubmit={(newValue) => editParam(name, newValue)}
              />
            );

          case world.BehaviorParamType.BOOLEAN:
            return (
              <FormControlLabel
                key={name}
                label={name}
                control={
                  <Checkbox
                    checked={value as boolean}
                    onChange={(event) => editParam(name, event.target.checked)}
                  />
                }
              />
            );

          case world.BehaviorParamType.ITEM: {
            const item = value as world.Item;
            return (
              <Stack key={name} direction="row" spacing={2} sx={{ alignItems: 'center' }}>
                <Typography>{name}:</Typography>
                <Select
                  size="small"
                  value={itemData.some((data) => data.Name === item.Id) ? item.Id : ''}
                  onChange={(event) => editParam(name, world.Item.createFrom({ ...item, Id: event.target.value }))}
                  sx={{ minWidth: '45%' }}
                >
                  {itemData.map((item) => (
                    <MenuItem key={item.Name} value={item.Name}>{item.Name}</MenuItem>
                  ))}
                </Select>
                <SubmitNumberField
                  label="Amount"
                  min={1}
                  value={item.Amount}
                  onSubmit={(amount) => editParam(name, world.Item.createFrom({ ...item, Amount: amount }))}
                />
              </Stack>
            );
          }

          case world.BehaviorParamType.DIRECTION:
            return (
              <Stack key={name} direction="row" spacing={2} sx={{ alignItems: 'center' }}>
                <Typography>{name}:</Typography>
                <Select
                  size="small"
                  value={value as string}
                  onChange={(event) => editParam(name, event.target.value)}
                  sx={{ minWidth: '45%' }}
                >
                  {Object.values(DIRECTION_NAMES).map((direction) => (
                    <MenuItem key={direction} value={direction}>{direction}</MenuItem>
                  ))}
                </Select>
              </Stack>
            );

          default:
            return null;
        }
      })}
    </Stack>
  );
}
