import { useCallback, useEffect, useState } from 'react';
import { editorStore } from '../store';
import {
  Box,
  Stack,
  IconButton,
  Divider,
  useColorScheme,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogContentText,
  DialogActions,
  Button,
} from '@mui/material';
import FolderOpenIcon from '@mui/icons-material/FolderOpen';
import SaveIcon from '@mui/icons-material/Save';
import UndoIcon from '@mui/icons-material/Undo';
import RedoIcon from '@mui/icons-material/Redo';
import DarkModeIcon from '@mui/icons-material/DarkMode';
import LightModeIcon from '@mui/icons-material/LightMode';
import { main } from '../../wailsjs/go/models';
import { ConfirmDiscardChanges, OpenWorld, SaveWorld, SaveWorldAs } from '../../wailsjs/go/main/EditorState';

type ErrorDialogState = {
  title: string;
  message: string;
}

export function Toolbar() {
  const { mode, systemMode, setMode } = useColorScheme();
  const resolvedMode = mode === 'system' ? systemMode : mode;
  const isDarkMode = resolvedMode === 'dark';

  const [errorDialog, setErrorDialog] = useState<ErrorDialogState | null>(null);

  const openWorld = useCallback(async () => {
    try {
      const shouldDiscard = await ConfirmDiscardChanges();
      if (!shouldDiscard) {
        return;
      }

      // OpenWorld resolves to null if the user cancelled the file dialog
      const editorWorld: main.EditorWorld | null = await OpenWorld();
      if (editorWorld) {
        editorStore.loadEditorWorld(editorWorld);
      }
    } catch (error) {
      setErrorDialog({ title: 'Could not open world', message: String(error) });
    }
  }, []);

  const saveWorld = useCallback(async (saveAs: boolean) => {
    try {
      const savePoint = editorStore.getTopAction();
      const editorWorld = editorStore.toEditorWorld();
      const didSave = saveAs ? await SaveWorldAs(editorWorld) : await SaveWorld(editorWorld);
      if (didSave) {
        editorStore.markSaved(savePoint);
      }
    } catch (error) {
      setErrorDialog({ title: 'Could not save world', message: String(error) });
    }
  }, []);

  // Shortcuts
  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.ctrlKey && event.code === 'KeyZ') {
        event.preventDefault();
        editorStore.undoAction();
      } else if (event.ctrlKey && event.code === 'KeyR') {
        event.preventDefault();
        editorStore.redoAction();
      } else if (event.ctrlKey && event.code === 'KeyO') {
        event.preventDefault();
        openWorld();
      } else if (event.ctrlKey && event.code === 'KeyS') {
        event.preventDefault();
        saveWorld(event.shiftKey);
      }
    };
    window.addEventListener('keydown', onKeyDown);

    return () => {
      window.removeEventListener('keydown', onKeyDown);
    };
  }, [openWorld, saveWorld]);

  return (
    <Stack
      direction="row"
      spacing={2}
      sx={{
        width: '100%',
        backgroundColor: 'background.default',
        borderBottom: '2px solid',
        borderColor: 'divider',
      }}
    >
      <IconButton title="Open (Ctrl+O)" aria-label="Open" onClick={openWorld}>
        <FolderOpenIcon/>
      </IconButton>
      <IconButton title="Save (Ctrl+S)" aria-label="Save" onClick={() => saveWorld(false)}>
        <SaveIcon/>
      </IconButton>

      <Divider orientation='vertical' flexItem />
      <IconButton onClick={() => editorStore.undoAction()}>
        <UndoIcon/>
      </IconButton>
      <IconButton onClick={() => editorStore.redoAction()}>
        <RedoIcon/>
      </IconButton>

      <Box sx={{ flexGrow: 1 }}/>
      <IconButton
        title={isDarkMode ? 'Switch to light theme' : 'Switch to dark theme'}
        aria-label={isDarkMode ? 'Switch to light theme' : 'Switch to dark theme'}
        onClick={() => setMode(isDarkMode ? 'light' : 'dark')}
      >
        {isDarkMode ? <LightModeIcon/> : <DarkModeIcon/>}
      </IconButton>

      <Dialog open={errorDialog !== null} onClose={() => setErrorDialog(null)}>
        <DialogTitle>{errorDialog?.title}</DialogTitle>
        <DialogContent>
          <DialogContentText sx={{ whiteSpace: 'pre-line' }}>
            {errorDialog?.message}
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setErrorDialog(null)}>OK</Button>
        </DialogActions>
      </Dialog>
    </Stack>
  );
}
