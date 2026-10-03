import { useSyncExternalStore } from "react";
import { editorStore, EditorConnection, EditorActionEditConnection } from "../store";
import { Box, Stack, Typography, IconButton, FormControlLabel, Checkbox } from '@mui/material';
import DeleteIcon from '@mui/icons-material/Delete';

type ConnectionEditorProps = {
  selectedConnection: EditorConnection;
}

export function ConnectionEditor({ selectedConnection }: ConnectionEditorProps) {
  const isLocked = useSyncExternalStore(editorStore.subscribe, () => editorStore.getSelectedConnectionIsLocked());

  return (
    <Stack spacing={2}>
      <Stack direction="row" spacing={2} sx={{ alignItems: 'center' }}>
        <Typography>Connection</Typography>
        <Box sx={{ flexGrow: 1 }}/>
        <IconButton>
          <DeleteIcon />
        </IconButton>
      </Stack>

      <FormControlLabel
        label="Locked"
        control={
          <Checkbox
            checked={isLocked}
            onChange={() => {
              editorStore.doAction(new EditorActionEditConnection({
                connection: selectedConnection,
                isLocked: !isLocked,
              }))
            }}
          />
        }
      />
    </Stack>
  )
}
