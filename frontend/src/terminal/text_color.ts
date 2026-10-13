export const TextColor = {
  WHITE: 'w',
  RED: 'r',
  GREEN: 'g',
  BLUE: 'b',
  YELLOW: 'y',
};
export type TextColor = typeof TextColor[keyof typeof TextColor];

export const TEXT_COLOR_VALUES = {
  [TextColor.WHITE]: '#fff',
  [TextColor.RED]: '#ff0000',
  [TextColor.GREEN]: '#00ff00',
  [TextColor.BLUE]: '#0000ff',
  [TextColor.YELLOW]: '#ffff00',
};
