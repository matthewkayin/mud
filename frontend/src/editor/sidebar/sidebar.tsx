import { Paper, Stack, Typography, IconButton } from '@mui/material';
import { RoomCluster } from './components/room';
import { EditorState } from '../state';
import * as mud from '../../mud/types';
import SaveIcon from '@mui/icons-material/Save';
import UndoIcon from '@mui/icons-material/Undo';
import RedoIcon from '@mui/icons-material/Redo';

export type SidebarProps = {
  generation: number;
  stateRef: React.RefObject<EditorState>;
  itemData: mud.ItemData[];
}

export function Sidebar({ generation, stateRef, itemData }: SidebarProps) {
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
      <Stack spacing={2}>
        <Stack direction="row" spacing={2}>
          <IconButton>
            <SaveIcon/>
          </IconButton>
          <IconButton onClick={() => stateRef.current.undoAction()}>
            <UndoIcon/>
          </IconButton>
          <IconButton onClick={() => stateRef.current.redoAction()}>
            <RedoIcon/>
          </IconButton>
        </Stack>

        <Typography>Room</Typography>
        <RoomCluster
          generation={generation}
          stateRef={stateRef}
          itemData={itemData}
        />
      </Stack>
    </Paper>
  );
};
