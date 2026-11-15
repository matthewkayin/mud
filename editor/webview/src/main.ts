import type { MapViewRequest, MapViewState } from "../../shared/map_types";
import { cellCenter, renderMap, screenToWorld, type Camera, type Colors, type Viewport } from "./render";

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

const ZOOM_MIN = 0.25;
const ZOOM_MAX = 2.0;
const ZOOM_SPEED = 0.0015;
const MOUSE_BUTTON_RIGHT = 2;
// The right button's bit in MouseEvent.buttons
const MOUSE_BUTTONS_RIGHT_MASK = 2;

const vscode = acquireVsCodeApi();

const canvas = getElement<HTMLCanvasElement>("map");
const elevationControls = getElement<HTMLDivElement>("elevation");
const elevationLabel = getElement<HTMLSpanElement>("elevation-label");
const elevationDownButton = getElement<HTMLButtonElement>("elevation-down");
const elevationUpButton = getElement<HTMLButtonElement>("elevation-up");
const errorPanel = getElement<HTMLDivElement>("error");

const context = canvas.getContext("2d")!;

let mapState: MapViewState | null = null;
// Null until the first layout arrives, when the view centers on the start room
let viewState: SavedViewState | null = vscode.getState() ?? null;
let viewport: Viewport = { width: 0, height: 0, pixelRatio: 1 };
let colors: Colors = readColors();

let isPanning = false;
let lastMouseX = 0;
let lastMouseY = 0;

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
    updateControls();
    requestRender();
});

// Input

canvas.addEventListener("contextmenu", (event) => event.preventDefault());

canvas.addEventListener("mousedown", (event) => {
    if (event.button === MOUSE_BUTTON_RIGHT) {
        isPanning = true;
        lastMouseX = event.clientX;
        lastMouseY = event.clientY;
    }
});

window.addEventListener("mousemove", (event) => {
    if (!isPanning || viewState === null) {
        return;
    }
    // A mouseup outside the webview is never delivered, so check the button is still held
    if ((event.buttons & MOUSE_BUTTONS_RIGHT_MASK) === 0) {
        isPanning = false;
        saveViewState();
        return;
    }
    viewState.camera.centerX -= (event.clientX - lastMouseX) / viewState.camera.zoom;
    viewState.camera.centerY -= (event.clientY - lastMouseY) / viewState.camera.zoom;
    lastMouseX = event.clientX;
    lastMouseY = event.clientY;
    requestRender();
});

window.addEventListener("mouseup", (event) => {
    if (event.button === MOUSE_BUTTON_RIGHT && isPanning) {
        isPanning = false;
        saveViewState();
    }
});

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
        requestRender();
    },
    { passive: false },
);

elevationDownButton.addEventListener("click", () => changeElevation(-1));
elevationUpButton.addEventListener("click", () => changeElevation(1));

function changeElevation(change: number): void {
    if (viewState === null) {
        return;
    }
    viewState.elevation += change;
    saveViewState();
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
}

// The map only redraws when something changes, so anything that affects the picture must call this
function requestRender(): void {
    if (isRenderRequested) {
        return;
    }
    isRenderRequested = true;
    requestAnimationFrame(() => {
        isRenderRequested = false;
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
