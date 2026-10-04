import { main, world } from './models';

// Talks to the editor's Go server. World files are opened and saved with the browser's
// File System Access API (Chromium only); the server converts between the world file
// format and the editor's format, since that logic lives in the mud/world package.

const WORLD_FILE_TYPES: FilePickerAcceptType[] = [
  { description: 'World JSON', accept: { 'application/json': ['.json'] } },
];

// The world file currently being edited, or null if it has never been saved
let fileHandle: FileSystemFileHandle | null = null;
// Set by the store whenever the editor has changes since the last save or open
let isDirty = false;

window.addEventListener('beforeunload', (event) => {
  if (isDirty) {
    event.preventDefault();
  }
});

updateTitle();

function updateTitle() {
  const fileName = fileHandle?.name ?? 'Untitled';
  const dirtyMarker = isDirty ? '*' : '';
  document.title = `Editor - ${fileName}${dirtyMarker}`;
}

function isAbortError(error: unknown): boolean {
  return error instanceof DOMException && error.name === 'AbortError';
}

// Errors from the server are plain text, so they can be shown to the user as-is
async function request(url: string, init?: RequestInit): Promise<Response> {
  const response = await fetch(url, init);
  if (!response.ok) {
    throw new Error(await response.text());
  }

  return response;
}

async function getJson<T>(url: string): Promise<T> {
  const response = await request(url);
  return response.json();
}

export function GetConstants(): Promise<main.EditorConstants> {
  return getJson('/api/constants');
}

export function GetItemData(): Promise<world.ItemData[]> {
  return getJson('/api/items');
}

export function GetNpcData(): Promise<world.NpcData[]> {
  return getJson('/api/npcs');
}

export function SetIsDirty(value: boolean) {
  isDirty = value;
  updateTitle();
}

// Asks the user whether to throw away unsaved changes. Returns true if there
// are no unsaved changes or the user chose to discard them.
export async function ConfirmDiscardChanges(): Promise<boolean> {
  if (!isDirty) {
    return true;
  }

  return window.confirm('You have unsaved changes. Discard them?');
}

// Lets the user pick a world file and loads it. Returns null if the user cancelled.
export async function OpenWorld(): Promise<main.EditorWorld | null> {
  let handle: FileSystemFileHandle;
  try {
    [handle] = await window.showOpenFilePicker({ types: WORLD_FILE_TYPES });
  } catch (error) {
    if (isAbortError(error)) {
      return null;
    }
    throw error;
  }

  const file = await handle.getFile();
  const response = await request('/api/world/decode', {
    method: 'POST',
    body: await file.text(),
  });
  const editorWorld: main.EditorWorld = await response.json();

  fileHandle = handle;
  isDirty = false;
  updateTitle();

  return editorWorld;
}

// Saves to the current world file, or asks for one if the world has never been saved.
// Returns false if the user cancelled.
export async function SaveWorld(editorWorld: main.EditorWorld): Promise<boolean> {
  if (!fileHandle) {
    return SaveWorldAs(editorWorld);
  }

  return saveWorldToHandle(fileHandle, editorWorld);
}

// Asks for a world file to save to. Returns false if the user cancelled.
export async function SaveWorldAs(editorWorld: main.EditorWorld): Promise<boolean> {
  let handle: FileSystemFileHandle;
  try {
    handle = await window.showSaveFilePicker({
      suggestedName: fileHandle?.name ?? 'world.json',
      types: WORLD_FILE_TYPES,
    });
  } catch (error) {
    if (isAbortError(error)) {
      return false;
    }
    throw error;
  }

  return saveWorldToHandle(handle, editorWorld);
}

async function saveWorldToHandle(handle: FileSystemFileHandle, editorWorld: main.EditorWorld): Promise<boolean> {
  // Encode first so that an invalid world never touches the file
  const response = await request('/api/world/encode', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(editorWorld),
  });
  const data = await response.text();

  // The writable writes to a temporary file and only replaces the original on close
  const writable = await handle.createWritable();
  await writable.write(data);
  await writable.close();

  fileHandle = handle;
  isDirty = false;
  updateTitle();

  return true;
}
