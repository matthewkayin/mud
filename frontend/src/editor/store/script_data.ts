import { editorStore } from './store';
import { GetItemData, GetNpcData, GetFixtureData } from '../api/editor_api';

// Loads the server's item, NPC and fixture data. Called on startup and after each
// reload, since a reload can change the scripts the data comes from.
export async function loadScriptData() {
  const [itemData, npcData, fixtureData] = await Promise.all([
    GetItemData(),
    GetNpcData(),
    GetFixtureData(),
  ]);

  editorStore.setItemData(itemData);
  editorStore.setNpcData(npcData);
  editorStore.setFixtureData(fixtureData);
}
