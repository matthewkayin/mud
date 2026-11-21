export const VIEW_MODES = ['help', 'chat', 'inventory', 'mobs'] as const;
export type ViewMode = typeof VIEW_MODES[number];

// Other names players can use for a mode
const VIEW_MODE_ALIASES: Record<string, ViewMode> = {
  equipment: 'inventory',
  inv: 'inventory',
};

// The mode a 'view' argument names, or null when it names none
export const parseViewMode = (value: string): ViewMode | null => {
  if ((VIEW_MODES as readonly string[]).includes(value)) {
    return value as ViewMode;
  }
  // hasOwn, so that names like 'constructor' don't match the object's prototype
  return Object.hasOwn(VIEW_MODE_ALIASES, value) ? VIEW_MODE_ALIASES[value] : null;
};
