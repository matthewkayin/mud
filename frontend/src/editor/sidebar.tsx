import { useEffect, useState } from 'react';
import {
  Checkbox,
  FormControlLabel,
  Paper,
  Stack,
  TextField,
  Typography,
  Button,
  MenuItem,
  Select,
} from '@mui/material';
import NumberField from './number_field.tsx';
import { EDITOR_ROOM_GRID_INDEX_NONE, EditorState, EditorActionType } from './state';
import * as mud from '../mud/types';

export type EditorSidebarProps = {
  generation: number;
  stateRef: React.RefObject<EditorState>;
  itemData: mud.ItemData[];
}

type EditorSidebarRoomClusterProps = {
  generation: number;
  stateRef: React.RefObject<EditorState>;
  itemData: mud.ItemData[];
}

type EditorSidebarInventoryClusterProps = {
  inventory: mud.Inventory;
  itemData: mud.ItemData[];
  onAddItem: () => void;
  onEditItem: (itemIndex: number, item: mud.Item) => void;
}

export const EditorSidebar = ({ generation, stateRef, itemData }: EditorSidebarProps) => {
  return (
    <Paper
      elevation={2}
      sx={{
        width: '25%',
        minWidth: '25%',
        height: '100%',
        borderRadius: 0,
        borderRight: '1px solid',
        borderColor: 'divider',
        p: 2,
        boxSixing: 'border-box',
        overflowY: 'auto',
      }}
    >
      <EditorSidebarRoomCluster generation={generation} stateRef={stateRef} itemData={itemData} />
    </Paper>
  );
};

const EditorSidebarRoomCluster = ({ generation, stateRef, itemData }: EditorSidebarRoomClusterProps) => {
  const [roomIndex, setRoomIndex] = useState(mud.ROOM_NONE);
  const [room, setRoom] = useState<mud.Room | null>(null);

  useEffect(() => {
    const updateRoom = () => {
      if (stateRef.current.selectedRoomGridIndex === EDITOR_ROOM_GRID_INDEX_NONE) {
        setRoomIndex(mud.ROOM_NONE);
        return;
      }

      const roomIndex = stateRef.current.roomGrid[stateRef.current.selectedRoomGridIndex];
      if (roomIndex === mud.ROOM_NONE) {
        setRoomIndex(roomIndex);
        return;
      }

      setRoomIndex(roomIndex);
      setRoom(structuredClone(stateRef.current.world.Rooms[roomIndex]));
    };
    updateRoom();

  }, [generation, stateRef, setRoomIndex, setRoom]);

  if (roomIndex === mud.ROOM_NONE) {
    return (
      <Typography>No Room Selected</Typography>
    );
  }

  const onSubmit = (value = room) => {
    stateRef.current.doAction({
      type: EditorActionType.EDIT_ROOM,
      data: {
        roomIndex: roomIndex,
        value: structuredClone(value),
        previous: structuredClone(stateRef.current.world.Rooms[roomIndex]),
      },
    })
  };

  const onTextFieldKeydown = (event) => {
    if (event.key === 'Enter') {
      onSubmit();
    }
  };

  return (
    <Stack spacing={2}>
      <TextField
        label="Name"
        value={room.Name}
        onChange={(event) => {
          setRoom((previous) => ({
            ...previous,
            Name: event.target.value,
          }))
        }}
        onKeyDown={onTextFieldKeydown}
        onBlur={() => onSubmit()}
        variant="standard"
      />

      <TextField
        label="Description"
        value={room.Description}
        onChange={(event) => {
          setRoom((previous) => ({
            ...previous,
            Description: event.target.value,
          }))
        }}
        onKeyDown={onTextFieldKeydown}
        onBlur={() => onSubmit()}
        multiline
        variant="standard"
      />

      <FormControlLabel
        label="Safe Zone"
        control={
          <Checkbox
            checked={room.IsSafeZone}
            onChange={(event) => {
              const newRoom: mud.Room = {
                ...room,
                IsSafeZone: event.target.checked,
              };
              setRoom(newRoom);
              onSubmit(newRoom);
            }}
          />
        }
      />

      <Typography>Room Inventory:</Typography>
      <EditorSidebarInventoryCluster
        inventory={room.Inventory}
        itemData={itemData}
        onAddItem={() => {
          const newRoom = structuredClone(room);
          newRoom.Inventory.Items.push({
            Id: 0,
            Amount: 1,
            Durability: 0,
          });
          setRoom(newRoom);
          onSubmit(newRoom);
        }}
        onEditItem={(itemIndex, item) => {
          const newRoom = structuredClone(room);
          newRoom.Inventory.Items[itemIndex] = item;
          setRoom(newRoom);
          onSubmit(newRoom);
        }}
      />
    </Stack>
  )
}

const EditorSidebarInventoryCluster = ({ inventory, itemData, onAddItem, onEditItem }: EditorSidebarInventoryClusterProps) => {
  return (
    <Stack spacing={2}>
      {inventory.Items.map((item: mud.Item, itemIndex: number) => (
        <>
          <Stack
            direction="row"
            spacing={1}
            sx={{
              alignItems: 'center',
            }}
          >
            <Typography>ID:</Typography>
            <Select
              label="Item"
              value={item.Id}
              onChange={(event) => onEditItem(itemIndex, { ...item, Id: event.target.value })}
            >{itemData.map((entry, index) => (
              <MenuItem value={index}>{entry.Name}</MenuItem>
            ))}
            </Select>

            <Typography>ID:</Typography>
            <NumberField
              min={1}
              value={item.Amount}
              size="small"
              onValueChange={(value) => onEditItem(itemIndex, { ...item, Amount: value })}
            />
          </Stack>
        </>
      ))}
      <Button onClick={onAddItem}>+ Add Item</Button>
    </Stack>
  )
}
