import {
    DIRECTION_DOWN,
    DIRECTION_EAST,
    DIRECTION_NORTH,
    DIRECTION_SOUTH,
    DIRECTION_UP,
    DIRECTION_WEST,
    findRoom,
    positionPlusDirection,
    positionsEqual,
    type Layout,
    type LayoutRoom,
    type Position,
} from "../../shared/map_types";
import type { HitTarget } from "./hit_test";

// World coordinates are in pixels at zoom 1. A room's cell is at (x * CELL_PITCH_X, y * CELL_PITCH_Y), with the gap
// between it and its neighbors to the east and south. North is y - 1, matching RoomPosition.PlusDirection on the server
export const ROOM_WIDTH = 160;
export const ROOM_HEIGHT = 64;
export const ROOM_GAP = 32;
export const CELL_PITCH_X = ROOM_WIDTH + ROOM_GAP;
export const CELL_PITCH_Y = ROOM_HEIGHT + ROOM_GAP;

const CONNECTION_LINE_WIDTH = 3;
const ROOM_LINE_WIDTH = 1.5;
const START_ROOM_LINE_WIDTH = 3;
const LABEL_PADDING = 10;
const ROOM_FONT_SIZE = 15;
const SMALL_FONT_SIZE = 11;
// The up and down arrows sit in square slots in a room's top-right corner, up rightmost
const ARROW_SLOT_SIZE = 14;
const ARROW_SLOT_MARGIN = 4;
const ARROW_SIZE = 8;
const SELECTION_OFFSET = 5;
const SELECTION_LINE_WIDTH = 2;
const DRAG_GHOST_ALPHA = 0.6;

export type Camera = {
    // The world point at the center of the view
    centerX: number;
    centerY: number;
    zoom: number;
};

// The canvas size in CSS pixels
export type Viewport = {
    width: number;
    height: number;
    pixelRatio: number;
};

export type Rect = {
    x: number;
    y: number;
    width: number;
    height: number;
};

export type Colors = {
    background: string;
    foreground: string;
    roomFill: string;
    accent: string;
    dim: string;
    fontFamily: string;
};

// A room being dragged to a new cell
export type RoomDrag = {
    from: Position;
    // Null until the mouse is over a cell
    to: Position | null;
};

export type MapRenderParams = {
    viewport: Viewport;
    camera: Camera;
    elevation: number;
    layout: Layout;
    roomNames: Record<string, string>;
    colors: Colors;
    hoverTarget: HitTarget | null;
    selectedPosition: Position | null;
    drag: RoomDrag | null;
    // The room a file from the explorer is being dragged over
    fileDropPosition: Position | null;
};

export function cellRect(x: number, y: number): Rect {
    return { x: x * CELL_PITCH_X, y: y * CELL_PITCH_Y, width: ROOM_WIDTH, height: ROOM_HEIGHT };
}

export function cellCenter(x: number, y: number): { x: number; y: number } {
    return { x: x * CELL_PITCH_X + ROOM_WIDTH / 2, y: y * CELL_PITCH_Y + ROOM_HEIGHT / 2 };
}

export function arrowSlotRect(roomRect: Rect, direction: string): Rect {
    const slotIndex = direction === DIRECTION_UP ? 0 : 1;
    return {
        x: roomRect.x + roomRect.width - ARROW_SLOT_MARGIN - ARROW_SLOT_SIZE * (slotIndex + 1),
        y: roomRect.y + ARROW_SLOT_MARGIN,
        width: ARROW_SLOT_SIZE,
        height: ARROW_SLOT_SIZE,
    };
}

// An up or down arrow can be clicked to remove an existing exit, or to connect to a room directly above or below
export function isVerticalArrowActive(layout: Layout, room: LayoutRoom, direction: string): boolean {
    if (room.exits.includes(direction)) {
        return true;
    }
    return findRoom(layout, positionPlusDirection(room.position, direction)!) !== undefined;
}

// A drag can drop on an empty cell other than the one it started from
export function isValidDragTarget(layout: Layout, drag: RoomDrag): boolean {
    return drag.to !== null && !positionsEqual(drag.from, drag.to) && findRoom(layout, drag.to) === undefined;
}

export function screenToWorld(viewport: Viewport, camera: Camera, screenX: number, screenY: number): { x: number; y: number } {
    return {
        x: (screenX - viewport.width / 2) / camera.zoom + camera.centerX,
        y: (screenY - viewport.height / 2) / camera.zoom + camera.centerY,
    };
}

export function worldToScreen(viewport: Viewport, camera: Camera, worldX: number, worldY: number): { x: number; y: number } {
    return {
        x: (worldX - camera.centerX) * camera.zoom + viewport.width / 2,
        y: (worldY - camera.centerY) * camera.zoom + viewport.height / 2,
    };
}

export function renderMap(context: CanvasRenderingContext2D, params: MapRenderParams): void {
    const { viewport, camera, colors } = params;

    context.setTransform(viewport.pixelRatio, 0, 0, viewport.pixelRatio, 0, 0);
    context.fillStyle = colors.background;
    context.fillRect(0, 0, viewport.width, viewport.height);

    // From here on everything is drawn in world coordinates
    context.translate(viewport.width / 2, viewport.height / 2);
    context.scale(camera.zoom, camera.zoom);
    context.translate(-camera.centerX, -camera.centerY);

    const rooms = params.layout.rooms.filter((room) => room.position.z === params.elevation);
    const occupiedCells = new Set(rooms.map((room) => `${room.position.x},${room.position.y}`));

    renderEmptyCells(context, params, occupiedCells);
    for (const room of rooms) {
        renderConnections(context, room, colors);
    }
    renderHoveredGap(context, params);
    for (const room of rooms) {
        renderRoom(context, room, params);
    }
    renderHoveredEmptyCell(context, params);
    renderSelection(context, params);
    renderDrag(context, params);
    renderFileDrop(context, params);
}

// Faint outlines of the empty cells in view, so the grid is visible
function renderEmptyCells(context: CanvasRenderingContext2D, params: MapRenderParams, occupiedCells: Set<string>): void {
    const topLeft = screenToWorld(params.viewport, params.camera, 0, 0);
    const bottomRight = screenToWorld(params.viewport, params.camera, params.viewport.width, params.viewport.height);
    const minX = Math.floor(topLeft.x / CELL_PITCH_X);
    const minY = Math.floor(topLeft.y / CELL_PITCH_Y);
    const maxX = Math.floor(bottomRight.x / CELL_PITCH_X);
    const maxY = Math.floor(bottomRight.y / CELL_PITCH_Y);

    context.save();
    context.strokeStyle = params.colors.dim;
    context.globalAlpha = 0.25;
    context.lineWidth = 1 / params.camera.zoom;
    context.setLineDash([4 / params.camera.zoom, 4 / params.camera.zoom]);
    for (let y = minY; y <= maxY; y++) {
        for (let x = minX; x <= maxX; x++) {
            if (occupiedCells.has(`${x},${y}`)) {
                continue;
            }
            const rect = cellRect(x, y);
            context.strokeRect(rect.x, rect.y, rect.width, rect.height);
        }
    }
    context.restore();
}

// An empty cell under the mouse is outlined, since clicking it creates a room
function renderHoveredEmptyCell(context: CanvasRenderingContext2D, params: MapRenderParams): void {
    const target = params.hoverTarget;
    if (params.drag !== null || target === null || target.kind !== "emptyCell") {
        return;
    }
    const rect = cellRect(target.position.x, target.position.y);
    context.save();
    context.strokeStyle = params.colors.accent;
    context.lineWidth = ROOM_LINE_WIDTH;
    context.setLineDash([6, 4]);
    context.strokeRect(rect.x, rect.y, rect.width, rect.height);
    context.restore();
}

// Clicking a gap toggles the connection across it. An unconnected gap previews the connection as a dotted line, and a
// connected one is highlighted, since clicking removes it
function renderHoveredGap(context: CanvasRenderingContext2D, params: MapRenderParams): void {
    const target = params.hoverTarget;
    if (params.drag !== null || target === null || target.kind !== "gap") {
        return;
    }
    const room = findRoom(params.layout, target.position);
    if (room === undefined) {
        return;
    }
    const rect = cellRect(target.position.x, target.position.y);
    context.save();
    context.lineWidth = CONNECTION_LINE_WIDTH;
    if (room.exits.includes(target.direction)) {
        context.strokeStyle = params.colors.accent;
    } else {
        context.strokeStyle = params.colors.foreground;
        context.setLineDash([3, 4]);
    }
    context.beginPath();
    if (target.direction === DIRECTION_EAST) {
        const centerY = rect.y + rect.height / 2;
        context.moveTo(rect.x + rect.width, centerY);
        context.lineTo(rect.x + rect.width + ROOM_GAP, centerY);
    } else {
        const centerX = rect.x + rect.width / 2;
        context.moveTo(centerX, rect.y + rect.height);
        context.lineTo(centerX, rect.y + rect.height + ROOM_GAP);
    }
    context.stroke();
    context.restore();
}

function renderSelection(context: CanvasRenderingContext2D, params: MapRenderParams): void {
    const selected = params.selectedPosition;
    if (selected === null || selected.z !== params.elevation || findRoom(params.layout, selected) === undefined) {
        return;
    }
    const rect = cellRect(selected.x, selected.y);
    context.save();
    context.strokeStyle = params.colors.accent;
    context.lineWidth = SELECTION_LINE_WIDTH;
    context.setLineDash([6, 4]);
    context.strokeRect(
        rect.x - SELECTION_OFFSET,
        rect.y - SELECTION_OFFSET,
        rect.width + SELECTION_OFFSET * 2,
        rect.height + SELECTION_OFFSET * 2,
    );
    context.restore();
}

// A room a file is being dragged over is outlined, since dropping assigns the file to it
function renderFileDrop(context: CanvasRenderingContext2D, params: MapRenderParams): void {
    const position = params.fileDropPosition;
    if (position === null || position.z !== params.elevation || findRoom(params.layout, position) === undefined) {
        return;
    }
    const rect = cellRect(position.x, position.y);
    context.save();
    context.strokeStyle = params.colors.accent;
    context.lineWidth = SELECTION_LINE_WIDTH;
    context.strokeRect(
        rect.x - SELECTION_OFFSET,
        rect.y - SELECTION_OFFSET,
        rect.width + SELECTION_OFFSET * 2,
        rect.height + SELECTION_OFFSET * 2,
    );
    context.restore();
}

// The dragged room is drawn faded at the cell under the mouse, outlined in the accent color if it can be dropped there
function renderDrag(context: CanvasRenderingContext2D, params: MapRenderParams): void {
    const drag = params.drag;
    if (drag === null || drag.to === null) {
        return;
    }
    const room = findRoom(params.layout, drag.from);
    if (room === undefined) {
        return;
    }
    const ghost: LayoutRoom = { position: drag.to, path: room.path, exits: [] };
    context.save();
    context.globalAlpha = DRAG_GHOST_ALPHA;
    renderRoom(context, ghost, { ...params, hoverTarget: null });
    context.restore();

    if (isValidDragTarget(params.layout, drag)) {
        const rect = cellRect(drag.to.x, drag.to.y);
        context.save();
        context.strokeStyle = params.colors.accent;
        context.lineWidth = START_ROOM_LINE_WIDTH;
        context.strokeRect(rect.x, rect.y, rect.width, rect.height);
        context.restore();
    }
}

// Each exit draws from its room's edge to the middle of the gap, so a two-way connection is a full line and a one-way
// exit is a stub
function renderConnections(context: CanvasRenderingContext2D, room: LayoutRoom, colors: Colors): void {
    const rect = cellRect(room.position.x, room.position.y);
    const centerX = rect.x + rect.width / 2;
    const centerY = rect.y + rect.height / 2;
    const halfGap = ROOM_GAP / 2;

    context.save();
    context.strokeStyle = colors.foreground;
    context.lineWidth = CONNECTION_LINE_WIDTH;
    context.lineCap = "butt";
    context.beginPath();
    for (const exit of room.exits) {
        switch (exit) {
            case DIRECTION_NORTH:
                context.moveTo(centerX, rect.y);
                context.lineTo(centerX, rect.y - halfGap);
                break;
            case DIRECTION_SOUTH:
                context.moveTo(centerX, rect.y + rect.height);
                context.lineTo(centerX, rect.y + rect.height + halfGap);
                break;
            case DIRECTION_WEST:
                context.moveTo(rect.x, centerY);
                context.lineTo(rect.x - halfGap, centerY);
                break;
            case DIRECTION_EAST:
                context.moveTo(rect.x + rect.width, centerY);
                context.lineTo(rect.x + rect.width + halfGap, centerY);
                break;
        }
    }
    context.stroke();
    context.restore();
}

function renderRoom(context: CanvasRenderingContext2D, room: LayoutRoom, params: MapRenderParams): void {
    const { colors, layout } = params;
    const rect = cellRect(room.position.x, room.position.y);
    const isStartRoom =
        room.position.x === layout.start.x && room.position.y === layout.start.y && room.position.z === layout.start.z;

    context.save();
    context.fillStyle = colors.roomFill;
    context.fillRect(rect.x, rect.y, rect.width, rect.height);
    context.strokeStyle = isStartRoom ? colors.accent : colors.foreground;
    context.lineWidth = isStartRoom ? START_ROOM_LINE_WIDTH : ROOM_LINE_WIDTH;
    context.strokeRect(rect.x, rect.y, rect.width, rect.height);

    // Label. A room whose file didn't load shows its path, and an unassigned room shows nothing
    const label = params.roomNames[room.path] ?? room.path;
    if (label !== "") {
        context.fillStyle = colors.foreground;
        context.font = `${ROOM_FONT_SIZE}px ${colors.fontFamily}`;
        context.textAlign = "center";
        context.textBaseline = "middle";
        const fittedLabel = fitText(context, label, rect.width - LABEL_PADDING * 2);
        context.fillText(fittedLabel, rect.x + rect.width / 2, rect.y + rect.height / 2);
    }

    context.font = `${SMALL_FONT_SIZE}px ${colors.fontFamily}`;
    context.textBaseline = "top";

    if (isStartRoom) {
        context.fillStyle = colors.accent;
        context.textAlign = "left";
        context.fillText("START", rect.x + 6, rect.y + 5);
    }

    for (const direction of [DIRECTION_UP, DIRECTION_DOWN]) {
        renderVerticalArrow(context, room, rect, direction, params);
    }

    context.restore();
}

// A solid arrow for an exit. While hovered, an existing exit is highlighted (clicking removes it) and a room above or
// below with no exit gets a dotted arrow (clicking connects it)
function renderVerticalArrow(
    context: CanvasRenderingContext2D,
    room: LayoutRoom,
    roomRect: Rect,
    direction: string,
    params: MapRenderParams,
): void {
    const target = params.hoverTarget;
    const isHovered =
        params.drag === null &&
        target !== null &&
        target.kind === "verticalArrow" &&
        target.direction === direction &&
        positionsEqual(target.position, room.position);
    const hasExit = room.exits.includes(direction);
    if (!hasExit && !(isHovered && isVerticalArrowActive(params.layout, room, direction))) {
        return;
    }

    const slot = arrowSlotRect(roomRect, direction);
    const centerX = slot.x + slot.width / 2;
    const centerY = slot.y + slot.height / 2;
    const halfSize = ARROW_SIZE / 2;
    const tipY = direction === DIRECTION_UP ? centerY - halfSize : centerY + halfSize;
    const baseY = direction === DIRECTION_UP ? centerY + halfSize : centerY - halfSize;

    context.save();
    context.beginPath();
    context.moveTo(centerX, tipY);
    context.lineTo(centerX + halfSize, baseY);
    context.lineTo(centerX - halfSize, baseY);
    context.closePath();
    if (hasExit) {
        context.fillStyle = isHovered ? params.colors.accent : params.colors.foreground;
        context.fill();
    } else {
        context.strokeStyle = params.colors.foreground;
        context.lineWidth = 1;
        context.setLineDash([2, 2]);
        context.stroke();
    }
    context.restore();
}

// Shortens text with an ellipsis until it fits in maxWidth, using the context's current font
function fitText(context: CanvasRenderingContext2D, text: string, maxWidth: number): string {
    if (context.measureText(text).width <= maxWidth) {
        return text;
    }
    let length = text.length;
    while (length > 0 && context.measureText(text.slice(0, length) + "…").width > maxWidth) {
        length--;
    }
    return text.slice(0, length) + "…";
}
