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
import { editorStore } from '../../store';
import { BehaviorParamsEditor, defaultBehaviorParams } from './behavior_params';

type FixturesEditorProps = {
  fixtures: world.Fixture[];
  onEdit: (fixtures: world.Fixture[]) => void;
}

export function FixturesEditor({ fixtures, onEdit }: FixturesEditorProps) {
  const fixtureData = useSyncExternalStore(editorStore.subscribe, editorStore.getFixtureData);
  const itemData = useSyncExternalStore(editorStore.subscribe, editorStore.getItemData);

  const editFixture = (index: number, edit: (fixture: world.Fixture) => void) => {
    const editedFixtures = structuredClone(fixtures);
    edit(editedFixtures[index]);
    onEdit(editedFixtures);
  };

  return (
    <Accordion slotProps={{ transition: { unmountOnExit: true } }}>
      <AccordionSummary expandIcon={<ExpandMoreIcon/>}>
        <Typography>Fixtures</Typography>
      </AccordionSummary>
      <AccordionDetails>
        <Stack spacing={2}>
          {fixtures.map((fixture, index) => {
            const data = fixtureData.find((data) => data.Name === fixture.Id);

            return (
              <Stack key={index} spacing={2}>
                <Stack direction="row" spacing={2} sx={{ alignItems: 'center' }}>
                  <Typography>ID:</Typography>
                  <Select
                    size="small"
                    value={data ? fixture.Id : ''}
                    onChange={(event) => {
                      const newData = fixtureData.find((data) => data.Name === event.target.value);
                      if (!newData) {
                        return;
                      }
                      editFixture(index, (editedFixture) => {
                        editedFixture.Id = newData.Name;
                        editedFixture.BehaviorParams = defaultBehaviorParams(newData.BehaviorParams, itemData);
                      });
                    }}
                    sx={{ minWidth: '45%' }}
                  >
                    {fixtureData.map((data) => (
                      <MenuItem key={data.Name} value={data.Name}>{data.Name}</MenuItem>
                    ))}
                  </Select>
                  <Box sx={{ marginLeft: 'auto' }}>
                    <IconButton
                      aria-label="Delete fixture"
                      onClick={() => {
                        const editedFixtures = structuredClone(fixtures);
                        editedFixtures.splice(index, 1);
                        onEdit(editedFixtures);
                      }}
                    >
                      <DeleteIcon/>
                    </IconButton>
                  </Box>
                </Stack>

                {data && <BehaviorParamsEditor
                  paramTypes={data.BehaviorParams}
                  params={fixture.BehaviorParams}
                  onEdit={(params) => {
                    editFixture(index, (editedFixture) => {
                      editedFixture.BehaviorParams = params;
                    });
                  }}
                />}
                <Divider/>
              </Stack>
            );
          })}

          <Button disabled={fixtureData.length === 0} onClick={() => {
            const editedFixtures = structuredClone(fixtures);
            editedFixtures.push(world.Fixture.createFrom({
              Id: fixtureData[0].Name,
              BehaviorParams: defaultBehaviorParams(fixtureData[0].BehaviorParams, itemData),
            }));
            onEdit(editedFixtures);
          }}>+ Add Fixture</Button>
        </Stack>
      </AccordionDetails>
    </Accordion>
  );
}
