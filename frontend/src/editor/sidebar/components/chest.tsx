import * as mud from '../../../mud/types';
import { Stack, IconButton } from '@mui/material';
import { TextField } from './text_field';
import { DropTableCluster } from './drop_table';
import DeleteIcon from '@mui/icons-material/Delete';

type ChestClusterProps = {
  chest: mud.Chest;
  itemData: mud.ItemData[];
  onNameEdit: (name: string) => void;
  onEdit: (editedChest: mud.Chest) => void;
  onDelete: () => void;
}

export function ChestCluster({ chest, itemData, onNameEdit, onEdit, onDelete }: ChestClusterProps) {
  return (
    <Stack spacing={2}>
      <Stack direction="row" spacing={1} sx={{ alignItems: 'center' }}>
        <TextField
          label="Name"
          value={chest.Name}
          onChange={(event) => {
            onNameEdit(event.target.value);
          }}
          onSubmit={() => {
            const editedChest = structuredClone(chest);
            onEdit(editedChest);
          }}
          sx={{
            width: '90%'
          }}
        />
        <IconButton
          onClick={onDelete}
        >
          <DeleteIcon/>
        </IconButton>
      </Stack>

      <DropTableCluster
        name="Items"
        dropTable={chest.DropTable}
        itemData={itemData}
        onEdit={(dropTable: mud.DropTable) => {
          const editedChest = structuredClone(chest);
          editedChest.DropTable = dropTable;
          onEdit(editedChest);
        }}
      />
    </Stack>
  )
}
