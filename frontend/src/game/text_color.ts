export const TextColor = {
  WHITE: 'w',
  RED: 'r',
  GREEN: 'g',
  BLUE: 'b',
  YELLOW: 'y',
  PURPLE: 'p',
};
export type TextColor = typeof TextColor[keyof typeof TextColor];

export const TEXT_COLOR_VALUES = {
  [TextColor.WHITE]: '#fff',
  [TextColor.RED]: '#ff0000',
  [TextColor.GREEN]: '#00ff00',
  [TextColor.BLUE]: '#5fcde4',
  [TextColor.YELLOW]: '#ffff00',
  [TextColor.PURPLE]: '#ff00ff',
};
