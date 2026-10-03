import { Box, Stack, Typography, FormControlLabel, Checkbox, IconButton } from '@mui/material';
import { useSyncExternalStore } from 'react';
import { editorStore } from '../store';
import { world } from '../../wailsjs/go/models';
import { SubmitTextField } from './components/submit_textfield';
import { DropTableEditor } from './components/drop_table';
import { ChestsEditor } from './components/chests';
import { EditorActionDeleteRoom, EditorActionEditRoom } from '../store/action';
import DeleteIcon from '@mui/icons-material/Delete';

export function Sidebar() {
  const selectedCell = useSyncExternalStore(editorStore.subscribe, () => editorStore.getSelectedCell());
  const room = useSyncExternalStore(editorStore.subscribe, () => editorStore.getSelectedRoom());
  const connections = useSyncExternalStore(editorStore.subscribe, () => editorStore.getSelectedRoomConnections());

  const sidebarBoxSx = {
    width: '25%',
    minWidth: '25%',
    height: '100%',
    backgroundColor: 'background.default',
    borderRight: '2px solid',
    borderColor: 'divider',
    padding: '4px',
    overflowY: 'auto'
  };

  if (!room || !selectedCell || !connections) {
    return (
      <Box sx={sidebarBoxSx}>
        <Stack spacing={2}>
          <Typography>Room: </Typography>
          <Typography>None Selected</Typography>
        </Stack>
      </Box>
    )
  }

  const commitRoomEdit = (edit: (editedRoom: world.Room) => void) => {
    const editedRoom = structuredClone(room);
    edit(editedRoom);

    editorStore.doAction(new EditorActionEditRoom({
      cell: selectedCell,
      previous: structuredClone(room),
      value: editedRoom,
    }));
  };

  return (
    <Box sx={sidebarBoxSx}>
      <Stack spacing={2}>
        <Stack direction="row" spacing={2} sx={{ alignItems: 'center' }}>
          <Typography>Room: </Typography>
          <Box sx={{ flexGrow: 1 }}/>
          <IconButton onClick={() => {
            const action = new EditorActionDeleteRoom({
              cell: selectedCell,
              room: structuredClone(room),
              connections: structuredClone(connections),
            });
            editorStore.doAction(action)
          }}>
            <DeleteIcon />
          </IconButton>
        </Stack>

        <SubmitTextField
          label="Name"
          value={room.Name}
          onSubmit={(value) => {
            commitRoomEdit((editedRoom) => {
              editedRoom.Name = value;
            });
          }}
        />

        <SubmitTextField
          label="Description"
          value={room.Description}
          multiline
          onSubmit={(value) => {
            commitRoomEdit((editedRoom) => {
              editedRoom.Description = value;
            });
          }}
        />

        <FormControlLabel
          label="Safe Zone"
          control={
            <Checkbox
              checked={room.IsSafeZone}
              onChange={(event) => {
                commitRoomEdit((editedRoom) => {
                  editedRoom.IsSafeZone = event.target.checked;
                });
              }}
            />
          }
        />

        <DropTableEditor
          name="Room Items"
          dropTable={room.DropTable}
          onEdit={(dropTable) => {
            commitRoomEdit((editedRoom) => {
              editedRoom.DropTable = dropTable;
            });
          }}
        />

        <ChestsEditor
          chests={room.Chests}
          onEdit={(chests) => {
            commitRoomEdit((editedRoom) => {
              editedRoom.Chests = chests;
            });
          }}
        />
      </Stack>
    </Box>
  )
}
