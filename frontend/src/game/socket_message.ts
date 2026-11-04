export type SocketMessage = {
  message: string | null;
};

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
