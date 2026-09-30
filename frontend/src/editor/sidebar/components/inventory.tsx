import * as mud from '../../../mud/types';
import NumberField from './number_field';
import {
  Stack,
  Typography,
  Select,
  MenuItem,
  Button,
} from '@mui/material';

type InventoryClusterProps = {
  inventory: mud.Inventory;
  itemData: mud.ItemData[];
  onAddItem: () => void;
  onEditItem: (itemIndex: number, item: mud.Item) => void;
}

export function InventoryCluster({ inventory, itemData, onAddItem, onEditItem }: InventoryClusterProps) {
  return (
    <Stack spacing={2}>
      {inventory.Items.map((item: mud.Item, itemIndex: number) => (
        <>
          <Stack
            direction="row"
            spacing={1}
            sx={{
              alignItems: 'center',
            }}
          >
            <Typography>ID:</Typography>
            <Select
              label="Item"
              value={item.Id}
              onChange={(event) => onEditItem(itemIndex, { ...item, Id: event.target.value })}
            >{itemData.map((entry, index) => (
              <MenuItem value={index}>{entry.Name}</MenuItem>
            ))}
            </Select>

            <Typography>ID:</Typography>
            <NumberField
              min={1}
              value={item.Amount}
              size="small"
              onValueChange={(value) => onEditItem(itemIndex, { ...item, Amount: value })}
            />
          </Stack>
        </>
      ))}
      <Button onClick={onAddItem}>+ Add Item</Button>
    </Stack>
  )
}
