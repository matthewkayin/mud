import { useSyncExternalStore } from "react";
import { world } from "../api/models";
import { editorStore, EditorCell, EditorConnection, EditorActionSetExitLock, DIRECTION_NAMES, LOCK_TYPE_NAMES } from "../store";
import { Box, Stack, Typography, IconButton, Select, MenuItem } from '@mui/material';
import DeleteIcon from '@mui/icons-material/Delete';

const LOCK_TYPE_NONE = 'none';

type ConnectionEditorProps = {
  selectedConnection: EditorConnection;
}

export function ConnectionEditor({ selectedConnection }: ConnectionEditorProps) {
  const direction = selectedConnection.getDirection() as Exclude<world.Direction, world.Direction.COUNT>;
  const reverseDirection = ((direction + 2) % world.Direction.COUNT) as Exclude<world.Direction, world.Direction.COUNT>;

  return (
    <Stack spacing={2}>
      <Stack direction="row" spacing={2} sx={{ alignItems: 'center' }}>
        <Typography>Connection</Typography>
        <Box sx={{ flexGrow: 1 }}/>
        <IconButton>
          <DeleteIcon />
        </IconButton>
      </Stack>

      <Typography>Locks:</Typography>
      <ExitLockEditor cell={selectedConnection.from} direction={direction} />
      <ExitLockEditor cell={selectedConnection.to} direction={reverseDirection} />
    </Stack>
  )
}

type ExitLockEditorProps = {
  cell: EditorCell;
  direction: Exclude<world.Direction, world.Direction.COUNT>;
}

// Edits the lock on one side of a connection. Locks are one-directional, so each side has its own.
function ExitLockEditor({ cell, direction }: ExitLockEditorProps) {
  const roomName = useSyncExternalStore(editorStore.subscribe, () => editorStore.getRooms().get(cell.toString())?.Name);
  const lockType = useSyncExternalStore(editorStore.subscribe, () => editorStore.getExitLockType(cell, direction));

  return (
    <Stack direction="row" spacing={2} sx={{ alignItems: 'center' }}>
      <Typography>{roomName}, {DIRECTION_NAMES[direction]} exit:</Typography>
      <Select
        size="small"
        value={lockType ?? LOCK_TYPE_NONE}
        onChange={(event) => {
          const value = event.target.value;
          editorStore.doAction(new EditorActionSetExitLock({
            cell: cell,
            direction: direction,
            lockType: value === LOCK_TYPE_NONE ? null : value as world.LockType,
            previousLockType: lockType,
          }));
        }}
        sx={{ minWidth: '35%', marginLeft: 'auto' }}
      >
        <MenuItem value={LOCK_TYPE_NONE}>None</MenuItem>
        {Object.entries(LOCK_TYPE_NAMES).map(([value, name]) => (
          <MenuItem key={value} value={Number(value)}>{name}</MenuItem>
        ))}
      </Select>
    </Stack>
  )
}
