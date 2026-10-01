import { useEffect, useState } from 'react';
import { EditorState } from '../../state/state';
import * as mud from '../../../mud/types';
import { EDITOR_GRID_INDEX_NONE, EditorActionType } from '../../state';
import { DropTableCluster } from './drop_table';
import { TextField } from './text_field';
import {
  Stack,
  FormControlLabel,
  Checkbox,
  Typography,
} from '@mui/material';
import { RoomChestsCluster } from './room_chests';

export type RoomClusterProps = {
  generation: number;
  stateRef: React.RefObject<EditorState>;
  itemData: mud.ItemData[];
}

export function RoomCluster({ generation, stateRef, itemData }: RoomClusterProps) {
  const [roomIndex, setRoomIndex] = useState(mud.ROOM_NONE);
  const [room, setRoom] = useState<mud.Room | null>(null);

  useEffect(() => {
    const updateRoom = () => {
      if (stateRef.current.selectedRoomGridIndex === EDITOR_GRID_INDEX_NONE) {
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
      <Typography>(None Selected)</Typography>
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
        onSubmit={onSubmit}
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
        onSubmit={onSubmit}
        multiline
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

      <DropTableCluster
        name="Room Items"
        dropTable={room.DropTable}
        itemData={itemData}
        onEdit={(dropTable) => {
          const newRoom = structuredClone(room);
          newRoom.DropTable = dropTable;
          setRoom(newRoom);
          onSubmit(newRoom);
        }}
      />

      <RoomChestsCluster
        room={room}
        itemData={itemData}
        onNameEdit={(name: string, index: number) => {
          const editedRoom = structuredClone(room);
          editedRoom.Chests[index].Name = name;
          setRoom(editedRoom);
        }}
        onEdit={(editedRoom: mud.Room) => {
          setRoom(editedRoom);
          onSubmit(editedRoom);
        }}
      />
    </Stack>
  )
}
