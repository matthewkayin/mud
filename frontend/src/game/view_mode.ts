export const VIEW_MODES = ['help', 'chat', 'inventory', 'equipment'] as const;
export type ViewMode = typeof VIEW_MODES[number];

export const isViewMode = (value: string): value is ViewMode => {
  return (VIEW_MODES as readonly string[]).includes(value);
};
