import { createTheme } from '@mui/material';

export const editorTheme = createTheme({
  colorSchemes: {
    light: {
      palette: {
        background: {
          default: '#f0f0f0',
          paper: '#f0f0f0',
        },
        divider: '#d6d6d6',
      },
    },
    dark: {
      palette: {
        background: {
          default: '#1e1e1e',
          paper: '#1e1e1e',
        },
        divider: '#3a3a3a',
      },
    },
  },
});
