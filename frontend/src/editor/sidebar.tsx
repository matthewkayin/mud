import { useEffect, useState } from 'react';
import { Checkbox, FormControlLabel, Paper, Stack, TextField, Typography } from '@mui/material';
import { EDITOR_ROOM_GRID_INDEX_NONE, EditorState, EditorActionType } from './state';
import * as mud from '../mud/types';

export type EditorSidebarProps = {
  generation: number;
  stateRef: React.RefObject<EditorState>;
}

type EditorSidebarRoomClusterProps = {
  generation: number;
  stateRef: React.RefObject<EditorState>;
}

export const EditorSidebar = ({ generation, stateRef }: EditorSidebarProps) => {
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
      <EditorSidebarRoomCluster generation={generation} stateRef={stateRef} />
    </Paper>
  );
};

const EditorSidebarRoomCluster = ({ generation, stateRef }: EditorSidebarRoomClusterProps) => {
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
              setRoom(() => newRoom);
              onSubmit(newRoom);
            }}
          />
        }
      />
    </Stack>
  )
}
