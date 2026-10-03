import { main } from '../../wailsjs/go/models';
import { GetConstants } from '../../wailsjs/go/main/EditorState';

// Constants from the world package that aren't enums. These are loaded from Go
// before the editor renders, so they can be read synchronously everywhere else.
let editorConstants: main.EditorConstants | null = null;

export async function loadEditorConstants() {
  editorConstants = await GetConstants();
}

export function getEditorConstants(): main.EditorConstants {
  if (!editorConstants) {
    throw new Error("editor constants were read before they were loaded");
  }

  return editorConstants;
}
