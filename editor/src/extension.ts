import * as path from "node:path";
import * as vscode from "vscode";
import {
    LanguageClient,
    type LanguageClientOptions,
    type ServerOptions,
} from "vscode-languageclient/node";
import { MapViewProvider } from "./map_view";

let client: LanguageClient | undefined;
let mapViewProvider: MapViewProvider | undefined;

export async function activate(context: vscode.ExtensionContext): Promise<void> {
    mapViewProvider = new MapViewProvider(context.extensionUri, () => client, getDataFolderUri);
    context.subscriptions.push(...mapViewProvider.register());

    await startClient();

    context.subscriptions.push(
        vscode.workspace.onDidChangeConfiguration(async (event) => {
            if (event.affectsConfiguration("mud.serverPath") || event.affectsConfiguration("mud.dataFolder")) {
                await stopClient();
                await startClient();
            }
        }),
    );
}

export async function deactivate(): Promise<void> {
    await stopClient();
}

// Starts `mud lsp` for the data folder in the first workspace folder
async function startClient(): Promise<void> {
    const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
    if (workspaceFolder === undefined) {
        return;
    }

    const config = vscode.workspace.getConfiguration("mud");
    const serverPath = resolveServerPath(config.get<string>("serverPath", "mud"), workspaceFolder.uri.fsPath);
    const dataFolderUri = getWorkspaceDataFolderUri(workspaceFolder);

    const serverOptions: ServerOptions = {
        command: serverPath,
        args: ["lsp", dataFolderUri.fsPath],
        options: { cwd: workspaceFolder.uri.fsPath },
    };

    const clientOptions: LanguageClientOptions = {
        documentSelector: [
            { scheme: "file", language: "lua", pattern: { baseUri: dataFolderUri.toString(), pattern: "**/*.lua" } },
            { scheme: "file", pattern: { baseUri: dataFolderUri.toString(), pattern: "layout.json" } },
        ],
        synchronize: {
            // Files created, deleted or renamed outside an editor tab still need revalidation
            fileEvents: vscode.workspace.createFileSystemWatcher(new vscode.RelativePattern(dataFolderUri, "**/*")),
        },
    };

    const newClient = new LanguageClient("mud", "MUD", serverOptions, clientOptions);
    try {
        await newClient.start();
    } catch (error) {
        const message = error instanceof Error ? error.message : String(error);
        void vscode.window.showErrorMessage(
            `Failed to start '${serverPath} lsp'. Check the mud.serverPath setting. ${message}`,
        );
        return;
    }
    client = newClient;
    mapViewProvider?.onClientStarted();
}

// The data folder in the first workspace folder
function getDataFolderUri(): vscode.Uri | undefined {
    const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
    if (workspaceFolder === undefined) {
        return undefined;
    }
    return getWorkspaceDataFolderUri(workspaceFolder);
}

// The data folder in a workspace folder, from the mud.dataFolder setting
function getWorkspaceDataFolderUri(workspaceFolder: vscode.WorkspaceFolder): vscode.Uri {
    const dataFolder = vscode.workspace.getConfiguration("mud").get<string>("dataFolder", "data");
    return vscode.Uri.joinPath(workspaceFolder.uri, dataFolder);
}

async function stopClient(): Promise<void> {
    if (client === undefined) {
        return;
    }
    const oldClient = client;
    client = undefined;
    await oldClient.stop();
}

// A bare name is looked up on PATH. A relative path is resolved against the workspace folder
function resolveServerPath(serverPath: string, workspaceFolderPath: string): string {
    const isBareName = !serverPath.includes("/") && !serverPath.includes("\\");
    if (isBareName || path.isAbsolute(serverPath)) {
        return serverPath;
    }
    return path.resolve(workspaceFolderPath, serverPath);
}
