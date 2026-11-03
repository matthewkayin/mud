import { useEffect, useState } from 'react';
import { CssBaseline, ThemeProvider } from '@mui/material';
import { Editor } from './editor';
import { editorTheme } from './theme';
import { loadEditorConstants } from './store/constants';

type EditorPageStatus = 'loading' | 'disabled' | 'not-god' | 'ready' | 'error';

// The editor only exists when the server runs with --edit, and only gods can use it
export const EditorPage = () => {
  const [status, setStatus] = useState<EditorPageStatus>('loading');
  const [error, setError] = useState('');

  useEffect(() => {
    const load = async () => {
      const response = await fetch('/api/auth/config');
      const config = await response.json();
      if (config.editor !== true) {
        setStatus('disabled');
        return;
      }
      if (config.god !== true) {
        setStatus('not-god');
        return;
      }

      // The store reads world constants synchronously, so load them before rendering
      await loadEditorConstants();
      setStatus('ready');
    };

    load().catch((loadError) => {
      setError(String(loadError));
      setStatus('error');
    });
  }, []);

  switch (status) {
    case 'loading':
      return null;
    case 'disabled':
      return <p>The editor is not enabled. Run the server with <code>--edit</code> to use it.</p>;
    case 'not-god':
      return <p>Only gods can use the editor. <a href="/">Log in</a> as a god first.</p>;
    case 'error':
      return <p>Failed to load the editor: {error}</p>;
    case 'ready':
      return (
        <ThemeProvider theme={editorTheme} defaultMode="system" storageManager={null} noSsr>
          <CssBaseline enableColorScheme/>
          <Editor/>
        </ThemeProvider>
      );
  }
};
