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

// Returns a value for each behavior param that the NPC data declares
export function defaultBehaviorParams(npcData: world.NpcData, itemData: world.ItemData[]): Record<string, any> {
  const params: Record<string, any> = {};
  for (const [name, paramType] of Object.entries(npcData.BehaviorParams)) {
    switch (paramType) {
      case world.NpcBehaviorParamType.STRING:
        params[name] = '';
        break;
      case world.NpcBehaviorParamType.NUMBER:
        params[name] = 0;
        break;
      case world.NpcBehaviorParamType.BOOLEAN:
        params[name] = false;
        break;
      case world.NpcBehaviorParamType.ITEM:
        params[name] = world.Item.createFrom({
          Id: itemData.length > 0 ? itemData[0].Name : '',
          Amount: 1,
          Durability: 0,
        });
        break;
      case world.NpcBehaviorParamType.DIRECTION:
        params[name] = DIRECTION_NAMES[world.Direction.NORTH];
        break;
    }
  }

  return params;
}

type BehaviorParamsEditorProps = {
  npcData: world.NpcData;
  params: Record<string, any>;
  onEdit: (params: Record<string, any>) => void;
}

export function BehaviorParamsEditor({ npcData, params, onEdit }: BehaviorParamsEditorProps) {
  const itemData = useSyncExternalStore(editorStore.subscribe, editorStore.getItemData);

  const editParam = (name: string, value: any) => {
    const editedParams = structuredClone(params);
    editedParams[name] = value;
    onEdit(editedParams);
  };

  // Sorted so that the params keep the same order between renders
  const paramEntries = Object.entries(npcData.BehaviorParams).sort(([a], [b]) => a.localeCompare(b));
  if (paramEntries.length === 0) {
    return null;
  }

  return (
    <Stack spacing={2}>
      <Typography>Behavior:</Typography>
      {paramEntries.map(([name, paramType]) => {
        const value = params[name];
        switch (paramType) {
          case world.NpcBehaviorParamType.STRING:
            return (
              <SubmitTextField
                key={name}
                label={name}
                value={value}
                onSubmit={(newValue) => editParam(name, newValue)}
              />
            );

          case world.NpcBehaviorParamType.NUMBER:
            return (
              <SubmitNumberField
                key={name}
                label={name}
                value={value}
                onSubmit={(newValue) => editParam(name, newValue)}
              />
            );

          case world.NpcBehaviorParamType.BOOLEAN:
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

          case world.NpcBehaviorParamType.ITEM:
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

          case world.NpcBehaviorParamType.DIRECTION:
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
