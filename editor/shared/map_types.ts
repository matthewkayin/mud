// Types shared by the extension and the map view webview

export type Position = {
    x: number;
    y: number;
    z: number;
};

export type LayoutRoom = {
    position: Position;
    // Relative to the rooms folder. Empty for a room with no file assigned
    path: string;
    exits: string[];
};

export type Layout = {
    start: Position;
    rooms: LayoutRoom[];
};

// Sent from the extension to the webview. Always the full state
export type MapViewState = {
    // Null when layout.json doesn't parse
    layout: Layout | null;
    layoutError: string | null;
    // Room file path relative to the rooms folder, to its name. Only files that loaded are present
    roomNames: Record<string, string>;
};

// Sent from the webview to the extension. The edit fields are intents: the extension applies them to the current
// layout.json text, so the webview's copy of the layout may be stale. Rooms are named by position
export type MapViewRequest = {
    // Sent once when the webview has loaded and can receive state
    isReady?: boolean;
    createRoom?: Position;
    deleteRoom?: Position;
    moveRoom?: { from: Position; to: Position };
    // Adds the exit and its matching exit back if the room doesn't have it, otherwise removes both
    toggleExit?: { position: Position; direction: string };
    setStart?: Position;
};

export const DIRECTION_NORTH = "north";
export const DIRECTION_EAST = "east";
export const DIRECTION_SOUTH = "south";
export const DIRECTION_WEST = "west";
export const DIRECTION_UP = "up";
export const DIRECTION_DOWN = "down";

export const ALL_DIRECTIONS = [
    DIRECTION_NORTH,
    DIRECTION_EAST,
    DIRECTION_SOUTH,
    DIRECTION_WEST,
    DIRECTION_UP,
    DIRECTION_DOWN,
];

// The position one step in a direction. North is y - 1 and up is z + 1, matching RoomPosition.PlusDirection on the server.
// Returns null for an unknown direction
export function positionPlusDirection(position: Position, direction: string): Position | null {
    switch (direction) {
        case DIRECTION_NORTH:
            return { x: position.x, y: position.y - 1, z: position.z };
        case DIRECTION_EAST:
            return { x: position.x + 1, y: position.y, z: position.z };
        case DIRECTION_SOUTH:
            return { x: position.x, y: position.y + 1, z: position.z };
        case DIRECTION_WEST:
            return { x: position.x - 1, y: position.y, z: position.z };
        case DIRECTION_UP:
            return { x: position.x, y: position.y, z: position.z + 1 };
        case DIRECTION_DOWN:
            return { x: position.x, y: position.y, z: position.z - 1 };
        default:
            return null;
    }
}

// Returns null for an unknown direction
export function oppositeDirection(direction: string): string | null {
    switch (direction) {
        case DIRECTION_NORTH:
            return DIRECTION_SOUTH;
        case DIRECTION_EAST:
            return DIRECTION_WEST;
        case DIRECTION_SOUTH:
            return DIRECTION_NORTH;
        case DIRECTION_WEST:
            return DIRECTION_EAST;
        case DIRECTION_UP:
            return DIRECTION_DOWN;
        case DIRECTION_DOWN:
            return DIRECTION_UP;
        default:
            return null;
    }
}

export function positionsEqual(a: Position, b: Position): boolean {
    return a.x === b.x && a.y === b.y && a.z === b.z;
}

// The first room at a position. A hand-edited layout may have more than one
export function findRoom(layout: Layout, position: Position): LayoutRoom | undefined {
    return layout.rooms.find((room) => positionsEqual(room.position, position));
}
