import { Paper } from '@mui/material';
import { RoomCluster } from './components/room';
import { EditorState } from '../state';
import * as mud from '../../mud/types';

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
      <RoomCluster
        generation={generation}
        stateRef={stateRef}
        itemData={itemData}
      />
    </Paper>
  );
};
