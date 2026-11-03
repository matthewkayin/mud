import { Box, Stack, Typography, IconButton } from '@mui/material';
import DeleteIcon from '@mui/icons-material/Delete';
import { editorStore, EditorConnection } from '../store';
import { EditorActionDisconnectRooms } from '../store/action';

type ConnectionEditorProps = {
  connection: EditorConnection;
}

export function ConnectionEditor({ connection }: ConnectionEditorProps) {
  return (
    <Stack spacing={2}>
      <Stack direction="row" spacing={2} sx={{ alignItems: 'center' }}>
        <Typography>Connection</Typography>
        <Box sx={{ flexGrow: 1 }}/>
        <IconButton onClick={() => {
          editorStore.doAction(new EditorActionDisconnectRooms({ connection }));
        }}>
          <DeleteIcon />
        </IconButton>
      </Stack>
    </Stack>
  )
}
