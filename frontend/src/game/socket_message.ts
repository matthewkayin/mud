export type PlayerStatus = {
  hp: number;
  maxHp: number;
  mp: number;
  maxMp: number;
  action: boolean;
};

export type InventoryItem = {
  name: string;
  amount: number;
  size: number;
};

export type Inventory = {
  size: number;
  capacity: number;
  items: InventoryItem[];
};

// name is empty when nothing is equipped in the slot
export type EquipmentSlot = {
  slot: string;
  name: string;
};

// One mob in the player's room. name and target include the mob's fuzzy number when several mobs in the room share its name
// target is empty when the mob isn't targeting a mob in the room
export type RoomMob = {
  name: string;
  condition: string;
  target: string;
};

// A named number on the character sheet, such as an attribute or stat
export type SheetValue = {
  name: string;
  value: number;
};

// exp is the experience earned toward the next level, out of expNext
export type CharacterSheet = {
  name: string;
  race: string;
  class: string;
  level: number;
  exp: number;
  expNext: number;
  attributes: SheetValue[];
  stats: SheetValue[];
  skills: string[];
};

export type SocketMessage = {
  message: string | null;
  status?: PlayerStatus;
  map?: number[];
  inv?: Inventory;
  equip?: EquipmentSlot[];
  mobs?: RoomMob[];
  sheet?: CharacterSheet;
};
