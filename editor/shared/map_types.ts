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

// Sent from the webview to the extension
export type MapViewRequest = {
    // Sent once when the webview has loaded and can receive state
    isReady?: boolean;
};

export const DIRECTION_NORTH = "north";
export const DIRECTION_EAST = "east";
export const DIRECTION_SOUTH = "south";
export const DIRECTION_WEST = "west";
export const DIRECTION_UP = "up";
export const DIRECTION_DOWN = "down";
