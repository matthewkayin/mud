export type PlayerStatus = {
  hp: number;
  maxHp: number;
  mp: number;
  maxMp: number;
  action: boolean;
};

export type SocketMessage = {
  message: string | null;
  status?: PlayerStatus;
  map?: number[];
  inv?: string[];
};
