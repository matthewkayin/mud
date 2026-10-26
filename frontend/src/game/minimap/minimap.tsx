import './minimap.css';

// Sizes are in SVG user units. Each room sits in the middle of its cell, leaving a gap between rooms for the exits.
const CELL_SIZE = 10;
const ROOM_SIZE = 6;
const ROOM_OFFSET = (CELL_SIZE - ROOM_SIZE) / 2;
const CELL_CENTER = CELL_SIZE / 2;

const MINIMAP_NO_ROOM = -1;

// Indexed by exit bit (north, east, south, west). Each line runs from the middle of the room's side to the edge of its cell,
// so the halves drawn by two connected rooms meet in the gap, and an exit off the edge of the map trails to the border.
const DIRECTION_EXIT_LINES = [
  { x1: CELL_CENTER, y1: ROOM_OFFSET, x2: CELL_CENTER, y2: 0 },
  { x1: ROOM_OFFSET + ROOM_SIZE, y1: CELL_CENTER, x2: CELL_SIZE, y2: CELL_CENTER },
  { x1: CELL_CENTER, y1: ROOM_OFFSET + ROOM_SIZE, x2: CELL_CENTER, y2: CELL_SIZE },
  { x1: ROOM_OFFSET, y1: CELL_CENTER, x2: 0, y2: CELL_CENTER },
];

export type MinimapProps = {
  // Row-major grid of rooms with the player's room in the center. Each cell is -1 for no room, or a bitfield of the room's exits.
  map: number[] | null;
};

export const Minimap = ({ map }: MinimapProps) => {
  if (map === null) {
    return null;
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
            <rect className={index === centerIndex ? 'minimap-room current' : 'minimap-room'} x={ROOM_OFFSET} y={ROOM_OFFSET} width={ROOM_SIZE} height={ROOM_SIZE} />
          </g>
        );
      })}
    </svg>
  );
};
