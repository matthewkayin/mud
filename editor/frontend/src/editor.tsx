import { Box } from '@mui/material';
import { Canvas } from './canvas';

export function Editor() {
  return (
    <Box sx={{
      display: 'flex',
      width: '100vw',
      height: '100vh',
      overflow: 'hidden'
    }}>
      <Canvas/>
    </Box>
  )
}
