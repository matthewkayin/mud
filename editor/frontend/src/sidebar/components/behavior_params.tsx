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
import { SubmitTextField } from './submit_textfield';
import { SubmitNumberField } from './submit_number_field';

// Maps each behavior param name that an NPC or fixture script declares to its type
type BehaviorParamTypes = Record<string, world.BehaviorParamType>;

// Returns a value for each declared behavior param
export function defaultBehaviorParams(paramTypes: BehaviorParamTypes, itemData: world.ItemData[]): Record<string, any> {
  const params: Record<string, any> = {};
  for (const [name, paramType] of Object.entries(paramTypes)) {
    switch (paramType) {
      case world.BehaviorParamType.STRING:
        params[name] = '';
        break;
      case world.BehaviorParamType.NUMBER:
        params[name] = 0;
        break;
      case world.BehaviorParamType.BOOLEAN:
        params[name] = false;
        break;
      case world.BehaviorParamType.ITEM:
        params[name] = world.Item.createFrom({
          Id: itemData.length > 0 ? itemData[0].Name : '',
          Amount: 1,
          Durability: 0,
        });
        break;
      case world.BehaviorParamType.DIRECTION:
        params[name] = DIRECTION_NAMES[world.Direction.NORTH];
        break;
    }
  }

  return params;
}

type BehaviorParamsEditorProps = {
  paramTypes: BehaviorParamTypes;
  params: Record<string, any>;
  onEdit: (params: Record<string, any>) => void;
}

export function BehaviorParamsEditor({ paramTypes, params, onEdit }: BehaviorParamsEditorProps) {
  const itemData = useSyncExternalStore(editorStore.subscribe, editorStore.getItemData);

  const editParam = (name: string, value: any) => {
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
                value={value}
                onSubmit={(newValue) => editParam(name, newValue)}
              />
            );

          case world.BehaviorParamType.NUMBER:
            return (
              <SubmitNumberField
                key={name}
                label={name}
                value={value}
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
                    checked={value}
                    onChange={(event) => editParam(name, event.target.checked)}
                  />
                }
              />
            );

          case world.BehaviorParamType.ITEM:
            return (
              <Stack key={name} direction="row" spacing={2} sx={{ alignItems: 'center' }}>
                <Typography>{name}:</Typography>
                <Select
                  size="small"
                  value={itemData.some((item) => item.Name === value.Id) ? value.Id : ''}
                  onChange={(event) => editParam(name, { ...value, Id: event.target.value })}
                  sx={{ minWidth: '45%' }}
                >
                  {itemData.map((item) => (
                    <MenuItem key={item.Name} value={item.Name}>{item.Name}</MenuItem>
                  ))}
                </Select>
                <SubmitNumberField
                  label="Amount"
                  min={1}
                  value={value.Amount}
                  onSubmit={(amount) => editParam(name, { ...value, Amount: amount })}
                />
              </Stack>
            );

          case world.BehaviorParamType.DIRECTION:
            return (
              <Stack key={name} direction="row" spacing={2} sx={{ alignItems: 'center' }}>
                <Typography>{name}:</Typography>
                <Select
                  size="small"
                  value={value}
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
