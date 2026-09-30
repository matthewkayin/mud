import * as mud from '../../../mud/types';
import NumberField from './number_field';
import {
  Stack,
  Typography,
  Select,
  MenuItem,
  Button,
  Box,
  Accordion,
  AccordionSummary,
  AccordionDetails,
} from '@mui/material';
import { RangeNumberPicker } from './range_number_picker';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';

type DropTableClusterProps = {
  name: string;
  dropTable: mud.DropTable;
  itemData: mud.ItemData[];
  onEdit: (dropTable: mud.DropTable) => void;
}

export function DropTableCluster({ name, dropTable, itemData, onEdit }: DropTableClusterProps) {
  const itemIdDropdownOptions = itemData.map((entry, index) => (
    <MenuItem value={index}>{entry.Name}</MenuItem>
  ));

  return (
    <Accordion>
      <AccordionSummary expandIcon={<ExpandMoreIcon />}>
        <Typography>{name}</Typography>
      </AccordionSummary>
      <AccordionDetails>
        <Stack spacing={2}>
          {dropTable.Entries.map((entry: mud.DropTableEntry, index: number) => (
            <Box sx={{borderBottom: '1px solid #555', paddingBottom: '16px' }}>
              <Stack direction="row" spacing={1} sx={{ alignItems: 'center' }}>
                <Typography>ID:</Typography>
                <Select
                  label="Item"
                  value={entry.ItemId}
                  onChange={(event) => {
                    const editedTable = structuredClone(dropTable);
                    editedTable.Entries[index].ItemId = event.target.value;
                    onEdit(editedTable);
                  }}
                  sx={{
                    width: '45%',
                  }}
                >
                  {itemIdDropdownOptions}
                </Select>

                <Typography>Chance:</Typography>
                <NumberField
                  min={1}
                  max={100}
                  step={1}
                  value={entry.DropChancePercent}
                  size="small"
                  onValueChange={(value) => {
                    const editedTable = structuredClone(dropTable);
                    editedTable.Entries[index].DropChancePercent = Math.floor(value);
                    onEdit(editedTable);
                  }}
                />
                <Typography>%</Typography>
              </Stack>

              <Typography>Amount:</Typography>
              <RangeNumberPicker
                min={1}
                value={entry.AmountRange}
                onValueChange={(value) => {
                  const editedTable = structuredClone(dropTable);
                  editedTable.Entries[index].AmountRange = value;
                  onEdit(editedTable);
                }}
              />

              <Typography>Durability Percent:</Typography>
              <RangeNumberPicker
                min={1}
                max={100}
                value={entry.DurabilityPercentRange}
                onValueChange={(value) => {
                  const editedTable = structuredClone(dropTable);
                  editedTable.Entries[index].DurabilityPercentRange = value;
                  onEdit(editedTable);
                }}
              />
            </Box>
          ))}

          <Button onClick={() => {
            const editedTable = structuredClone(dropTable);
            editedTable.Entries.push({
              ItemId: 0,
              AmountRange: { Min: 1, Max: 1 },
              DurabilityPercentRange: { Min: 100, Max: 100 },
              DropChancePercent: 100,
            })
            onEdit(editedTable);
          }}>+ Add Item</Button>
        </Stack>
      </AccordionDetails>
    </Accordion>
  )
}
