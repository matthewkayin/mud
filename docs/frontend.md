# Frontend

The web client is a small React + Vite + TypeScript site in `frontend/`. It has no component library and few dependencies: `react`, `react-dom` and `react-router-dom` at runtime, plus Vite, TypeScript and ESLint for development.

## Running

```
cd frontend
npm install
npm run dev     # Vite on :5173
npm run build   # tsc -b && vite build, writes frontend/dist
npm run lint
```

The dev server proxies `/api` (and the websocket at `/api/websocket`) to the game server at `http://localhost:7272`. Set `MUD_BACKEND_URL` to point it elsewhere. The client only uses same-origin URLs, so it works behind the proxy and when served by the game server.

## Layout

```
src/
  main.tsx              StrictMode > BrowserRouter > AuthProvider > App
  app.tsx               Layout + route table
  styles.css            the whole site's styling
  auth/auth_api.ts      fetch wrappers for /api/auth/*
  auth/auth_context.tsx AuthProvider and useAuth()
  components/layout.tsx header, nav, main, footer
  components/nav.tsx    site links and the signed-in / signed-out links
  pages/                one component per text page
  game/game.tsx         game page: websocket connection, terminal, status and minimap state
  game/socket_message.ts shape of the messages the server sends over the websocket
  game/terminal.tsx     terminal component
  game/parser.ts        color-tag parser for terminal lines
  game/text_line.tsx    renders one color-tagged line, shared by the terminal and the view panel
  game/text_color.ts    terminal text colors
  game/status_bar.tsx   HP and MP meters
  game/minimap.tsx      SVG map of the rooms around the player
  game/view.tsx         view panel (help, chat, inventory or equipment)
  game/inventory_table.ts formats the server's inventory into the view panel's table
  game/equipment_table.ts formats the server's equipment into the view panel's table
  game/view_format.ts   the view panel's line width, header format and table helpers
  game/view_mode.ts     view panel modes
```

File names are snake_case; components are PascalCase named exports.

## Routes

| Path              | Page                                   | Notes                                              |
|-------------------|----------------------------------------|----------------------------------------------------|
| `/`               | `pages/home.tsx`                       | What the game is, link to play                     |
| `/signin`         | `pages/signin.tsx`                     | Redirects to `/game` when already signed in        |
| `/create-account` | `pages/create.tsx`                     | Redirects to `/game` when already signed in        |
| `/game`           | `game/game.tsx`                        | Redirects to `/signin` when signed out             |
| `*`               | `pages/not_found.tsx`                  |                                                    |

The game page renders its panels inside an `article.game` in two columns: the terminal with the status bar below it, and a sidebar to its right holding the minimap and the view panel. The panels are always on the page. The status bar and view panel are empty until the player is in the world (the server is sending a status), and the minimap is empty until the server sends one. The server sends a 17×17 grid centered on the player, and the minimap shows all 17 columns of its middle 9 rows, so the panel is wider than it is tall. It cuts off whole rows, so the exit lines of rooms on every edge still run to the panel's border, showing that the map continues. The minimap draws each discovered room as a box with lines for its north, east, south and west exits, and carets in the gap east of the room for up and down exits: a `^` near the top of its east side for up and a `v` near the bottom for down. While the status's `action` is true (the player has queued an action that hasn't resolved yet), the terminal prompt is replaced by a spinner.

## View panel

The view panel (`game/view.tsx`) shows text only, one mode at a time, under a header: a 70-column (`VIEW_WIDTH`) `+---- <Mode> ----+` line with the mode's name centered, built by `formatViewHeader` in `game/view_format.ts`. Headers never wrap; when the panel is narrower than them, the panel scrolls sideways. It shows nothing, in any mode, while the player isn't in the world:

- `help`: explains the panel and the `view` command. The default.
- `chat`: the chat history.
- `inventory`: a bordered table that `game/inventory_table.ts` builds from the server's `inv`, with `Item`, `Amount` and `Size (<used> / <capacity>)` columns and the table's top border as its header. The Item column is always 38 characters wide, and longer names are cut off with `...`. The Amount column fits `Amount`, and the Size column fits at least `Size (100 / 100)`, so the table is 70 columns wide for a capacity under 1000. An empty inventory shows one `(empty)` row. Table rows never wrap: when the panel is narrower than the table, the panel scrolls sideways.
- `equipment`: a bordered table that `game/equipment_table.ts` builds from the server's `equip`, with `Slot` and `Item` columns and the table's top border as its header. Each column is only as wide as its longest entry, so the table is usually narrower than the panel. Item names longer than the inventory's 38-character item width are cut off with `...`. An empty slot shows `---`. Item cells, including `---` and the `Item` heading, are centered.

The client handles `view <mode>` itself in `game.tsx` and never sends it to the server: it switches the mode, or writes an error to the terminal when the player isn't in the world, or a usage line when the mode is missing or unknown. The server has a `view` help topic so `help view` still describes it. The client already holds the lines for every mode:

- Any message that starts with a bracketed tag, optionally inside a color tag (such as `<p>[Chat]</p> ...` or `[Whisper] ...`), is also appended to the chat history, which keeps the last 1024 lines. These messages still appear in the terminal. Other server text must not start with a bracketed tag, or it will be taken for chat.
- The server sends `inv` (an object with the inventory's `size`, its `capacity` and its `items`, each with a `name`, `amount` and `size`, where an item's size is for the whole stack) and `equip` (an array with one object per slot, each with the slot's name in `slot` and the equipped item's `name`, which is empty for an empty slot; a two-handed weapon appears in both hands) in a heartbeat whenever the player's inventory or equipment changes, including on entering the world. They are cleared when a status-less heartbeat arrives.

While the panel is scrolled to the bottom it follows new lines, like the terminal, and switching modes scrolls to the bottom.

## Auth

`AuthProvider` asks `GET /api/auth/config` and then `GET /api/auth/me` once on load, and exposes `{ user, loading, debugAuth, signIn, createAccount, debugSignIn, signOut }` through `useAuth()`. Pages that depend on login state render nothing while `loading` is true, so they don't redirect before the check finishes. Errors thrown by `auth_api.ts` carry a message meant to be shown to the user.

The server must implement these routes. Sessions are carried by a cookie the server sets; the client never handles tokens.

| Route                     | Body                     | Success                              | Failure            |
|---------------------------|--------------------------|--------------------------------------|--------------------|
| `GET /api/auth/me`        |                          | `200 { "username": string }`         | `401`              |
| `POST /api/auth/register` | `{ username, password }` | `200 { "username": string }` + cookie | `4xx { "error": string }` |
| `POST /api/auth/login`    | `{ username, password }` | `200 { "username": string }` + cookie | `4xx { "error": string }` |
| `POST /api/auth/logout`   |                          | `204`, clears the cookie             | `4xx { "error": string }` |
| `GET /api/auth/config`    |                          | `200 { "debugAuth": boolean }`       |                    |
| `POST /api/auth/debug-login` | `{ username }`        | `200 { "username": string }`, no cookie | `4xx { "error": string }` |

The `error` string is shown to the user as is. Register returns `409` when the username is taken.

### Debug auth

When the server has debug auth enabled (see `docs/backend/accounts.md`), `debugAuth` is true and the sign-in page asks only for a username, calling `debugSignIn`. The returned username is kept in `sessionStorage` under `mud_debug_username`, which is per tab, so each tab can be a different player. On load, a stored debug username is used instead of `/api/auth/me`. The game page adds `?debugUser=<name>` to the websocket URL when one is stored. Signing out of a debug sign-in clears it from `sessionStorage` without calling `/api/auth/logout`, which would clear the cookie other tabs share.

Usernames are 3–20 characters of ASCII letters, digits and `_`, and are unique regardless of capitalization. Passwords are 8–72 bytes. The server returns the username with the capitalization used at registration, even when the user signs in with different capitalization. Server-side details are in `docs/backend/accounts.md`.

## Styling

All styling lives in `src/styles.css`. The look is a parchment page with serif type, small-caps headings, double-ruled borders, underlined oxblood links and plain bevelled buttons, closer to an early-2000s personal site than a modern app. Keep it that way:

- Use the color variables on `:root` rather than new literal colors.
- Prefer plain HTML elements (`fieldset`, `hr`, links) styled globally over new wrapper components.
- Don't add a CSS framework or component library.
- The `.page` frame sizes to its content, is never narrower than 760px (except on small screens) and never wider than the viewport. Each page's content sets the width:
  - Text pages render a top-level `<article>`, which the stylesheet fixes at 690px, the width of the old 760px column inside the frame.
  - The game page's `article.game` is a two-column grid with one row. The `.game-main` column holds the terminal (128 columns wide plus room for its scrollbar, which sets the column's width) with the status bar below it, and the `.game-sidebar` takes the fixed-width second column, so together they set the frame's width. The sidebar stretches to the height of the first column: the minimap sits at its top and the view panel grows to fill the rest, lining its bottom edge up with the status bar's. The minimap is as wide as the sidebar with a 17:9 aspect ratio (the map's columns by its visible rows), and the map scales to fit it; the view panel takes the rest of the column, so a taller terminal makes only the view panel taller. The view panel's content never affects the sidebar's height: it scrolls whatever doesn't fit. On small screens the grid becomes one column, and the view panel is a fixed 20em tall.
- No page may scroll horizontally at phone width; the `@media (max-width: 600px)` block handles small screens.
