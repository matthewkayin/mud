// Helpers for the room files the map view assigns to rooms. Room paths are relative to the rooms folder and always use
// "/" separators, as layout.json stores them.
//
// Imports use .ts extensions so `npm test` can run this file directly with Node
import * as path from "node:path";

const ROOM_FILE_EXTENSION = ".lua";

export type ParseRoomFileNameResult = {
    path: string | null;
    error: string | null;
};

// The path of a file relative to the rooms folder, or null if it isn't a .lua file inside it. Both paths are file
// system paths
export function roomPathFromFilePath(roomsFolderPath: string, filePath: string): string | null {
    const relativePath = path.relative(path.resolve(roomsFolderPath), path.resolve(filePath));
    if (relativePath === "" || relativePath.startsWith("..") || path.isAbsolute(relativePath)) {
        return null;
    }
    if (!relativePath.endsWith(ROOM_FILE_EXTENSION)) {
        return null;
    }
    return relativePath.split(path.sep).join("/");
}

// Turns a name the author typed into a room path. ".lua" is added if missing, and subfolders are allowed
export function parseNewRoomFileName(input: string): ParseRoomFileNameResult {
    const name = input.trim();
    if (name === "") {
        return { path: null, error: "Enter a file name" };
    }
    if (name.includes("\\")) {
        return { path: null, error: "Use / to separate folders" };
    }
    if (name.startsWith("/")) {
        return { path: null, error: "The name must be relative to the rooms folder" };
    }
    const segments = name.split("/");
    for (const segment of segments) {
        if (segment === "" || segment === "." || segment === "..") {
            return { path: null, error: "Folder and file names can't be empty, '.' or '..'" };
        }
    }
    const roomPath = name.endsWith(ROOM_FILE_EXTENSION) ? name : name + ROOM_FILE_EXTENSION;
    if (roomPath === ROOM_FILE_EXTENSION || roomPath.endsWith("/" + ROOM_FILE_EXTENSION)) {
        return { path: null, error: "Enter a file name" };
    }
    return { path: roomPath, error: null };
}

// The text of a new room file. Its name comes from the file name, so dining_hall.lua is "Dining Hall"
export function roomFileTemplate(roomPath: string): string {
    const lines = [
        "local room = {}",
        "",
        `room.name = ${JSON.stringify(roomNameFromPath(roomPath))}`,
        `room.description = ""`,
        "",
        "return room",
    ];
    return lines.join("\n") + "\n";
}

function roomNameFromPath(roomPath: string): string {
    const fileName = roomPath.split("/").at(-1)!;
    const baseName = fileName.slice(0, fileName.length - ROOM_FILE_EXTENSION.length);
    return baseName
        .split(/[_\-\s]+/)
        .filter((word) => word !== "")
        .map((word) => word[0].toUpperCase() + word.slice(1))
        .join(" ");
}
