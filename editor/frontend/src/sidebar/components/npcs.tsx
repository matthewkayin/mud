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
import { editorStore, NPC_MOVEMENT_TYPE_NAMES } from '../../store';
import { getEditorConstants } from '../../store/constants';
import { SubmitRangeNumberPicker } from './submit_range_number_picker';
import { DropTableEditor } from './drop_table';
import { BehaviorParamsEditor, defaultBehaviorParams } from './behavior_params';

type NpcsEditorProps = {
  npcs: world.Npc[];
  onEdit: (npcs: world.Npc[]) => void;
}

// An empty MovementTypeOverride means the NPC uses its NPC data's movement type
const MOVEMENT_TYPE_DEFAULT = '';

export function NpcsEditor({ npcs, onEdit }: NpcsEditorProps) {
  const npcData = useSyncExternalStore(editorStore.subscribe, editorStore.getNpcData);
  const itemData = useSyncExternalStore(editorStore.subscribe, editorStore.getItemData);

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
          {npcs.map((npc, index) => {
            const data = npcData.find((data) => data.Id === npc.Id);

            return (
              <Stack key={index} spacing={2}>
                <Stack direction="row" spacing={2} sx={{ alignItems: 'center' }}>
                  <Typography>ID:</Typography>
                  <Select
                    size="small"
                    value={data ? npc.Id : ''}
                    onChange={(event) => {
                      const newData = npcData.find((data) => data.Id === event.target.value);
                      if (!newData) {
                        return;
                      }
                      editNpc(index, (editedNpc) => {
                        editedNpc.Id = newData.Id;
                        editedNpc.BehaviorParams = defaultBehaviorParams(newData, itemData);
                      });
                    }}
                    sx={{ minWidth: '45%' }}
                  >
                    {npcData.map((data) => (
                      <MenuItem key={data.Id} value={data.Id}>{data.Id}</MenuItem>
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
                  <Typography>Movement:</Typography>
                  <Select
                    size="small"
                    value={npc.MovementTypeOverride}
                    onChange={(event) => {
                      editNpc(index, (editedNpc) => {
                        editedNpc.MovementTypeOverride = event.target.value;
                      });
                    }}
                    sx={{ minWidth: '45%' }}
                  >
                    <MenuItem value={MOVEMENT_TYPE_DEFAULT}>
                      Default{data ? ` (${NPC_MOVEMENT_TYPE_NAMES[data.MovementType]})` : ''}
                    </MenuItem>
                    {Object.values(NPC_MOVEMENT_TYPE_NAMES).map((name) => (
                      <MenuItem key={name} value={name}>{name}</MenuItem>
                    ))}
                  </Select>
                </Stack>

                {data && <BehaviorParamsEditor
                  npcData={data}
                  params={npc.BehaviorParams}
                  onEdit={(params) => {
                    editNpc(index, (editedNpc) => {
                      editedNpc.BehaviorParams = params;
                    });
                  }}
                />}

                <DropTableEditor
                  name="Drop Table Override"
                  dropTable={npc.DropTableOverride}
                  onEdit={(dropTable) => {
                    editNpc(index, (editedNpc) => {
                      editedNpc.DropTableOverride = dropTable;
                    });
                  }}
                />
                <Typography variant="caption" color="text.secondary">
                  Leave the override empty to use the NPC's default drop table.
                </Typography>
                <Divider/>
              </Stack>
            );
          })}

          <Button disabled={npcData.length === 0} onClick={() => {
            const editedNpcs = structuredClone(npcs);
            editedNpcs.push(world.Npc.createFrom({
              Id: npcData[0].Id,
              LevelRange: { Min: 1, Max: 1 },
              MovementTypeOverride: MOVEMENT_TYPE_DEFAULT,
              DropTableOverride: {
                Entries: [],
              },
              BehaviorParams: defaultBehaviorParams(npcData[0], itemData),
              // Filled in when the world is saved
              SpawnRoom: getEditorConstants().RoomNone,
            }));
            onEdit(editedNpcs);
          }}>+ Add NPC</Button>
        </Stack>
      </AccordionDetails>
    </Accordion>
  );
}
