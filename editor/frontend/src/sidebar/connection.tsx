import { Box, Stack, Typography, IconButton } from '@mui/material';
import DeleteIcon from '@mui/icons-material/Delete';

export function ConnectionEditor() {
  return (
    <Stack spacing={2}>
      <Stack direction="row" spacing={2} sx={{ alignItems: 'center' }}>
        <Typography>Connection</Typography>
        <Box sx={{ flexGrow: 1 }}/>
        <IconButton>
          <DeleteIcon />
        </IconButton>
      </Stack>
    </Stack>
  )
}
