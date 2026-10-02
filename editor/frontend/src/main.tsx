import React from 'react'
import {createRoot} from 'react-dom/client'
import { CssBaseline, ThemeProvider } from '@mui/material';
import { Editor }  from './editor'
import { editorTheme } from './theme';

const container = document.getElementById('root')

const root = createRoot(container!)

root.render(
    <React.StrictMode>
      <ThemeProvider theme={editorTheme} defaultMode="system" storageManager={null} noSsr>
        <CssBaseline enableColorScheme/>
        <Editor/>
      </ThemeProvider>
    </React.StrictMode>
)
