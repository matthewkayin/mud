import { Paper, TextField, Typography } from '@mui/material';
import * as mud from '../mud/types';

export type EditorSidebarProps = {
  room: mud.Room | null;
}

type EditorSidebarRoomClusterProps = {
  room: mud.Room;
}

export const EditorSidebar = ({ room }: EditorSidebarProps) => {
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
      { room === null && <Typography>No Room Selected</Typography> }
      { room !== null && <EditorSidebarRoomCluster room={room}/> }
    </Paper>
  );
};

const EditorSidebarRoomCluster = ({ room }: EditorSidebarRoomClusterProps) => {
  return (
    <TextField
      label="Name"
      value={room.Name}
      onChange={(event) => {
        room.Name = event.target.value;
      }}
      variant="standard"
    />
  )
}
