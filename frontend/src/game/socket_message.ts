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

export type SocketMessage = {
  message: string | null;
  status?: PlayerStatus;
  map?: number[];
  inv?: Inventory;
  equip?: EquipmentSlot[];
};
