import { editorStore } from './store';
import { GetItemData, GetNpcData } from '../api/editor_api';

// Loads the server's item and NPC data. Called on startup and after each
// reload, since a reload can change the scripts the data comes from.
export async function loadScriptData() {
  const [itemData, npcData] = await Promise.all([
    GetItemData(),
    GetNpcData(),
  ]);

  editorStore.setItemData(itemData);
  editorStore.setNpcData(npcData);
}
