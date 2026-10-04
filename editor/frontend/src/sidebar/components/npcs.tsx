import { useSyncExternalStore } from 'react';
import {
  Accordion,
  AccordionSummary,
  AccordionDetails,
  Button,
  Typography,
  Stack,
  Divider,
  IconButton,
  Select,
  MenuItem,
  Box,
} from '@mui/material';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import DeleteIcon from '@mui/icons-material/Delete';
import { world } from '../../api/models';
import {
  editorStore,
  minutesToTicks,
  NPC_DISPOSITION_NAMES,
  NPC_MOVEMENT_TYPE_NAMES,
} from '../../store';
import { getEditorConstants } from '../../store/constants';
import { SubmitRangeNumberPicker } from './submit_range_number_picker';
import { DropTableEditor } from './drop_table';
import { DurationField } from './duration_field';

type NpcsEditorProps = {
  npcs: world.Npc[];
  onEdit: (npcs: world.Npc[]) => void;
}

export function NpcsEditor({ npcs, onEdit }: NpcsEditorProps) {
  const npcData = useSyncExternalStore(editorStore.subscribe, editorStore.getNpcData);

  const editNpc = (index: number, edit: (npc: world.Npc) => void) => {
    const editedNpcs = structuredClone(npcs);
    edit(editedNpcs[index]);
    onEdit(editedNpcs);
  };

  return (
    <Accordion slotProps={{ transition: { unmountOnExit: true } }}>
      <AccordionSummary expandIcon={<ExpandMoreIcon/>}>
        <Typography>NPCs</Typography>
      </AccordionSummary>
      <AccordionDetails>
        <Stack spacing={2}>
          {npcs.map((npc, index) => (
            <Stack key={index} spacing={2}>
              <Stack direction="row" spacing={2} sx={{ alignItems: 'center' }}>
                <Typography>Type:</Typography>
                <Select
                  size="small"
                  value={npcData.length > npc.Type ? npc.Type : ''}
                  onChange={(event) => {
                    editNpc(index, (editedNpc) => {
                      editedNpc.Type = Number(event.target.value);
                    });
                  }}
                  sx={{ minWidth: '45%' }}
                >
                  {npcData.map((data, npcType) => (
                    <MenuItem key={npcType} value={npcType}>{data.Name}</MenuItem>
                  ))}
                </Select>
                <Box sx={{ marginLeft: 'auto' }}>
                  <IconButton
                    aria-label="Delete NPC"
                    onClick={() => {
                      const editedNpcs = structuredClone(npcs);
                      editedNpcs.splice(index, 1);
                      onEdit(editedNpcs);
                    }}
                  >
                    <DeleteIcon/>
                  </IconButton>
                </Box>
              </Stack>

              <Typography>Level:</Typography>
              <SubmitRangeNumberPicker
                min={1}
                value={npc.LevelRange}
                onSubmit={(value) => {
                  editNpc(index, (editedNpc) => {
                    editedNpc.LevelRange = value;
                  });
                }}
              />

              <Stack direction="row" spacing={2} sx={{ alignItems: 'center' }}>
                <Typography>Disposition:</Typography>
                <Select
                  size="small"
                  value={npc.StartingDisposition}
                  onChange={(event) => {
                    editNpc(index, (editedNpc) => {
                      editedNpc.StartingDisposition = Number(event.target.value);
                    });
                  }}
                  sx={{ minWidth: '45%' }}
                >
                  {Object.entries(NPC_DISPOSITION_NAMES).map(([disposition, name]) => (
                    <MenuItem key={disposition} value={Number(disposition)}>{name}</MenuItem>
                  ))}
                </Select>
              </Stack>

              <Stack direction="row" spacing={2} sx={{ alignItems: 'center' }}>
                <Typography>Movement:</Typography>
                <Select
                  size="small"
                  value={npc.MovementType}
                  onChange={(event) => {
                    editNpc(index, (editedNpc) => {
                      editedNpc.MovementType = Number(event.target.value);
                    });
                  }}
                  sx={{ minWidth: '45%' }}
                >
                  {Object.entries(NPC_MOVEMENT_TYPE_NAMES).map(([movementType, name]) => (
                    <MenuItem key={movementType} value={Number(movementType)}>{name}</MenuItem>
                  ))}
                </Select>
              </Stack>

              {npc.MovementType !== world.NpcMovementType.SENTINEL && <DurationField
                label="Movement Step"
                ticks={npc.MovementStepDuration}
                onSubmit={(ticks) => {
                  editNpc(index, (editedNpc) => {
                    editedNpc.MovementStepDuration = ticks;
                  });
                }}
              />}

              <DurationField
                label="Respawn"
                ticks={npc.RespawnDuration}
                onSubmit={(ticks) => {
                  editNpc(index, (editedNpc) => {
                    editedNpc.RespawnDuration = ticks;
                  });
                }}
              />

              <DurationField
                label="Awake"
                ticks={npc.AwakeDuration}
                onSubmit={(ticks) => {
                  editNpc(index, (editedNpc) => {
                    editedNpc.AwakeDuration = ticks;
                  });
                }}
              />

              <DurationField
                label="Sleep"
                ticks={npc.SleepDuration}
                onSubmit={(ticks) => {
                  editNpc(index, (editedNpc) => {
                    editedNpc.SleepDuration = ticks;
                  });
                }}
              />
              <Typography variant="caption" color="text.secondary">
                Set Awake and Sleep to 0 to disable sleeping.
              </Typography>

              <DropTableEditor
                name="Drop Table"
                dropTable={npc.DropTable}
                onEdit={(dropTable) => {
                  editNpc(index, (editedNpc) => {
                    editedNpc.DropTable = dropTable;
                  });
                }}
              />
              <Divider/>
            </Stack>
          ))}

          <Button onClick={() => {
            const editedNpcs = structuredClone(npcs);
            editedNpcs.push(world.Npc.createFrom({
              Type: 0,
              LevelRange: { Min: 1, Max: 1 },
              StartingDisposition: world.NpcDisposition.NEUTRAL,
              MovementType: world.NpcMovementType.SENTINEL,
              Behavior: { Hooks: null },
              // Filled in when the world is saved
              SpawnRoom: getEditorConstants().RoomNone,
              RespawnDuration: minutesToTicks(1),
              SleepDuration: 0,
              AwakeDuration: 0,
              MovementStepDuration: 0,
              DropTable: {
                Entries: [],
              },
            }));
            onEdit(editedNpcs);
          }}>+ Add NPC</Button>
        </Stack>
      </AccordionDetails>
    </Accordion>
  );
}
