import { useSyncExternalStore } from 'react';
import {
  Stack,
  Typography,
  Select,
  MenuItem,
  Button,
  Accordion,
  AccordionSummary,
  AccordionDetails,
  Divider,
  IconButton,
  Box,
} from '@mui/material';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import DeleteIcon from '@mui/icons-material/Delete';
import { world } from '../../api/models';
import { editorStore } from '../../store';
import { SubmitNumberField } from './submit_number_field';
import { SubmitRangeNumberPicker } from './submit_range_number_picker';

type DropTableEditorProps = {
  name: string;
  dropTable: world.DropTable;
  onEdit: (dropTable: world.DropTable) => void;
}

export function DropTableEditor({ name, dropTable, onEdit }: DropTableEditorProps) {
  const itemData = useSyncExternalStore(editorStore.subscribe, editorStore.getItemData);

  const editEntry = (index: number, edit: (entry: world.DropTableEntry) => void) => {
    const editedTable = structuredClone(dropTable);
    edit(editedTable.Entries[index]);
    onEdit(editedTable);
  };

  return (
    <Accordion slotProps={{ transition: { unmountOnExit: true } }}>
      <AccordionSummary expandIcon={<ExpandMoreIcon/>}>
        <Typography>{name}</Typography>
      </AccordionSummary>
      <AccordionDetails>
        <Stack spacing={2}>
          {dropTable.Entries.map((entry, index) => (
            <Stack key={index} spacing={2}>
              <Stack direction="row" spacing={2} sx={{ alignItems: 'center' }}>
                <Typography>Item:</Typography>
                <Select
                  size="small"
                  value={itemData.some((item) => item.Name === entry.ItemId) ? entry.ItemId : ''}
                  onChange={(event) => {
                    editEntry(index, (editedEntry) => {
                      editedEntry.ItemId = event.target.value;
                    });
                  }}
                  sx={{ minWidth: '45%' }}
                >
                  {itemData.map((item) => (
                    <MenuItem key={item.Name} value={item.Name}>{item.Name}</MenuItem>
                  ))}
                </Select>
                <Box sx={{ marginLeft: 'auto' }}>
                  <IconButton
                    aria-label="Delete entry"
                    onClick={() => {
                      const editedTable = structuredClone(dropTable);
                      editedTable.Entries.splice(index, 1);
                      onEdit(editedTable);
                    }}
                  >
                    <DeleteIcon/>
                  </IconButton>
                </Box>
              </Stack>

              <Stack direction="row" spacing={1} sx={{ alignItems: 'center' }}>
                <Typography>Drop Chance:</Typography>
                <SubmitNumberField
                  min={1}
                  max={100}
                  value={entry.DropChancePercent}
                  onSubmit={(value) => {
                    editEntry(index, (editedEntry) => {
                      editedEntry.DropChancePercent = value;
                    });
                  }}
                />
                <Typography>%</Typography>
              </Stack>

              <Typography>Amount:</Typography>
              <SubmitRangeNumberPicker
                min={1}
                value={entry.AmountRange}
                onSubmit={(value) => {
                  editEntry(index, (editedEntry) => {
                    editedEntry.AmountRange = value;
                  });
                }}
              />

              <Typography>Durability Percent:</Typography>
              <SubmitRangeNumberPicker
                min={1}
                max={100}
                value={entry.DurabilityPercentRange}
                onSubmit={(value) => {
                  editEntry(index, (editedEntry) => {
                    editedEntry.DurabilityPercentRange = value;
                  });
                }}
              />
              <Divider/>
            </Stack>
          ))}

          <Button disabled={itemData.length === 0} onClick={() => {
            const editedTable = structuredClone(dropTable);
            editedTable.Entries.push(world.DropTableEntry.createFrom({
              ItemId: itemData[0].Name,
              AmountRange: { Min: 1, Max: 1 },
              DurabilityPercentRange: { Min: 100, Max: 100 },
              DropChancePercent: 100,
            }));
            onEdit(editedTable);
          }}>+ Add Item</Button>
        </Stack>
      </AccordionDetails>
    </Accordion>
  );
}
