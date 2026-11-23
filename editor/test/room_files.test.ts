import assert from "node:assert/strict";
import { test } from "node:test";
import { parseNewRoomFileName, roomFileTemplate, roomPathFromFilePath } from "../src/room_files.ts";

test("roomPathFromFilePath gives the path relative to the rooms folder", () => {
    assert.equal(roomPathFromFilePath("/data/rooms", "/data/rooms/kitchen.lua"), "kitchen.lua");
    assert.equal(roomPathFromFilePath("/data/rooms/", "/data/rooms/castle/hall.lua"), "castle/hall.lua");
});

test("roomPathFromFilePath rejects files outside the rooms folder and non-Lua files", () => {
    assert.equal(roomPathFromFilePath("/data/rooms", "/data/items/bread.lua"), null);
    assert.equal(roomPathFromFilePath("/data/rooms", "/data/rooms_old/kitchen.lua"), null);
    assert.equal(roomPathFromFilePath("/data/rooms", "/data/rooms/notes.txt"), null);
    assert.equal(roomPathFromFilePath("/data/rooms", "/data/rooms"), null);
});

test("parseNewRoomFileName adds .lua and allows subfolders", () => {
    assert.deepEqual(parseNewRoomFileName(" dining_hall "), { path: "dining_hall.lua", error: null });
    assert.deepEqual(parseNewRoomFileName("castle/hall.lua"), { path: "castle/hall.lua", error: null });
});

test("parseNewRoomFileName rejects empty, absolute and escaping names", () => {
    for (const input of ["", "   ", ".lua", "/abs", "../out", "a//b", "castle/", "a\\b", "castle/.lua"]) {
        const result = parseNewRoomFileName(input);
        assert.equal(result.path, null, input);
        assert.ok(result.error, input);
    }
});

test("roomFileTemplate names the room after its file", () => {
    assert.match(roomFileTemplate("castle/dining_hall.lua"), /^room\.name = "Dining Hall"$/m);
});
