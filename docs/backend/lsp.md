# Validation Language Server

`mud lsp` is a language server that validates the world's data folder (see `world_data.md`) while an author edits it. It runs the same loader the game server uses, so anything that would stop the server from starting shows up in the editor as an error, on the file and line it's about. It speaks the Language Server Protocol over stdio, so it works in any editor with LSP support.

The code is in `backend/lsp/`.

## Running

```
mud lsp <data folder>
```

From `backend/` during development:

```
go run . lsp data
```

The data folder is relative to the working directory. Stdout carries the protocol, so nothing else writes to it. Logs go to stderr, and so does the output of a script calling `print` while it loads. LSP mode doesn't read `.env` and doesn't create `logs/` or `saves/`.

## What it validates

Everything `world.NewWorld` reports as a `ScriptProblem`: Lua syntax and runtime errors in the top level of scripts, missing, invalid or unknown fields, references to data that doesn't exist (such as an unknown item in a loot table), missing core folders, and every `layout.json` problem. Errors and warnings keep their severity.

It doesn't run behavior functions such as a room's `reset`, so errors inside them only show up when the game runs them.

Problems can span files. A file with an error is left out of the world, so everything that refers to it gets an error too. For example, breaking an item file adds errors to every room whose loot names that item.

## How it works

- The server keeps the text of every open document inside the data folder. The loader reads the data folder through an overlay filesystem that serves that text in place of the file on disk, so unsaved changes are validated. Everything else, including directory listings, comes from disk.
- Opening, changing, closing or saving a document, or a `workspace/didChangeWatchedFiles` notification, schedules a validation. Validation waits until 300ms pass without another change.
- Each validation loads a new world from the whole data folder, closes it, and publishes diagnostics for every file with problems. Files that had problems last time and don't now get an empty list, so stale diagnostics clear.
- One goroutine owns all the server's state and is the only writer to stdout. Another goroutine reads messages from stdin and passes them to it over a channel.

It handles `initialize`, `initialized`, `shutdown`, `exit`, `textDocument/didOpen`, `didChange`, `didClose`, `didSave`, `workspace/didChangeWatchedFiles` and `mud/roomNames`. It uses full document sync. Other requests get a "method not found" error, and other notifications are ignored. The server doesn't register for file watching itself. The client has to send `workspace/didChangeWatchedFiles` for files created, deleted or renamed outside the editor.

## Where diagnostics go

A problem's position comes from the first of these that applies (`backend/lsp/lines.go`):

1. A Lua error with a line number goes on that line.
2. A problem about a field goes on the first place the field's key is written as a table key: `key =`, `["key"] =` or `['key'] =`. A key used more than once in a file may put the diagnostic on the wrong one.
3. A `layout.json` problem about a room entry goes on the line where that entry starts.
4. Anything else goes on line 1.

A missing core folder's problem is published on the folder itself.

## `mud/roomNames`

A custom request with no params. The result maps the path of every room file that loaded, relative to `rooms/`, to its `name`:

```json
{ "kitchen.lua": "Kitchen", "church.lua": "Church" }
```

Room files that failed to load are left out. If a validation is waiting on the debounce when the request arrives, it runs first, so the names match the current text. The map view uses this request for its room labels.

## Other editors

Point the editor at `mud lsp data` for Lua and JSON files, and start it from the folder that contains the data folder (for example, `backend/`).

Neovim (0.11 or later):

```lua
vim.lsp.config('mud', {
    cmd = { 'mud', 'lsp', 'data' },
    filetypes = { 'lua', 'json' },
})
vim.lsp.enable('mud')
```

Helix (`languages.toml`):

```toml
[language-server.mud]
command = "mud"
args = ["lsp", "data"]

[[language]]
name = "lua"
language-servers = ["lua-language-server", "mud"]

[[language]]
name = "json"
language-servers = ["vscode-json-language-server", "mud"]
```

## Tests

`lines_test.go` and `overlay_test.go` test line finding and the overlay filesystem with in-memory data. There are no tests that read a data folder from disk.
