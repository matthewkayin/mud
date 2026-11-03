# PRD: Mudkit

## Outcome
This MUD should become a platform that other content authors can use to make their own adventures.

## Users
People who are interested in making an adventure for a fantasy MUD. The target user is a storyteller who has a non-trivial amount of technical skill. They are primarily interested in writing an
adventure, but they are not afraid of getting hands-on with Lua scripts and NPC behaviors. The user is interested in making content in an existing system and hosting the server on their own.

## Constraints
Rules every story must follow: stack, libraries to use or avoid,
security requirements, code conventions. Write them once, here.

## Release 1
The smallest set of stories that delivers the outcome. Build these first, in order.

### Story: Host Your Own World
As an author, I want to be able to download this MUD and host my own instance of it. I don't want to compile the MUD server myself.

Requirements:
- **Release binaries.** A GitHub Actions workflow builds the server for linux/amd64, linux/arm64, darwin/arm64 and windows/amd64 when a version tag is pushed, and attaches the binaries to a GitHub Release. The web client and the editor are embedded in the binary. The build uses no cgo.
- **Starter world.** Each release also includes a minimal starter world folder that boots as is. It contains a `data/` folder with the start room at 0,0,0, a couple of linked rooms, one race, one class, a few items, one NPC and one spell, plus `world.d.lua` for the Lua language server. It also contains a `banner.txt`, a `.env.example` and a `.gitignore` that ignores `saves/`, `logs/` and `.env`. The starter world is kept in this repo and is separate from RC Disco's `backend/data`.
- **The working directory is the world folder.** The server runs from the author's world folder and uses `./data`, `./banner.txt`, `./saves`, `./logs` and the gods JSON file from there, as it does today.
- **Config.** The server reads `./.env` if it exists. Real env vars override values from it.

### Story: Craft Your Own World
As an author, I want to be able to edit the world using the provided editor. I don't want to compile the editor myself. I am willing to get my hands dirty with technical content. I expect to 
need to edit Lua files manually. I expect to manually use version control and handle merge conflicts between my changes and those of my collaborators. 

Requirements:
- **The editor is in the server binary.** It's enabled with the `--edit` flag, served at `/editor`, and only available to gods. The plan is in [mudkit_editor_migration.md](mudkit_editor_migration.md).
- **Reload.** The editor's Reload button reloads the world from the saved files, and `/reload` reloads it after hand edits. See the reload decision below.

### Story: Username and password accounts
As a player, I can create an account from the web client with a username and password, and log in with it. Usernames are unique. Passwords are stored only as hashes. There's no OAuth and no email verification.

As a hoster, I name the first gods with the `GOD_USERS` env variable (a comma-separated list of usernames). God permissions are saved in a JSON file, and a god can make another user a god with `/deify <username>`. `/deify` takes an account username, not an in-game character name.

Requirements:
- **Signup rules.** A username is 3–20 characters of letters, digits and underscores, and is unique ignoring case. A password is at least 8 characters. Passwords are hashed with bcrypt. There's no rate limiting or lockout.
- **The username is the account's identity.** Usernames can't be changed, so character saves (in place of today's integer `PlayerId`), sessions and the gods file all store the username.
- **Gods.** A user is a god if they're listed in `GOD_USERS` or in the gods JSON file. On startup the server writes every `GOD_USERS` entry into the JSON file, so the file ends up listing every god. `/deify` adds a username to the file and fails if no account has that username. There's no in-game command to remove a god. The hoster removes the user from `GOD_USERS` and from the JSON file, then restarts the server.
- **Removed.** Recurse Center OAuth (the `RC_*` env vars) and debug auth (`ENABLE_DEBUG_AUTH`) are removed from this repo. Developers test with ordinary accounts. The RC Disco maintainers keep RC OAuth in their own fork.

### Story: Rooms as Lua
As an author, I want the world to be made of Lua files that I can edit by hand or with the editor, so that my whole world lives in my repo as readable text.

The Lua files are the world. The editor is a helper tool that makes editing those Lua files easier, so anything it saves is Lua, and anything an author writes by hand still works when the editor opens the file.

In release 1, the editor edits **rooms only**: the world map, room fields, exits, locks, and the NPCs and items placed in a room. Every other kind of content (items, spells, classes, races, NPC definitions, conditions, quests) is edited by hand in Lua. Editing it in the editor is in [Later](#later).

As a non-functional outcome of this story, the world JSON goes away. Each room is a Lua script, and rooms no longer point to each other with room indexes for exits. Instead, rooms are marked with X,Y,Z coordinates and list their exits by direction (see Decisions). Player positions are X,Y,Z coordinates too. Navigation uses the directions north, east, south, west, up, and down. When content loads, the server validates that no two rooms share the same coordinates.

Additionally, the flat NPC array (`World.Npcs`) goes away. NPCs live inside the room that spawns them and are updated during the room's update. Each room script lists the NPCs it spawns, by NPC id along with their behavior params. The spawn room owns its NPCs: it updates each one every tick wherever the NPC's mob currently is (a wanderer may be several rooms away), and respawns it in the spawn room when it dies.

Requirements:
- **One file per room.** Room scripts live under `data/rooms/`, one room per file, in whatever subfolders the author likes (e.g. `data/rooms/forest/`). Like other content, each room has a stable string `id` in its script. When the editor creates a room, the author picks the folder.
- **Positions are coordinates.** Character saves store the player's position as X,Y,Z coordinates.
- **The start room is at 0,0,0.** New characters start there, and players whose room no longer exists after a reload are moved there. Validation fails if there's no room at 0,0,0.

### Story: Thorough God command suite
As a god, I can use god commands to manage the world. The purpose of these commands is playtesting and content management during authoring. 
- I can spawn monsters, 
- I can kill monsters and players, 
- I can spawn items
- I can de-spawn items
- I can delete all items of a specific Id from the world
- I can teleport to specific rooms
- I can set my level, race, and class so as to playtest in certain situations

## Later
Stories that are agreed but not in release 1. The agent does not build these.

### Story: Edit all content in the editor
As an author, I want to edit classes, races, NPC definitions, items, spells, conditions and quests in the editor, not only rooms. Like rooms, the editor edits the data fields of these Lua files in place and leaves their hook functions to the author.

## Out of scope
What this PRD deliberately doesn't cover, so the agent doesn't add it.
- Changing the core game rules (combat, stats, leveling) from content scripts, and non-fantasy genres that would need different rules.

## Decisions
- 2026-11-02: Content can be deleted, but only when nothing references it. Every content script (item, spell, class, race, NPC, room) gets a stable string `id` that is separate from its display name, and all references, including player saves, use that `id`. We chose this over Eclipse-style fixed slots because content lives in Lua files whose identity is the file/id, not a numbered slot; because repurposing a slot silently changes what players already own; and because stable IDs make renames safe (today, saves reference content by display name, so a rename breaks them just like a delete).
- 2026-11-02: Player save files count as references. Deleting content scans `saves/` along with live content and online players, and the delete is blocked while any character references the content. Gods clean up references with god commands before deleting. We chose this because it is strict and predictable, and it never silently changes a character.
- 2026-11-02: The engine and the content are split. This repo ships prebuilt per-platform binaries for the server (with the web client embedded) and the editor. `backend/data` is RC Disco's game content and isn't compiled into a release. Each author keeps their own data folder in their own repo and points the binaries at it. Authors get engine updates by downloading a new release binary.
- 2026-11-02: The Lua API isn't a versioned contract. Mudkit is mainly about wrapping up this project neatly, not about supporting a platform with many users, so release 1 favors the lightest option that meets each story.
- 2026-11-02: The editor moves into the game server, so a release ships one binary. The editor is only enabled when the server runs in editor mode on the author's machine, never on a deployed server, because it writes files in the author's git checkout. We chose this over a separate editor binary to drop the second Go module, the second frontend build and the Chromium window. See [mudkit_editor_migration.md](mudkit_editor_migration.md).
- 2026-11-02: Content changes reach a running server through a whole-world reload, never by patching the live world. A reload is only run on request (after an editor save, or the `/reload` god command), builds and validates a new world next to the old one, and swaps it in only if validation passes. Otherwise the old world keeps running and the errors are reported. Online players are saved and spawned again at the same coordinates. Transient state (NPC instances, room items, combat) is lost. We chose this over patching the live world because patching needs reconciliation logic for every content type and can leave the server half-updated, while a reload reuses the startup loading and validation code.
- 2026-11-02: A room's X,Y,Z coordinates give its position, and the room still lists its exits by direction (north, east, south, west, up, down). An exit always leads to the neighbouring coordinate in its direction. Validation requires that the neighbour exists and has the exit back, so exits stay two-way. We chose this over "adjacent rooms are always connected" so that neighbouring rooms can be separated by a wall, and existing areas play the same.
- 2026-11-02: The Lua files are the source of truth, and the editor edits them in place. It parses a content file, finds the top-level data assignments (`room.field = <value>`), replaces only those text ranges, and leaves hook functions, comments and formatting untouched. A field the editor can't represent (for example, a value computed by a function call) is shown read-only, not overwritten. We chose this over splitting each script into an editor-owned data file and a hand-written hooks file, and over regenerating whole files, because authors should be able to treat the editor as a convenience on top of files they own.
- 2026-11-02: In release 1, the editor only edits rooms. Editing the other kinds of content is in Later. This keeps the scope small while the project is being wrapped up.
- 2026-11-02: Username and password is the only login method. We dropped configurable OAuth providers to keep the scope small. RC OAuth lives in the RC Disco maintainers' fork instead.
- 2026-11-02: Releases ship a minimal starter world, not RC Disco's content. RC Disco's world is too big to delete content from, especially since content can only be deleted when nothing references it.
- 2026-11-02: The game rules (combat, stats, ticks, leveling) stay in Go, and Mudkit has no LPMud-style mudlib tier. Creators build content inside a fixed fantasy ruleset; classes and races are parameters to that ruleswhy am i getting so much election spamet, not new rules. We chose this because the target creator is a storyteller, not a game-system designer, and moving the rules into Lua would mean rewriting the engine and losing type safety for flexibility the target users don't need.

## Open questions
- [NEEDS CLARIFICATION] <question the agent must ask about, not guess>
