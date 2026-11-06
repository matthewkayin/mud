// Sizes are in SVG user units, which are one pixel at the panel's size. Each room sits in the middle of its cell,
// leaving a gap between rooms for the exits.
const CELL_SIZE = 36;
const ROOM_SIZE = 16;
const ROOM_OFFSET = (CELL_SIZE - ROOM_SIZE) / 2;
const CELL_CENTER = CELL_SIZE / 2;

const MINIMAP_NO_ROOM = -1;

const EXIT_BIT_UP = 4;
const EXIT_BIT_DOWN = 5;

// Up and down exits are drawn as carets in the middle of the gap east of the room, outside the box. The gap is wide
// enough that a room's carets never touch the room to its east, even when that room has carets of its own.
const CARET_HALF_WIDTH = 3.5;
const CARET_CENTER_X = CELL_SIZE - CARET_HALF_WIDTH;
const CARET_HEIGHT = 4;
// How far a caret's tip sits from the room's top or bottom line
const CARET_INSET = 0;
const CARET_UP_TIP = ROOM_OFFSET + CARET_INSET;
const CARET_DOWN_TIP = ROOM_OFFSET + ROOM_SIZE - CARET_INSET;

const caretPoints = (centerX: number, tipY: number, baseY: number) =>
  `${centerX - CARET_HALF_WIDTH},${baseY} ${centerX},${tipY} ${centerX + CARET_HALF_WIDTH},${baseY} ${centerX - CARET_HALF_WIDTH},${baseY}`;

// The up caret sits near the top of the room's east side and the down caret near the bottom, either side of the east exit line
const VERTICAL_EXIT_CARETS = [
  { bit: EXIT_BIT_UP, points: caretPoints(CARET_CENTER_X, CARET_UP_TIP, CARET_UP_TIP + CARET_HEIGHT) },
  { bit: EXIT_BIT_DOWN, points: caretPoints(CARET_CENTER_X, CARET_DOWN_TIP, CARET_DOWN_TIP - CARET_HEIGHT) },
];

// Indexed by exit bit, for the horizontal directions only (north, east, south, west). Each line runs from the middle of the room's side to the edge of its cell,
// so the halves drawn by two connected rooms meet in the gap, and an exit off the edge of the map trails to the border.
const DIRECTION_EXIT_LINES = [
  { x1: CELL_CENTER, y1: ROOM_OFFSET, x2: CELL_CENTER, y2: 0 },
  { x1: ROOM_OFFSET + ROOM_SIZE, y1: CELL_CENTER, x2: CELL_SIZE, y2: CELL_CENTER },
  { x1: CELL_CENTER, y1: ROOM_OFFSET + ROOM_SIZE, x2: CELL_CENTER, y2: CELL_SIZE },
  { x1: ROOM_OFFSET, y1: CELL_CENTER, x2: 0, y2: CELL_CENTER },
];

export type MinimapProps = {
  // Row-major grid of rooms with the player's room in the center. Each cell is -1 for no room, or a bitfield of the room's exits
  // with bits ordered north, east, south, west, up, down.
  map: number[] | null;
};

export const Minimap = ({ map }: MinimapProps) => {
  // With no map (the player isn't in the world), show the empty panel
  if (map === null) {
    return <svg className="minimap" />;
  }

  const width = Math.sqrt(map.length);
  const centerIndex = Math.floor(map.length / 2);

  return (
    <svg className="minimap" viewBox={`0 0 ${width * CELL_SIZE} ${width * CELL_SIZE}`}>
      {map.map((exits, index) => {
        if (exits === MINIMAP_NO_ROOM) {
          return null;
        }

        const x = (index % width) * CELL_SIZE;
        const y = Math.floor(index / width) * CELL_SIZE;

        return (
          <g key={index} transform={`translate(${x} ${y})`}>
            {DIRECTION_EXIT_LINES.map((line, direction) => (exits & (1 << direction)) !== 0 && (
              <line key={direction} className="minimap-exit" {...line} />
            ))}
            {VERTICAL_EXIT_CARETS.map(({ bit, points }) => (exits & (1 << bit)) !== 0 && (
              <polyline key={bit} className="minimap-exit" points={points} />
            ))}
            <rect className={index === centerIndex ? 'minimap-room current' : 'minimap-room'} x={ROOM_OFFSET} y={ROOM_OFFSET} width={ROOM_SIZE} height={ROOM_SIZE} />
          </g>
        );
      })}
    </svg>
  );
};
