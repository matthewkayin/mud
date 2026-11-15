// Map view edits to the layout. Each function returns a new layout, or null when the edit changes nothing, so no
// empty undo step is recorded. Rooms are named by position; if a hand-edited layout has two rooms at one position, the
// first is used.
//
// Imports use .ts extensions so `npm test` can run this file directly with Node
import {
    ALL_DIRECTIONS,
    oppositeDirection,
    positionPlusDirection,
    positionsEqual,
    type Layout,
    type LayoutRoom,
    type Position,
} from "../shared/map_types.ts";

export function createRoom(layout: Layout, position: Position): Layout | null {
    if (findRoomIndex(layout, position) !== -1) {
        return null;
    }
    const room: LayoutRoom = { position: { ...position }, path: "", exits: [] };
    return { start: layout.start, rooms: [...cloneRooms(layout), room] };
}

// Leaves start alone, even when it pointed at this room. The validation server reports a start with no room
export function deleteRoom(layout: Layout, position: Position): Layout | null {
    const index = findRoomIndex(layout, position);
    if (index === -1) {
        return null;
    }
    const rooms = cloneRooms(layout);
    removeExitsInto(rooms, position);
    rooms.splice(index, 1);
    return { start: layout.start, rooms };
}

// Moving a room removes all of its connections, since its neighbors are different at the new position. The start
// moves with the room
export function moveRoom(layout: Layout, from: Position, to: Position): Layout | null {
    if (from.z !== to.z || positionsEqual(from, to)) {
        return null;
    }
    const index = findRoomIndex(layout, from);
    if (index === -1 || findRoomIndex(layout, to) !== -1) {
        return null;
    }
    const rooms = cloneRooms(layout);
    removeExitsInto(rooms, from);
    rooms[index] = { position: { ...to }, path: rooms[index].path, exits: [] };
    const start = positionsEqual(layout.start, from) ? { ...to } : layout.start;
    return { start, rooms };
}

// Edits both sides of a connection together. If the room has the exit, it and the exit back are removed; otherwise both
// are added. This also repairs a one-way exit made by hand
export function toggleExit(layout: Layout, position: Position, direction: string): Layout | null {
    const neighborPosition = positionPlusDirection(position, direction);
    const backDirection = oppositeDirection(direction);
    if (neighborPosition === null || backDirection === null) {
        return null;
    }
    const index = findRoomIndex(layout, position);
    const neighborIndex = findRoomIndex(layout, neighborPosition);
    if (index === -1 || neighborIndex === -1) {
        return null;
    }

    const rooms = cloneRooms(layout);
    const room = rooms[index];
    const neighbor = rooms[neighborIndex];
    if (room.exits.includes(direction)) {
        room.exits = room.exits.filter((exit) => exit !== direction);
        neighbor.exits = neighbor.exits.filter((exit) => exit !== backDirection);
    } else {
        room.exits.push(direction);
        if (!neighbor.exits.includes(backDirection)) {
            neighbor.exits.push(backDirection);
        }
    }
    return { start: layout.start, rooms };
}

export function setStart(layout: Layout, position: Position): Layout | null {
    if (positionsEqual(layout.start, position) || findRoomIndex(layout, position) === -1) {
        return null;
    }
    return { start: { ...position }, rooms: cloneRooms(layout) };
}

// Assigns a room file, relative to the rooms folder. Exits and start are unchanged
export function setRoomPath(layout: Layout, position: Position, path: string): Layout | null {
    const index = findRoomIndex(layout, position);
    if (index === -1 || layout.rooms[index].path === path) {
        return null;
    }
    const rooms = cloneRooms(layout);
    rooms[index].path = path;
    return { start: layout.start, rooms };
}

// The fixed format the map view always writes, with one room per line so edits make small, readable diffs
export function formatLayout(layout: Layout): string {
    const lines = ["{", `    "start": ${formatPosition(layout.start)},`];
    if (layout.rooms.length === 0) {
        lines.push(`    "rooms": []`);
    } else {
        lines.push(`    "rooms": [`);
        layout.rooms.forEach((room, index) => {
            const separator = index === layout.rooms.length - 1 ? "" : ",";
            const exits = room.exits.map((exit) => JSON.stringify(exit)).join(", ");
            lines.push(
                `        { "position": ${formatPosition(room.position)}, "path": ${JSON.stringify(room.path)}, "exits": [${exits}] }${separator}`,
            );
        });
        lines.push(`    ]`);
    }
    lines.push("}");
    return lines.join("\n") + "\n";
}

function formatPosition(position: Position): string {
    return `{ "x": ${position.x}, "y": ${position.y}, "z": ${position.z} }`;
}

function findRoomIndex(layout: Layout, position: Position): number {
    return layout.rooms.findIndex((room) => positionsEqual(room.position, position));
}

function cloneRooms(layout: Layout): LayoutRoom[] {
    return layout.rooms.map((room) => ({ position: { ...room.position }, path: room.path, exits: [...room.exits] }));
}

// Removes every neighbor's exit that leads into position, including one-way exits
function removeExitsInto(rooms: LayoutRoom[], position: Position): void {
    for (const direction of ALL_DIRECTIONS) {
        const neighborPosition = positionPlusDirection(position, direction)!;
        const backDirection = oppositeDirection(direction)!;
        for (const room of rooms) {
            if (positionsEqual(room.position, neighborPosition)) {
                room.exits = room.exits.filter((exit) => exit !== backDirection);
            }
        }
    }
}
