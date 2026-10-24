export type PlayerStatus = {
  health: number;
  maxHealth: number;
  mana: number;
  maxMana: number;
};

// A message from the server. message is null for a heartbeat.
// The status fields are only sent while the player is in-game.
export type SocketMessage = { message: string | null } & Partial<PlayerStatus>;

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
  const { health, maxHealth, mana, maxMana } = socketMessage;
  if (typeof health !== 'number' || typeof maxHealth !== 'number' || typeof mana !== 'number' || typeof maxMana !== 'number') {
    return null;
  }

  return { health, maxHealth, mana, maxMana };
};
