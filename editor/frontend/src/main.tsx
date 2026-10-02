import React from 'react'
import {createRoot} from 'react-dom/client'
import { CssBaseline } from '@mui/material';
import { Editor }  from './editor'

const container = document.getElementById('root')

const root = createRoot(container!)

root.render(
    <React.StrictMode>
      <CssBaseline/>
      <Editor/>
    </React.StrictMode>
)
