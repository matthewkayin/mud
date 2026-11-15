import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { test } from "node:test";
import type { Layout, LayoutRoom, Position } from "../shared/map_types.ts";
import { parseLayout } from "../src/layout.ts";
import { createRoom, deleteRoom, formatLayout, moveRoom, setStart, toggleExit } from "../src/layout_edit.ts";

function position(x: number, y: number, z: number): Position {
    return { x, y, z };
}

function room(x: number, y: number, z: number, exits: string[], path = ""): LayoutRoom {
    return { position: position(x, y, z), path, exits };
}

function exitsAt(layout: Layout, x: number, y: number, z: number): string[] {
    const found = layout.rooms.find((room) => room.position.x === x && room.position.y === y && room.position.z === z);
    assert.ok(found, `no room at ${x},${y},${z}`);
    return found.exits;
}

// Two connected rooms side by side, and one above the first that isn't connected
function sampleLayout(): Layout {
    return {
        start: position(0, 0, 0),
        rooms: [room(0, 0, 0, ["east"], "a.lua"), room(1, 0, 0, ["west"], "b.lua"), room(0, 0, 1, [], "c.lua")],
    };
}

test("createRoom adds an unassigned room to an empty cell", () => {
    const result = createRoom(sampleLayout(), position(0, 1, 0));
    assert.ok(result);
    assert.deepEqual(result.rooms.at(-1), room(0, 1, 0, []));
    assert.equal(result.rooms.length, 4);
});

test("createRoom does nothing on an occupied cell", () => {
    assert.equal(createRoom(sampleLayout(), position(1, 0, 0)), null);
});

test("edits don't change the layout they're given", () => {
    const layout = sampleLayout();
    const before = structuredClone(layout);
    toggleExit(layout, position(0, 0, 0), "up");
    deleteRoom(layout, position(0, 0, 0));
    moveRoom(layout, position(0, 0, 0), position(0, 1, 0));
    assert.deepEqual(layout, before);
});

test("toggleExit adds both sides of a connection", () => {
    const result = toggleExit(sampleLayout(), position(0, 0, 0), "up");
    assert.ok(result);
    assert.deepEqual(exitsAt(result, 0, 0, 0), ["east", "up"]);
    assert.deepEqual(exitsAt(result, 0, 0, 1), ["down"]);
});

test("toggleExit removes both sides of a connection", () => {
    const result = toggleExit(sampleLayout(), position(1, 0, 0), "west");
    assert.ok(result);
    assert.deepEqual(exitsAt(result, 0, 0, 0), []);
    assert.deepEqual(exitsAt(result, 1, 0, 0), []);
});

test("toggleExit repairs a one-way exit", () => {
    const layout: Layout = { start: position(0, 0, 0), rooms: [room(0, 0, 0, []), room(1, 0, 0, ["west"])] };
    const result = toggleExit(layout, position(0, 0, 0), "east");
    assert.ok(result);
    assert.deepEqual(exitsAt(result, 0, 0, 0), ["east"]);
    assert.deepEqual(exitsAt(result, 1, 0, 0), ["west"]);
});

test("toggleExit does nothing without a room on both ends", () => {
    assert.equal(toggleExit(sampleLayout(), position(0, 0, 0), "north"), null);
    assert.equal(toggleExit(sampleLayout(), position(5, 5, 5), "north"), null);
    assert.equal(toggleExit(sampleLayout(), position(0, 0, 0), "sideways"), null);
});

test("deleteRoom removes the room and every exit into it, leaving start alone", () => {
    const layout = sampleLayout();
    // A one-way exit down into the room being deleted
    layout.rooms[2].exits = ["down"];
    const result = deleteRoom(layout, position(0, 0, 0));
    assert.ok(result);
    assert.equal(result.rooms.length, 2);
    assert.deepEqual(exitsAt(result, 1, 0, 0), []);
    assert.deepEqual(exitsAt(result, 0, 0, 1), []);
    assert.deepEqual(result.start, position(0, 0, 0));
});

test("deleteRoom does nothing on an empty cell", () => {
    assert.equal(deleteRoom(sampleLayout(), position(3, 3, 0)), null);
});

test("moveRoom clears connections on both sides and moves start with the room", () => {
    const result = moveRoom(sampleLayout(), position(0, 0, 0), position(0, 1, 0));
    assert.ok(result);
    assert.deepEqual(result.rooms[0], room(0, 1, 0, [], "a.lua"));
    assert.deepEqual(exitsAt(result, 1, 0, 0), []);
    assert.deepEqual(result.start, position(0, 1, 0));
});

test("moveRoom keeps start when moving another room", () => {
    const result = moveRoom(sampleLayout(), position(1, 0, 0), position(2, 0, 0));
    assert.ok(result);
    assert.deepEqual(result.start, position(0, 0, 0));
});

test("moveRoom refuses an occupied cell, another level or the same cell", () => {
    assert.equal(moveRoom(sampleLayout(), position(0, 0, 0), position(1, 0, 0)), null);
    assert.equal(moveRoom(sampleLayout(), position(0, 0, 0), position(0, 1, 1)), null);
    assert.equal(moveRoom(sampleLayout(), position(0, 0, 0), position(0, 0, 0)), null);
    assert.equal(moveRoom(sampleLayout(), position(9, 9, 0), position(8, 8, 0)), null);
});

test("setStart moves start to a room", () => {
    const result = setStart(sampleLayout(), position(1, 0, 0));
    assert.ok(result);
    assert.deepEqual(result.start, position(1, 0, 0));
});

test("setStart does nothing on an empty cell or the current start", () => {
    assert.equal(setStart(sampleLayout(), position(4, 4, 0)), null);
    assert.equal(setStart(sampleLayout(), position(0, 0, 0)), null);
});

test("formatLayout writes the game's layout.json unchanged", () => {
    const text = readFileSync(new URL("../../backend/data/layout.json", import.meta.url), "utf8");
    const parsed = parseLayout(text);
    assert.ok(parsed.layout, parsed.error ?? "");
    assert.equal(formatLayout(parsed.layout), text);
});

test("formatLayout writes an empty room list and escapes paths", () => {
    assert.equal(
        formatLayout({ start: position(0, 0, 0), rooms: [] }),
        '{\n    "start": { "x": 0, "y": 0, "z": 0 },\n    "rooms": []\n}\n',
    );
    const formatted = formatLayout({ start: position(0, 0, 0), rooms: [room(0, 0, 0, [], 'a "b".lua')] });
    assert.deepEqual(parseLayout(formatted).layout?.rooms[0].path, 'a "b".lua');
});
