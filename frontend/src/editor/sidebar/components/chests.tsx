import {
  Accordion,
  AccordionSummary,
  AccordionDetails,
  Button,
  Typography,
  Stack,
  Divider,
  IconButton,
} from '@mui/material';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import DeleteIcon from '@mui/icons-material/Delete';
import { world } from '../../api/models';
import { SubmitTextField } from './submit_textfield';
import { DropTableEditor } from './drop_table';

type ChestsEditorProps = {
  chests: world.Chest[];
  onEdit: (chests: world.Chest[]) => void;
}

export function ChestsEditor({ chests, onEdit }: ChestsEditorProps) {
  const editChest = (index: number, edit: (chest: world.Chest) => void) => {
    const editedChests = structuredClone(chests);
    edit(editedChests[index]);
    onEdit(editedChests);
  };

  return (
    <Accordion slotProps={{ transition: { unmountOnExit: true } }}>
      <AccordionSummary expandIcon={<ExpandMoreIcon/>}>
        <Typography>Chests</Typography>
      </AccordionSummary>
      <AccordionDetails>
        <Stack spacing={2}>
          {chests.map((chest, index) => (
            <Stack key={index} spacing={2}>
              <Stack direction="row" spacing={1} sx={{ alignItems: 'center' }}>
                <SubmitTextField
                  label="Name"
                  value={chest.Name}
                  onSubmit={(name) => {
                    editChest(index, (editedChest) => {
                      editedChest.Name = name;
                    });
                  }}
                  sx={{ flexGrow: 1 }}
                />
                <IconButton
                  aria-label="Delete chest"
                  onClick={() => {
                    const editedChests = structuredClone(chests);
                    editedChests.splice(index, 1);
                    onEdit(editedChests);
                  }}
                >
                  <DeleteIcon/>
                </IconButton>
              </Stack>

              <DropTableEditor
                name="Items"
                dropTable={chest.DropTable}
                onEdit={(dropTable) => {
                  editChest(index, (editedChest) => {
                    editedChest.DropTable = dropTable;
                  });
                }}
              />
              <Divider/>
            </Stack>
          ))}

          <Button onClick={() => {
            const editedChests = structuredClone(chests);
            editedChests.push(world.Chest.createFrom({
              Name: 'New Chest',
              Type: world.ChestType.CHEST,
              Timer: 0,
              DropTable: {
                Entries: [],
              },
              Inventory: {
                Items: [],
              },
            }));
            onEdit(editedChests);
          }}>+ Add Chest</Button>
        </Stack>
      </AccordionDetails>
    </Accordion>
  );
}
