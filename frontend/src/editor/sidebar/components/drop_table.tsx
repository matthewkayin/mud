import * as mud from '../../../mud/types';
import NumberField from './number_field';
import {
  Stack,
  Typography,
  Select,
  MenuItem,
  Button,
} from '@mui/material';

type DropTableClusterProps = {
  dropTable: mud.DropTable;
  itemData: mud.ItemData[];
  onEdit: (dropTable: mud.DropTable) => void;
}

export function DropTableCluster({ dropTable, itemData, onEdit }: DropTableClusterProps) {
  const itemIdDropdownOptions = itemData.map((entry, index) => (
    <MenuItem value={index}>{entry.Name}</MenuItem>
  ));

  return (
    <Stack spacing={2}>
      {dropTable.Entries.map((entry: mud.DropTableEntry, index: number) => (
        <>
          <Stack direction="row" spacing={1} sx={{ alignItems: 'center', }}>
            <Typography>ID:</Typography>
            <Select
              label="Item"
              value={entry.ItemId}
              onChange={(event) => {
                const editedTable = structuredClone(dropTable);
                editedTable.Entries[index].ItemId = event.target.value;
                onEdit(editedTable);
              }}
            >
              {itemIdDropdownOptions}
            </Select>
          </Stack>

          <Stack direction="row" spacing={1} sx={{ alignItems: 'center', }}>
            <Typography>Drop Chance:</Typography>
            <NumberField
              min={1}
              value={entry.DropChance}
              size="small"
              onValueChange={(value) => {
                const editedTable = structuredClone(dropTable);
                editedTable.Entries[index].DropChance = value;
                onEdit(editedTable);
              }}
            />
            <Typography>%</Typography>
          </Stack>

          <Stack direction="row" spacing={1} sx={{ alignItems: 'center', }}>
            <Typography>Min:</Typography>
            <NumberField
              min={1}
              value={entry.AmountRange.Min}
              size="small"
              onValueChange={(value) => {
                const editedTable = structuredClone(dropTable);
                editedTable.Entries[index].AmountRange.Min = value;
                onEdit(editedTable);
              }}
            />
            <Typography>Max:</Typography>
            <NumberField
              min={entry.AmountRange.Min}
              value={entry.AmountRange.Max}
              size="small"
              onValueChange={(value) => {
                const editedTable = structuredClone(dropTable);
                editedTable.Entries[index].AmountRange.Max = value;
                onEdit(editedTable);
              }}
            />
          </Stack>

          <Stack direction="row" spacing={1} sx={{ alignItems: 'center', }}>
            <Typography>Durability:</Typography>
            <NumberField
              min={1}
              value={entry.Durability}
              size="small"
              onValueChange={(value) => {
                const editedTable = structuredClone(dropTable);

                editedTable.Entries[index].Durability = value;
                onEdit(editedTable);
              }}
            />
            <Typography>%</Typography>
          </Stack>
        </>
      ))}

      <Button onClick={() => {
        const editedTable = structuredClone(dropTable);
        editedTable.Entries.push({
          ItemId: 0,
          AmountRange: { Min: 1, Max: 1 },
          Durability: 1.0,
          DropChance: 1.0,
        })
        onEdit(editedTable);
      }}>+ Add Item</Button>
    </Stack>
  )
}
