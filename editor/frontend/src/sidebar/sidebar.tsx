import { Box, Stack, Typography } from '@mui/material';
import { useEffect, useSyncExternalStore } from 'react';
import { editorStore } from '../store';
import { SubmitTextField } from './components/submit_textfield';
import { EditorActionEditRoom } from '../store/action';

export function Sidebar() {
  const roomIndex = useSyncExternalStore(editorStore.subscribe, () => editorStore.getSelectedRoomIndex());
  const room = useSyncExternalStore(editorStore.subscribe, () => editorStore.getSelectedRoom());

  const sidebarBoxSx = {
    width: '25%',
    minWidth: '25%',
    height: '100%',
    backgroundColor: '#f0f0f0',
    borderRight: '2px solid #d6d6d6',
    padding: '4px',
    overflowY: 'auto'
  };

  if (!room || roomIndex === undefined) {
    return (
      <Box sx={sidebarBoxSx}>
        <Stack spacing={2}>
          <Typography>Room: </Typography>
          <Typography>None Selected</Typography>
        </Stack>
      </Box>
    )
  }

  return (
    <Box sx={sidebarBoxSx}>
      <Stack spacing={2}>
        <Typography>Room: </Typography>
        <SubmitTextField
          label="Name"
          value={room.Name}
          onSubmit={(value) => {
            const editedRoom = structuredClone(room);
            editedRoom.Name = value;

            editorStore.doAction(new EditorActionEditRoom({
              roomIndex: roomIndex,
              previous: structuredClone(room),
              value: editedRoom,
            }));
          }}
        />
      </Stack>
    </Box>
  )
}
