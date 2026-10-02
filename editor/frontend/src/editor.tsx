import { Box, Stack } from '@mui/material';
import { Canvas } from './canvas';
import { Toolbar } from './toolbar';
import { Sidebar } from './sidebar';

export function Editor() {
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
