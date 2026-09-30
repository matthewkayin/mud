import * as mud from '../../../mud/types';
import { useState } from 'react';
import { Stack } from '@mui/material';
import { TextField } from './text_field';
import { DropTableCluster } from './drop_table';

type ChestClusterProps = {
  chest: mud.Chest;
  itemData: mud.ItemData[];
  onEdit: (editedChest: mud.Chest) => void;
}

export function ChestCluster({ chest, itemData, onEdit }: ChestClusterProps) {
  const [chestName, setChestName] = useState(chest.Name);

  return (
    <Stack spacing={2}>
      <TextField
        label="Name"
        value={chest.Name}
        onChange={(event) => {
          setChestName(event.target.value)
        }}
        onSubmit={() => {
          const editedChest = structuredClone(chest);
          editedChest.Name = chestName;
          onEdit(editedChest);
        }}
      />
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
