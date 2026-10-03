import { Box, Stack, Typography } from '@mui/material';
import { useSyncExternalStore } from 'react';
import { editorStore } from '../store';
import { RoomEditor } from './room';
import { ConnectionEditor } from './connection';

export function Sidebar() {
  const selectedCell = useSyncExternalStore(editorStore.subscribe, () => editorStore.getSelectedCell());
  const selectedConnection = useSyncExternalStore(editorStore.subscribe, () => editorStore.getSelectedConnection());
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

  // Room editor
  if (selectedCell && room && connections) {
    return (
      <Box sx={sidebarBoxSx}>
        <RoomEditor selectedCell={selectedCell} room={room} connections={connections} />
      </Box>
    )
  }

  // Connection editor
  if (selectedConnection) {
    return (
      <Box sx={sidebarBoxSx}>
        <ConnectionEditor selectedConnection={selectedConnection} />
      </Box>
    );
  }

  // No selection
  return (
    <Box sx={sidebarBoxSx}>
      <Stack spacing={2}>
        <Typography>No selection</Typography>
      </Stack>
    </Box>
  )
}
