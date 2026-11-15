import { randomBytes } from "node:crypto";
import * as vscode from "vscode";
import type { LanguageClient } from "vscode-languageclient/node";
import type { Layout, MapViewRequest, MapViewState } from "../shared/map_types";
import { parseLayout } from "./layout";
import { createRoom, deleteRoom, formatLayout, moveRoom, setStart, toggleExit } from "./layout_edit";

const ROOM_NAMES_REFRESH_DEBOUNCE_MS = 300;

type MapPanel = {
    document: vscode.TextDocument;
    webview: vscode.Webview;
    // Messages posted before the webview script loads are lost, so state is only sent once it says it's ready
    isReady: boolean;
};

// The map view: a custom editor for layout.json that draws the rooms on a grid
export class MapViewProvider implements vscode.CustomTextEditorProvider {
    static readonly VIEW_TYPE = "mud.mapView";

    private readonly panels = new Set<MapPanel>();
    private roomNames: Record<string, string> = {};
    private roomNamesRefreshTimer: NodeJS.Timeout | undefined;
    // Only the response to the latest request is used, in case responses arrive out of order
    private roomNamesRequestCount = 0;

    constructor(
        private readonly extensionUri: vscode.Uri,
        private readonly getClient: () => LanguageClient | undefined,
        private readonly getDataFolderUri: () => vscode.Uri | undefined,
    ) {}

    register(): vscode.Disposable[] {
        // Room names come from Lua files, so any change to one may rename a room. The validation server only republishes
        // diagnostics when problems change, so it can't be relied on to signal a rename
        const luaWatcher = vscode.workspace.createFileSystemWatcher("**/*.lua");
        const onLuaFileEvent = (uri: vscode.Uri) => {
            if (this.isInDataFolder(uri)) {
                this.scheduleRoomNamesRefresh();
            }
        };

        return [
            vscode.window.registerCustomEditorProvider(MapViewProvider.VIEW_TYPE, this),
            vscode.workspace.onDidChangeTextDocument((event) => {
                if (event.document.languageId === "lua" && this.isInDataFolder(event.document.uri)) {
                    this.scheduleRoomNamesRefresh();
                }
            }),
            luaWatcher,
            luaWatcher.onDidCreate(onLuaFileEvent),
            luaWatcher.onDidChange(onLuaFileEvent),
            luaWatcher.onDidDelete(onLuaFileEvent),
            { dispose: () => clearTimeout(this.roomNamesRefreshTimer) },
        ];
    }

    // Called whenever the language client starts, since the old client's names may be out of date
    onClientStarted(): void {
        void this.refreshRoomNames();
    }

    resolveCustomTextEditor(document: vscode.TextDocument, webviewPanel: vscode.WebviewPanel): void {
        const distUri = vscode.Uri.joinPath(this.extensionUri, "dist");
        webviewPanel.webview.options = {
            enableScripts: true,
            localResourceRoots: [distUri],
        };
        webviewPanel.webview.html = this.getHtml(webviewPanel.webview, vscode.Uri.joinPath(distUri, "webview.js"));

        const panel: MapPanel = { document, webview: webviewPanel.webview, isReady: false };
        this.panels.add(panel);

        const subscriptions = [
            webviewPanel.webview.onDidReceiveMessage((request: MapViewRequest) => {
                if (request.isReady === true) {
                    panel.isReady = true;
                    this.postState(panel);
                    void this.refreshRoomNames();
                }
                void this.applyEditRequest(document, request);
            }),
            vscode.workspace.onDidChangeTextDocument((event) => {
                if (event.document.uri.toString() === document.uri.toString()) {
                    this.postState(panel);
                }
            }),
        ];

        webviewPanel.onDidDispose(() => {
            this.panels.delete(panel);
            for (const subscription of subscriptions) {
                subscription.dispose();
            }
        });
    }

    // Applies an edit intent to the document's current text as one WorkspaceEdit, so VS Code's text buffer gives undo,
    // redo, unsaved state and saving. The whole document is rewritten in the fixed format
    private async applyEditRequest(document: vscode.TextDocument, request: MapViewRequest): Promise<void> {
        const editLayout = getEditFunction(request);
        if (editLayout === null) {
            return;
        }
        // While layout.json doesn't parse the map view shows the error, so there is nothing to edit
        const layout = parseLayout(document.getText()).layout;
        if (layout === null) {
            return;
        }
        const newLayout = editLayout(layout);
        if (newLayout === null) {
            return;
        }

        const fullRange = new vscode.Range(document.positionAt(0), document.positionAt(document.getText().length));
        const edit = new vscode.WorkspaceEdit();
        edit.replace(document.uri, fullRange, formatLayout(newLayout));
        const isApplied = await vscode.workspace.applyEdit(edit);
        if (!isApplied) {
            console.warn(`Failed to apply map edit to ${document.uri.toString()}`);
        }
    }

    private postState(panel: MapPanel): void {
        if (!panel.isReady) {
            return;
        }
        const parseResult = parseLayout(panel.document.getText());
        const state: MapViewState = {
            layout: parseResult.layout,
            layoutError: parseResult.error,
            roomNames: this.roomNames,
        };
        void panel.webview.postMessage(state);
    }

    private scheduleRoomNamesRefresh(): void {
        if (this.panels.size === 0) {
            return;
        }
        clearTimeout(this.roomNamesRefreshTimer);
        this.roomNamesRefreshTimer = setTimeout(() => {
            void this.refreshRoomNames();
        }, ROOM_NAMES_REFRESH_DEBOUNCE_MS);
    }

    private async refreshRoomNames(): Promise<void> {
        clearTimeout(this.roomNamesRefreshTimer);
        const client = this.getClient();
        if (client === undefined || this.panels.size === 0) {
            return;
        }

        this.roomNamesRequestCount++;
        const requestNumber = this.roomNamesRequestCount;
        let roomNames: Record<string, string>;
        try {
            roomNames = await client.sendRequest<Record<string, string>>("mud/roomNames");
        } catch (error) {
            // The client may have stopped while the request was in flight. The next start refreshes again
            const message = error instanceof Error ? error.message : String(error);
            console.warn(`mud/roomNames request failed: ${message}`);
            return;
        }
        if (requestNumber !== this.roomNamesRequestCount) {
            return;
        }

        this.roomNames = roomNames ?? {};
        for (const panel of this.panels) {
            this.postState(panel);
        }
    }

    private isInDataFolder(uri: vscode.Uri): boolean {
        const dataFolderUri = this.getDataFolderUri();
        if (dataFolderUri === undefined) {
            return false;
        }
        const dataFolderPath = dataFolderUri.toString().replace(/\/$/, "") + "/";
        return uri.toString().startsWith(dataFolderPath);
    }

    private getHtml(webview: vscode.Webview, scriptUri: vscode.Uri): string {
        const nonce = randomBytes(16).toString("hex");
        return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta http-equiv="Content-Security-Policy" content="default-src 'none'; style-src ${webview.cspSource} 'unsafe-inline'; script-src 'nonce-${nonce}';">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Map</title>
    <style>
        html, body {
            margin: 0;
            padding: 0;
            width: 100%;
            height: 100%;
            overflow: hidden;
        }
        #map {
            position: absolute;
            inset: 0;
            width: 100%;
            height: 100%;
            display: block;
        }
        #elevation {
            position: absolute;
            top: 8px;
            right: 8px;
            display: flex;
            align-items: center;
            gap: 6px;
            padding: 4px;
            background: var(--vscode-editorWidget-background);
            border: 1px solid var(--vscode-editorWidget-border, transparent);
            border-radius: 4px;
            color: var(--vscode-editorWidget-foreground, var(--vscode-editor-foreground));
            font-family: var(--vscode-font-family);
            font-size: var(--vscode-font-size);
            user-select: none;
        }
        #elevation button {
            width: 24px;
            height: 24px;
            padding: 0;
            border: none;
            border-radius: 3px;
            background: var(--vscode-button-secondaryBackground);
            color: var(--vscode-button-secondaryForeground);
            font-size: 16px;
            line-height: 24px;
            cursor: pointer;
        }
        #elevation button:hover {
            background: var(--vscode-button-secondaryHoverBackground);
        }
        #elevation-label {
            min-width: 40px;
            text-align: center;
        }
        #room-actions {
            position: absolute;
            transform: translate(-50%, -100%);
            display: flex;
            gap: 4px;
            padding: 4px;
            background: var(--vscode-editorWidget-background);
            border: 1px solid var(--vscode-editorWidget-border, transparent);
            border-radius: 4px;
            font-family: var(--vscode-font-family);
            font-size: var(--vscode-font-size);
            white-space: nowrap;
            user-select: none;
        }
        #room-actions button {
            padding: 2px 8px;
            border: none;
            border-radius: 3px;
            background: var(--vscode-button-secondaryBackground);
            color: var(--vscode-button-secondaryForeground);
            font: inherit;
            cursor: pointer;
        }
        #room-actions button:hover:not(:disabled) {
            background: var(--vscode-button-secondaryHoverBackground);
        }
        #room-actions button:disabled {
            opacity: 0.5;
            cursor: default;
        }
        #error {
            padding: 16px;
            color: var(--vscode-errorForeground);
            font-family: var(--vscode-editor-font-family);
            white-space: pre-wrap;
        }
        [hidden] {
            display: none !important;
        }
    </style>
</head>
<body>
    <canvas id="map"></canvas>
    <div id="elevation">
        <button id="elevation-down" title="Go down a level">&minus;</button>
        <span id="elevation-label">Z: 0</span>
        <button id="elevation-up" title="Go up a level">+</button>
    </div>
    <div id="room-actions" hidden>
        <button id="set-start" title="New characters start in this room">Set as start</button>
        <button id="delete-room" title="Delete this room and its connections (Delete)">Delete</button>
    </div>
    <div id="error" hidden></div>
    <script nonce="${nonce}" src="${webview.asWebviewUri(scriptUri).toString()}"></script>
</body>
</html>`;
    }
}

// The layout edit a request asks for, or null if it isn't an edit
function getEditFunction(request: MapViewRequest): ((layout: Layout) => Layout | null) | null {
    const createPosition = request.createRoom;
    const deletePosition = request.deleteRoom;
    const move = request.moveRoom;
    const toggle = request.toggleExit;
    const startPosition = request.setStart;
    if (createPosition !== undefined) {
        return (layout) => createRoom(layout, createPosition);
    }
    if (deletePosition !== undefined) {
        return (layout) => deleteRoom(layout, deletePosition);
    }
    if (move !== undefined) {
        return (layout) => moveRoom(layout, move.from, move.to);
    }
    if (toggle !== undefined) {
        return (layout) => toggleExit(layout, toggle.position, toggle.direction);
    }
    if (startPosition !== undefined) {
        return (layout) => setStart(layout, startPosition);
    }
    return null;
}
