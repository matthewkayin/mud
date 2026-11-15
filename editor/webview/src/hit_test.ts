import {
    DIRECTION_DOWN,
    DIRECTION_EAST,
    DIRECTION_SOUTH,
    DIRECTION_UP,
    findRoom,
    positionPlusDirection,
    positionsEqual,
    type Layout,
    type Position,
} from "../../shared/map_types";
import {
    arrowSlotRect,
    CELL_PITCH_X,
    CELL_PITCH_Y,
    cellRect,
    isVerticalArrowActive,
    ROOM_HEIGHT,
    ROOM_WIDTH,
    type Rect,
} from "./render";

// What the mouse is over. Only targets that do something when clicked are returned
export type HitTarget =
    | { kind: "room"; position: Position }
    | { kind: "emptyCell"; position: Position }
    // The gap east or south of a room, when there is a room on the other side too. Clicking it toggles the connection
    | { kind: "gap"; position: Position; direction: string }
    // A room's up or down arrow slot, when it has that exit or there is a room above or below. Clicking it toggles the
    // connection
    | { kind: "verticalArrow"; position: Position; direction: string };

// The cell a world point is in, counting the gaps east and south of a cell as part of it
export function cellAt(worldX: number, worldY: number, elevation: number): Position {
    return { x: Math.floor(worldX / CELL_PITCH_X), y: Math.floor(worldY / CELL_PITCH_Y), z: elevation };
}

export function hitTest(layout: Layout, elevation: number, worldX: number, worldY: number): HitTarget | null {
    const cell = cellAt(worldX, worldY, elevation);
    const localX = worldX - cell.x * CELL_PITCH_X;
    const localY = worldY - cell.y * CELL_PITCH_Y;
    const isInRoomColumn = localX < ROOM_WIDTH;
    const isInRoomRow = localY < ROOM_HEIGHT;

    if (isInRoomColumn && isInRoomRow) {
        const room = findRoom(layout, cell);
        if (room === undefined) {
            return { kind: "emptyCell", position: cell };
        }
        const rect = cellRect(cell.x, cell.y);
        for (const direction of [DIRECTION_UP, DIRECTION_DOWN]) {
            if (isInRect(arrowSlotRect(rect, direction), worldX, worldY) && isVerticalArrowActive(layout, room, direction)) {
                return { kind: "verticalArrow", position: cell, direction };
            }
        }
        return { kind: "room", position: cell };
    }

    // Corners between four cells do nothing
    if (isInRoomColumn === isInRoomRow) {
        return null;
    }
    const direction = isInRoomRow ? DIRECTION_EAST : DIRECTION_SOUTH;
    const neighborPosition = positionPlusDirection(cell, direction)!;
    if (findRoom(layout, cell) === undefined || findRoom(layout, neighborPosition) === undefined) {
        return null;
    }
    return { kind: "gap", position: cell, direction };
}

export function hitTargetsEqual(a: HitTarget | null, b: HitTarget | null): boolean {
    if (a === null || b === null) {
        return a === b;
    }
    if (a.kind !== b.kind || !positionsEqual(a.position, b.position)) {
        return false;
    }
    const directionA = "direction" in a ? a.direction : "";
    const directionB = "direction" in b ? b.direction : "";
    return directionA === directionB;
}

function isInRect(rect: Rect, x: number, y: number): boolean {
    return x >= rect.x && x < rect.x + rect.width && y >= rect.y && y < rect.y + rect.height;
}
