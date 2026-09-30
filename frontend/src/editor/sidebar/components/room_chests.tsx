import * as mud from '../../../mud/types';
import {
  Accordion,
  AccordionSummary,
  AccordionDetails,
  Button,
  Typography,
  Stack,
  Divider,
} from '@mui/material';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import { ChestCluster } from './chest';

type RoomChestsClusterProps = {
  room: mud.Room;
  itemData: mud.ItemData[];
  onEdit: (editedRoom: mud.Room) => void;
}

export function RoomChestsCluster({ room, itemData, onEdit }: RoomChestsClusterProps) {
  return (
    <Accordion>
      <AccordionSummary expandIcon={<ExpandMoreIcon/>}>
        <Typography>Chests</Typography>
      </AccordionSummary>
      <AccordionDetails>
        <Stack spacing={2}>
          {room.Chests.map((chest, index) => (
          <Stack spacing={2}>
            <ChestCluster
              chest={chest}
              itemData={itemData}
              onEdit={(editedChest: mud.Chest) => {
                const editedRoom = structuredClone(room);
                editedRoom.Chests[index] = editedChest;
                onEdit(editedRoom);
              }}
            />
            <Divider/>
          </Stack>
          ))}
        </Stack>
        <Button onClick={() => {
          const editedRoom = structuredClone(room);
          editedRoom.Chests.push({
            Name: '',
            Type: mud.CHEST_TYPE_CHEST,
            Timer: 0,
            RespawnDuration: 60,
            DropTable: {
              Entries: [],
            },
            Inventory: {
              Items: [],
            }
          })
          onEdit(editedRoom);
        }}>+ Add Chest</Button>
      </AccordionDetails>
    </Accordion>
  )
}
