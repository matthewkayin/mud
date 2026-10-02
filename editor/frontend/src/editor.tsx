import { useEffect } from 'react';
import { Box } from '@mui/material';
import { Canvas } from './canvas';
import { Toolbar } from './toolbar';
import { Sidebar } from './sidebar';
import { editorStore } from './store';
import { GetItemData } from '../wailsjs/go/main/EditorState';

export function Editor() {
  // Load item data
  useEffect(() => {
    GetItemData().then(editorStore.setItemData);
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
