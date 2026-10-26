export type PlayerStatus = {
  health: number;
  maxHealth: number;
  mana: number;
  maxMana: number;
  action: boolean;
};

// A message from the server. message is null for a heartbeat.
// The status fields are only sent while the player is in-game.
// map is only sent when the player's minimap has changed.
export type SocketMessage = { message: string | null, map?: unknown } & Partial<PlayerStatus>;

export const parseSocketMessage = (data: string): SocketMessage | null => {
  let parsed: unknown;
  try {
    parsed = JSON.parse(data);
  } catch {
    return null;
  }

  if (typeof parsed !== 'object' || parsed === null || !('message' in parsed)) {
    return null;
  }
  if (typeof parsed.message !== 'string' && parsed.message !== null) {
    return null;
  }

  return parsed as SocketMessage;
};

// Returns the player's status if the message carries all of it
export const getPlayerStatus = (socketMessage: SocketMessage): PlayerStatus | null => {
  const { health, maxHealth, mana, maxMana, action } = socketMessage;
  if (typeof health !== 'number' || typeof maxHealth !== 'number' || typeof mana !== 'number' || typeof maxMana !== 'number' || typeof action !== 'boolean' ) {
    return null;
  }

  return { health, maxHealth, mana, maxMana, action };
};

// Returns the minimap if the message carries a valid one.
// A minimap is a square grid with an odd width, where each cell is -1 for no room or a bitfield of the room's exits.
export const getMinimap = (socketMessage: SocketMessage): number[] | null => {
  const { map } = socketMessage;
  if (map === undefined) {
    return null;
  }

  const width = Array.isArray(map) ? Math.sqrt(map.length) : 0;
  if (!Array.isArray(map) || !Number.isInteger(width) || width % 2 !== 1 || !map.every((cell) => Number.isInteger(cell) && cell >= -1)) {
    console.warn('Received an invalid minimap from the server.', map);
    return null;
  }

  return map;
};
