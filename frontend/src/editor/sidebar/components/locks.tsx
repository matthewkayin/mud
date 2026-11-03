import {
  Accordion,
  AccordionSummary,
  AccordionDetails,
  Button,
  Typography,
  Stack,
  IconButton,
  Select,
  MenuItem,
} from '@mui/material';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import DeleteIcon from '@mui/icons-material/Delete';
import { world } from '../../api/models';
import { DIRECTION_NAMES, LOCK_TYPE_NAMES } from '../../store';

type LockDirection = Exclude<world.Direction, world.Direction.COUNT>;

type LocksEditorProps = {
  locks: world.Lock[];
  // The directions the room has exits in. Locks can only be placed on these.
  connectedDirections: world.Direction[];
  onEdit: (locks: world.Lock[]) => void;
}

// Locks are one-directional, so a lock here only blocks leaving this room through its exit
export function LocksEditor({ locks, connectedDirections, onEdit }: LocksEditorProps) {
  const editLock = (index: number, edit: (lock: world.Lock) => void) => {
    const editedLocks = structuredClone(locks);
    edit(editedLocks[index]);
    onEdit(editedLocks);
  };

  const lockedDirections = locks.map((lock) => lock.Direction);
  const unlockedDirections = connectedDirections.filter((direction) => !lockedDirections.includes(direction));

  return (
    <Accordion slotProps={{ transition: { unmountOnExit: true } }}>
      <AccordionSummary expandIcon={<ExpandMoreIcon/>}>
        <Typography>Locks</Typography>
      </AccordionSummary>
      <AccordionDetails>
        <Stack spacing={2}>
          {locks.map((lock, index) => {
            // A lock can use its own direction or any connected direction that isn't locked yet
            const directionOptions = [lock.Direction, ...unlockedDirections].sort((a, b) => a - b) as LockDirection[];
            const hasExit = connectedDirections.includes(lock.Direction);

            return (
              <Stack key={index} spacing={1}>
                <Stack direction="row" spacing={2} sx={{ alignItems: 'center' }}>
                  <Select
                    size="small"
                    value={lock.Direction}
                    onChange={(event) => {
                      editLock(index, (editedLock) => {
                        editedLock.Direction = event.target.value as world.Direction;
                      });
                    }}
                    sx={{ minWidth: '35%' }}
                  >
                    {directionOptions.map((direction) => (
                      <MenuItem key={direction} value={direction}>{DIRECTION_NAMES[direction]}</MenuItem>
                    ))}
                  </Select>
                  <Select
                    size="small"
                    value={lock.Type}
                    onChange={(event) => {
                      editLock(index, (editedLock) => {
                        editedLock.Type = event.target.value as world.LockType;
                      });
                    }}
                    sx={{ minWidth: '35%' }}
                  >
                    {Object.entries(LOCK_TYPE_NAMES).map(([value, name]) => (
                      <MenuItem key={value} value={Number(value)}>{name}</MenuItem>
                    ))}
                  </Select>
                  <IconButton
                    aria-label="Delete lock"
                    sx={{ marginLeft: 'auto' }}
                    onClick={() => {
                      const editedLocks = structuredClone(locks);
                      editedLocks.splice(index, 1);
                      onEdit(editedLocks);
                    }}
                  >
                    <DeleteIcon/>
                  </IconButton>
                </Stack>
                {!hasExit && <Typography variant="caption" color="warning.main">
                  This room has no {DIRECTION_NAMES[lock.Direction as LockDirection]} exit, so this lock will be removed when the world is saved.
                </Typography>}
              </Stack>
            );
          })}

          <Button disabled={unlockedDirections.length === 0} onClick={() => {
            const editedLocks = structuredClone(locks);
            editedLocks.push(world.Lock.createFrom({
              Type: world.LockType.PHYSICAL,
              Direction: unlockedDirections[0],
            }));
            onEdit(editedLocks);
          }}>+ Add Lock</Button>
        </Stack>
      </AccordionDetails>
    </Accordion>
  );
}
