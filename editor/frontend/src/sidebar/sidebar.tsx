import { Box, Stack, Typography } from '@mui/material';

export function Sidebar() {
  return (
    <Box sx={{
      width: '25%',
      minWidth: '25%',
      height: '100%',
      backgroundColor: '#f0f0f0',
      borderRight: '2px solid #d6d6d6',
      overflowY: 'auto'
    }}>
      <Stack spacing={2}>
        <Typography>Hello friend</Typography>
      </Stack>
    </Box>
  )
}
