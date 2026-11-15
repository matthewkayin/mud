import type { Layout, LayoutRoom, Position } from "../shared/map_types";

export type ParseLayoutResult = {
    layout: Layout | null;
    error: string | null;
};

// Parses layout.json and checks that every field has the right type, so the map view can trust it.
// Rules such as unique positions and two-way exits are left to the validation server
export function parseLayout(text: string): ParseLayoutResult {
    let json: unknown;
    try {
        json = JSON.parse(text);
    } catch (error) {
        const message = error instanceof Error ? error.message : String(error);
        return { layout: null, error: `layout.json is not valid JSON: ${message}` };
    }

    try {
        const root = expectObject(json, "layout.json");
        const start = parsePosition(root["start"], "start");
        const roomsJson = root["rooms"];
        if (!Array.isArray(roomsJson)) {
            throw new Error("rooms must be an array");
        }
        const rooms = roomsJson.map((roomJson, index) => parseRoom(roomJson, `rooms[${index}]`));
        return { layout: { start, rooms }, error: null };
    } catch (error) {
        const message = error instanceof Error ? error.message : String(error);
        return { layout: null, error: `layout.json is invalid: ${message}` };
    }
}

function parseRoom(json: unknown, name: string): LayoutRoom {
    const room = expectObject(json, name);
    const position = parsePosition(room["position"], `${name}.position`);

    const path = room["path"];
    if (typeof path !== "string") {
        throw new Error(`${name}.path must be a string`);
    }

    const exits = room["exits"];
    if (!Array.isArray(exits) || !exits.every((exit) => typeof exit === "string")) {
        throw new Error(`${name}.exits must be an array of strings`);
    }

    return { position, path, exits };
}

function parsePosition(json: unknown, name: string): Position {
    const position = expectObject(json, name);
    const x = position["x"];
    const y = position["y"];
    const z = position["z"];
    if (!Number.isInteger(x) || !Number.isInteger(y) || !Number.isInteger(z)) {
        throw new Error(`${name} must have integer x, y and z`);
    }
    return { x: x as number, y: y as number, z: z as number };
}

function expectObject(json: unknown, name: string): Record<string, unknown> {
    if (typeof json !== "object" || json === null || Array.isArray(json)) {
        throw new Error(`${name} must be an object`);
    }
    return json as Record<string, unknown>;
}
