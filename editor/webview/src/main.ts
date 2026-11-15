import { findRoom, positionsEqual, type MapViewRequest, type MapViewState, type Position } from "../../shared/map_types";
import { cellAt, hitTargetsEqual, hitTest, type HitTarget } from "./hit_test";
import {
    cellCenter,
    cellRect,
    isValidDragTarget,
    renderMap,
    screenToWorld,
    worldToScreen,
    type Camera,
    type Colors,
    type RoomDrag,
    type Viewport,
} from "./render";

type VsCodeApi = {
    postMessage(request: MapViewRequest): void;
    getState(): SavedViewState | undefined;
    setState(state: SavedViewState): void;
};

declare function acquireVsCodeApi(): VsCodeApi;

// Kept by VS Code across tab switches and reloads, so the view doesn't jump back to the start room
type SavedViewState = {
    camera: Camera;
    elevation: number;
};

// A left button press that hasn't been released yet. It becomes a drag if it started on a room and the mouse moves far
// enough, and a click otherwise
type LeftPress = {
    clientX: number;
    clientY: number;
    target: HitTarget | null;
};

const ZOOM_MIN = 0.25;
const ZOOM_MAX = 2.0;
const ZOOM_SPEED = 0.0015;
const MOUSE_BUTTON_LEFT = 0;
const MOUSE_BUTTON_RIGHT = 2;
// The buttons' bits in MouseEvent.buttons
const MOUSE_BUTTONS_LEFT_MASK = 1;
const MOUSE_BUTTONS_RIGHT_MASK = 2;
// How far, in CSS pixels, the mouse must move with the left button held before a press on a room becomes a drag
const DRAG_THRESHOLD = 4;
// The gap between the selected room's top edge and the room actions bar, in CSS pixels
const ROOM_ACTIONS_OFFSET = 10;

const vscode = acquireVsCodeApi();

const canvas = getElement<HTMLCanvasElement>("map");
const elevationControls = getElement<HTMLDivElement>("elevation");
const elevationLabel = getElement<HTMLSpanElement>("elevation-label");
const elevationDownButton = getElement<HTMLButtonElement>("elevation-down");
const elevationUpButton = getElement<HTMLButtonElement>("elevation-up");
const roomActions = getElement<HTMLDivElement>("room-actions");
const setStartButton = getElement<HTMLButtonElement>("set-start");
const assignFileButton = getElement<HTMLButtonElement>("assign-file");
const newFileButton = getElement<HTMLButtonElement>("new-file");
const deleteRoomButton = getElement<HTMLButtonElement>("delete-room");
const errorPanel = getElement<HTMLDivElement>("error");

const context = canvas.getContext("2d")!;

let mapState: MapViewState | null = null;
// Null until the first layout arrives, when the view centers on the start room
let viewState: SavedViewState | null = vscode.getState() ?? null;
let viewport: Viewport = { width: 0, height: 0, pixelRatio: 1 };
let colors: Colors = readColors();

let isPanning = false;
let panLastX = 0;
let panLastY = 0;
// The mouse's last position in client coordinates, for redoing the hover hit test when the map changes under it
let mouseX = 0;
let mouseY = 0;

let hoverTarget: HitTarget | null = null;
let leftPress: LeftPress | null = null;
let drag: RoomDrag | null = null;
let selectedPosition: Position | null = null;
// The room a file from the explorer is being dragged over
let fileDropPosition: Position | null = null;

let isRenderRequested = false;

// Messages from the extension

window.addEventListener("message", (event: MessageEvent<MapViewState>) => {
    mapState = event.data;
    if (viewState === null && mapState.layout !== null) {
        const start = mapState.layout.start;
        const center = cellCenter(start.x, start.y);
        viewState = { camera: { centerX: center.x, centerY: center.y, zoom: 1 }, elevation: start.z };
        saveViewState();
    }
    // The selected room may have been deleted or moved, by the map or by editing the JSON
    if (selectedPosition !== null && (mapState.layout === null || findRoom(mapState.layout, selectedPosition) === undefined)) {
        selectedPosition = null;
    }
    if (drag !== null && (mapState.layout === null || findRoom(mapState.layout, drag.from) === undefined)) {
        drag = null;
    }
    updateHoverTarget();
    updateControls();
    requestRender();
});

// Mouse input. Edits are sent to the extension, which applies them to layout.json and sends back the new state

canvas.addEventListener("contextmenu", (event) => event.preventDefault());

canvas.addEventListener("mousedown", (event) => {
    if (event.button === MOUSE_BUTTON_RIGHT) {
        isPanning = true;
        panLastX = event.clientX;
        panLastY = event.clientY;
        updateCursor();
    } else if (event.button === MOUSE_BUTTON_LEFT) {
        leftPress = { clientX: event.clientX, clientY: event.clientY, target: hitTestClient(event.clientX, event.clientY) };
    }
});

window.addEventListener("mousemove", (event) => {
    mouseX = event.clientX;
    mouseY = event.clientY;
    if (isPanning) {
        pan(event);
        return;
    }

    // A mouseup outside the webview is never delivered, so check the button is still held
    if (leftPress !== null && (event.buttons & MOUSE_BUTTONS_LEFT_MASK) === 0) {
        leftPress = null;
        drag = null;
    }
    if (leftPress !== null && drag === null && leftPress.target?.kind === "room") {
        const distance = Math.hypot(event.clientX - leftPress.clientX, event.clientY - leftPress.clientY);
        if (distance >= DRAG_THRESHOLD) {
            drag = { from: leftPress.target.position, to: null };
        }
    }
    if (drag !== null && viewState !== null) {
        const world = clientToWorld(event.clientX, event.clientY);
        drag.to = cellAt(world.x, world.y, viewState.elevation);
    }

    updateHoverTarget();
    updateCursor();
    requestRender();
});

canvas.addEventListener("mouseleave", () => {
    if (!isPanning && drag === null) {
        setHoverTarget(null);
    }
});

window.addEventListener("mouseup", (event) => {
    if (event.button === MOUSE_BUTTON_RIGHT && isPanning) {
        isPanning = false;
        saveViewState();
        updateHoverTarget();
        updateCursor();
        return;
    }
    if (event.button !== MOUSE_BUTTON_LEFT || leftPress === null) {
        return;
    }

    if (drag !== null) {
        finishDrag(drag);
    } else if (event.target === canvas) {
        // Only act if the press and release are on the same target, so a slip doesn't create a room
        const target = hitTestClient(event.clientX, event.clientY);
        if (hitTargetsEqual(target, leftPress.target)) {
            onClick(target);
        }
    }
    leftPress = null;
    drag = null;
    updateHoverTarget();
    updateCursor();
    requestRender();
});

function onClick(target: HitTarget | null): void {
    if (target === null) {
        setSelectedPosition(null);
        return;
    }
    switch (target.kind) {
        case "emptyCell":
            setSelectedPosition(null);
            vscode.postMessage({ createRoom: target.position });
            break;
        case "gap":
        case "verticalArrow":
            vscode.postMessage({ toggleExit: { position: target.position, direction: target.direction } });
            break;
        case "room":
            setSelectedPosition(target.position);
            break;
    }
}

// Double-clicking a room opens its file. The first click of the two has already selected it
canvas.addEventListener("dblclick", (event) => {
    const target = hitTestClient(event.clientX, event.clientY);
    if (target?.kind !== "room") {
        return;
    }
    const room = mapState?.layout == null ? undefined : findRoom(mapState.layout, target.position);
    if (room !== undefined && room.path !== "") {
        vscode.postMessage({ openRoomFile: target.position });
    }
});

// Files dragged from the explorer. VS Code only delivers them to a webview while Shift is held

canvas.addEventListener("dragover", (event) => {
    event.preventDefault();
    const target = hitTestClient(event.clientX, event.clientY);
    setFileDropPosition(target?.kind === "room" ? target.position : null);
});

canvas.addEventListener("dragleave", () => setFileDropPosition(null));

canvas.addEventListener("drop", (event) => {
    event.preventDefault();
    setFileDropPosition(null);
    const target = hitTestClient(event.clientX, event.clientY);
    const uri = readDroppedUri(event.dataTransfer);
    if (target?.kind === "room" && uri !== null) {
        vscode.postMessage({ dropRoomFile: { position: target.position, uri } });
    }
});

// The first URI in a drop's text/uri-list, which has one URI per line and # comment lines
function readDroppedUri(dataTransfer: DataTransfer | null): string | null {
    const uriList = dataTransfer?.getData("text/uri-list") ?? "";
    const uri = uriList
        .split(/\r?\n/)
        .map((line) => line.trim())
        .find((line) => line !== "" && !line.startsWith("#"));
    return uri ?? null;
}

function setFileDropPosition(position: Position | null): void {
    const isSame =
        position === null || fileDropPosition === null ? position === fileDropPosition : positionsEqual(position, fileDropPosition);
    if (isSame) {
        return;
    }
    fileDropPosition = position;
    requestRender();
}

function finishDrag(finishedDrag: RoomDrag): void {
    if (mapState?.layout == null || finishedDrag.to === null || !isValidDragTarget(mapState.layout, finishedDrag)) {
        return;
    }
    vscode.postMessage({ moveRoom: { from: finishedDrag.from, to: finishedDrag.to } });
    // A selected room stays selected at its new cell
    if (selectedPosition !== null && positionsEqual(selectedPosition, finishedDrag.from)) {
        selectedPosition = finishedDrag.to;
    }
}

function pan(event: MouseEvent): void {
    if (viewState === null) {
        return;
    }
    // A mouseup outside the webview is never delivered, so check the button is still held
    if ((event.buttons & MOUSE_BUTTONS_RIGHT_MASK) === 0) {
        isPanning = false;
        saveViewState();
        updateCursor();
        return;
    }
    viewState.camera.centerX -= (event.clientX - panLastX) / viewState.camera.zoom;
    viewState.camera.centerY -= (event.clientY - panLastY) / viewState.camera.zoom;
    panLastX = event.clientX;
    panLastY = event.clientY;
    requestRender();
}

// Zooms around the cursor, so the point under it stays put
canvas.addEventListener(
    "wheel",
    (event) => {
        event.preventDefault();
        if (viewState === null) {
            return;
        }
        const camera = viewState.camera;
        const canvasBounds = canvas.getBoundingClientRect();
        const mouseX = event.clientX - canvasBounds.left;
        const mouseY = event.clientY - canvasBounds.top;

        const worldBefore = screenToWorld(viewport, camera, mouseX, mouseY);
        camera.zoom = clamp(camera.zoom * Math.exp(-event.deltaY * ZOOM_SPEED), ZOOM_MIN, ZOOM_MAX);
        const worldAfter = screenToWorld(viewport, camera, mouseX, mouseY);
        camera.centerX += worldBefore.x - worldAfter.x;
        camera.centerY += worldBefore.y - worldAfter.y;

        saveViewState();
        updateHoverTarget();
        requestRender();
    },
    { passive: false },
);

function clientToWorld(clientX: number, clientY: number): { x: number; y: number } {
    const canvasBounds = canvas.getBoundingClientRect();
    return screenToWorld(viewport, viewState!.camera, clientX - canvasBounds.left, clientY - canvasBounds.top);
}

function hitTestClient(clientX: number, clientY: number): HitTarget | null {
    if (mapState?.layout == null || viewState === null) {
        return null;
    }
    const world = clientToWorld(clientX, clientY);
    return hitTest(mapState.layout, viewState.elevation, world.x, world.y);
}

// Recomputes what's under the mouse, since the layout or camera may have changed under it
function updateHoverTarget(): void {
    const isOverCanvas = document.elementFromPoint(mouseX, mouseY) === canvas;
    setHoverTarget(isOverCanvas && !isPanning ? hitTestClient(mouseX, mouseY) : null);
}

function setHoverTarget(target: HitTarget | null): void {
    if (hitTargetsEqual(target, hoverTarget)) {
        return;
    }
    hoverTarget = target;
    updateCursor();
    requestRender();
}

function updateCursor(): void {
    if (isPanning) {
        canvas.style.cursor = "grabbing";
    } else if (drag !== null) {
        canvas.style.cursor = "move";
    } else if (hoverTarget !== null) {
        canvas.style.cursor = "pointer";
    } else {
        canvas.style.cursor = "default";
    }
}

// Keyboard input

window.addEventListener("keydown", (event) => {
    if (selectedPosition === null) {
        return;
    }
    if (event.key === "Escape") {
        setSelectedPosition(null);
    } else if (event.key === "Delete" || event.key === "Backspace") {
        event.preventDefault();
        vscode.postMessage({ deleteRoom: selectedPosition });
    }
});

// Controls

elevationDownButton.addEventListener("click", () => changeElevation(-1));
elevationUpButton.addEventListener("click", () => changeElevation(1));

setStartButton.addEventListener("click", () => {
    if (selectedPosition !== null) {
        vscode.postMessage({ setStart: selectedPosition });
    }
});

assignFileButton.addEventListener("click", () => {
    if (selectedPosition !== null) {
        vscode.postMessage({ chooseRoomFile: selectedPosition });
    }
});

newFileButton.addEventListener("click", () => {
    if (selectedPosition !== null) {
        vscode.postMessage({ createRoomFile: selectedPosition });
    }
});

deleteRoomButton.addEventListener("click", () => {
    if (selectedPosition !== null) {
        vscode.postMessage({ deleteRoom: selectedPosition });
    }
});

function changeElevation(change: number): void {
    if (viewState === null) {
        return;
    }
    viewState.elevation += change;
    selectedPosition = null;
    hoverTarget = null;
    saveViewState();
    updateControls();
    requestRender();
}

function setSelectedPosition(position: Position | null): void {
    selectedPosition = position;
    updateControls();
    requestRender();
}

// Size and theme

// Resizing the canvas clears it, so it always needs a redraw
new ResizeObserver(() => {
    viewport = { width: canvas.clientWidth, height: canvas.clientHeight, pixelRatio: window.devicePixelRatio };
    canvas.width = Math.round(viewport.width * viewport.pixelRatio);
    canvas.height = Math.round(viewport.height * viewport.pixelRatio);
    requestRender();
}).observe(canvas);

// VS Code changes the body's class when the theme changes
new MutationObserver(() => {
    colors = readColors();
    requestRender();
}).observe(document.body, { attributes: true, attributeFilter: ["class"] });

function readColors(): Colors {
    const style = getComputedStyle(document.documentElement);
    const read = (name: string, fallback: string) => {
        const value = style.getPropertyValue(name).trim();
        return value === "" ? fallback : value;
    };
    const background = read("--vscode-editor-background", "#1e1e1e");
    const foreground = read("--vscode-editor-foreground", "#d4d4d4");
    return {
        background,
        foreground,
        roomFill: read("--vscode-editorWidget-background", background),
        accent: read("--vscode-focusBorder", "#007fd4"),
        dim: read("--vscode-descriptionForeground", foreground),
        fontFamily: read("--vscode-font-family", "sans-serif"),
    };
}

// Drawing

function updateControls(): void {
    const layoutError = mapState?.layoutError ?? null;
    errorPanel.hidden = layoutError === null;
    errorPanel.textContent = layoutError ?? "";
    canvas.hidden = layoutError !== null;
    elevationControls.hidden = layoutError !== null || viewState === null;
    if (viewState !== null) {
        elevationLabel.textContent = `Z: ${viewState.elevation}`;
    }
    updateRoomActions();
}

// The room actions bar floats above the selected room, so it moves whenever the map is drawn
function updateRoomActions(): void {
    const layout = mapState?.layout ?? null;
    if (layout === null || viewState === null || selectedPosition === null) {
        roomActions.hidden = true;
        return;
    }
    const rect = cellRect(selectedPosition.x, selectedPosition.y);
    const topCenter = worldToScreen(viewport, viewState.camera, rect.x + rect.width / 2, rect.y);
    roomActions.style.left = `${topCenter.x}px`;
    roomActions.style.top = `${topCenter.y - ROOM_ACTIONS_OFFSET}px`;
    setStartButton.disabled = positionsEqual(layout.start, selectedPosition);
    const roomPath = findRoom(layout, selectedPosition)?.path ?? "";
    assignFileButton.title = roomPath === "" ? "Choose this room's Lua file" : `Change this room's Lua file (rooms/${roomPath})`;
    newFileButton.hidden = roomPath !== "";
    roomActions.hidden = false;
}

// The map only redraws when something changes, so anything that affects the picture must call this
function requestRender(): void {
    if (isRenderRequested) {
        return;
    }
    isRenderRequested = true;
    requestAnimationFrame(() => {
        isRenderRequested = false;
        updateRoomActions();
        if (mapState === null || mapState.layout === null || viewState === null) {
            return;
        }
        renderMap(context, {
            viewport,
            camera: viewState.camera,
            elevation: viewState.elevation,
            layout: mapState.layout,
            roomNames: mapState.roomNames,
            colors,
            hoverTarget,
            selectedPosition,
            drag,
            fileDropPosition,
        });
    });
}

function saveViewState(): void {
    if (viewState !== null) {
        vscode.setState(viewState);
    }
}

function clamp(value: number, min: number, max: number): number {
    return Math.min(Math.max(value, min), max);
}

function getElement<T extends HTMLElement>(id: string): T {
    const element = document.getElementById(id);
    if (element === null) {
        throw new Error(`Missing element #${id}`);
    }
    return element as T;
}

updateControls();
vscode.postMessage({ isReady: true });
