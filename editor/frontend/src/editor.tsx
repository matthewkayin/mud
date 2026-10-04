import { useEffect } from 'react';
import { Box } from '@mui/material';
import { Canvas } from './canvas';
import { Toolbar } from './toolbar';
import { Sidebar } from './sidebar';
import { editorStore } from './store';
import { GetItemData, GetNpcData } from './api/editor_api';

export function Editor() {
  // Load item and NPC data
  useEffect(() => {
    GetItemData().then(editorStore.setItemData);
    GetNpcData().then(editorStore.setNpcData);
  }, []);

  return (
    <Box sx={{
      display: 'flex',
      width: '100vw',
      height: '100vh',
      overflow: 'hidden'
    }}>
      <Box sx={{ width: '100%' }}>
        <Toolbar/>
        <Box sx={{ display: 'flex', width: '100%', height: '100%' }}>
          <Sidebar/>
          <Canvas/>
        </Box>
      </Box>
    </Box>
  )
}
