import { Stack, Typography, Box, IconButton, FormControlLabel, Checkbox } from '@mui/material';
import DeleteIcon from '@mui/icons-material/Delete';
import { SubmitTextField } from './components/submit_textfield';
import { DropTableEditor } from './components/drop_table';
import { ChestsEditor } from './components/chests';
import { editorStore, EditorCell } from '../store';
import { EditorActionDeleteRoom, EditorActionEditRoom } from '../store/action';
import { world } from '../../wailsjs/go/models';

type RoomEditorProps = {
  selectedCell: EditorCell;
  room: world.Room;
  connections: string[];
}

export function RoomEditor({ selectedCell, room, connections }: RoomEditorProps) {
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
  )
}
