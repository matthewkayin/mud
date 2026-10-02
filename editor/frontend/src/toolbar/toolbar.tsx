import { Stack, IconButton } from '@mui/material';
import SaveIcon from '@mui/icons-material/Save';

export function Toolbar() {
  return (
    <Stack
      direction="row"
      spacing={2}
      sx={{
        width: '100%',
        backgroundColor: '#f0f0f0',
        borderBottom: '2px solid #d6d6d6',
      }}
    >
      <IconButton>
        <SaveIcon/>
      </IconButton>
    </Stack>
  );
}
