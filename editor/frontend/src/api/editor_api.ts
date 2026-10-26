import { main, world } from './models';

// Talks to the editor's Go server. The server reads and writes the world file in the
// backend's data folder and converts between its format and the editor's, since that
// logic lives in the mud/world package.

// Set by the store whenever the editor has changes since the last save or load
let isDirty = false;

window.addEventListener('beforeunload', (event) => {
  if (isDirty) {
    event.preventDefault();
  }
});

updateTitle();

function updateTitle() {
  const dirtyMarker = isDirty ? '*' : '';
  document.title = `Editor - world.json${dirtyMarker}`;
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

export function GetFixtureData(): Promise<world.FixtureData[]> {
  return getJson('/api/fixtures');
}

export function SetIsDirty(value: boolean) {
  isDirty = value;
  updateTitle();
}

export async function LoadWorld(): Promise<main.EditorWorld> {
  const editorWorld: main.EditorWorld = await getJson('/api/world');

  isDirty = false;
  updateTitle();

  return editorWorld;
}

export async function SaveWorld(editorWorld: main.EditorWorld) {
  await request('/api/world', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(editorWorld),
  });

  isDirty = false;
  updateTitle();
}
