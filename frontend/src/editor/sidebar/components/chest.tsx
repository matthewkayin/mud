import * as mud from '../../../mud/types';
import { Stack, IconButton, Typography } from '@mui/material';
import { TextField } from './text_field';
import { DropTableCluster } from './drop_table';
import DeleteIcon from '@mui/icons-material/Delete';
import NumberField from './number_field';

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

      <Stack direction="row" spacing={1} sx={{ alignItems: 'center' }}>
        <Typography>Refresh Duration (Minutes):</Typography>
        <NumberField
          min={0}
          step={1}
          value={Math.floor((chest.RespawnDuration / 60) / mud.WORLD_SECONDS_PER_UPDATE)}
          onValueChange={(value) => {
            const editedChest = structuredClone(chest);
            editedChest.RespawnDuration = value * 60 * mud.WORLD_SECONDS_PER_UPDATE;
            onEdit(editedChest);
          }}
        />
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
