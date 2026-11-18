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

export type SocketMessage = {
  message: string | null;
  status?: PlayerStatus;
  map?: number[];
  inv?: Inventory;
  equip?: string[];
};
