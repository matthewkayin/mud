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
  game/text_color.ts    terminal text colors
  game/status_bar.tsx   HP and MP meters
  game/minimap.tsx      SVG map of the rooms around the player
  game/inventory.tsx    inventory panel (placeholder until there is inventory data)
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

The game page renders its panels inside an `article.game` in two columns: the terminal with the status bar below it, and a sidebar to its right holding the minimap and inventory. The panels are always on the page. The status bar and inventory are empty until the player is in the world (the server is sending a status), and the minimap is empty until the server sends one. The minimap draws each discovered room as a box with lines for its north, east, south and west exits, and carets in the gap east of the room for up and down exits: a `^` near the top of its east side for up and a `v` near the bottom for down.

## Auth

`AuthProvider` asks `GET /api/auth/me` once on load and exposes `{ user, loading, signIn, createAccount, signOut }` through `useAuth()`. Pages that depend on login state render nothing while `loading` is true, so they don't redirect before the check finishes. Errors thrown by `auth_api.ts` carry a message meant to be shown to the user.

The server must implement these routes. Sessions are carried by a cookie the server sets; the client never handles tokens.

| Route                     | Body                     | Success                              | Failure            |
|---------------------------|--------------------------|--------------------------------------|--------------------|
| `GET /api/auth/me`        |                          | `200 { "username": string }`         | `401`              |
| `POST /api/auth/register` | `{ username, password }` | `200 { "username": string }` + cookie | `4xx { "error": string }` |
| `POST /api/auth/login`    | `{ username, password }` | `200 { "username": string }` + cookie | `4xx { "error": string }` |
| `POST /api/auth/logout`   |                          | `204`, clears the cookie             | `4xx { "error": string }` |

The `error` string is shown to the user as is. Register returns `409` when the username is taken.

Usernames are 3–20 characters of ASCII letters, digits and `_`, and are unique regardless of capitalization. Passwords are 8–72 bytes. The server returns the username with the capitalization used at registration, even when the user signs in with different capitalization. Server-side details are in `docs/backend/accounts.md`.

## Styling

All styling lives in `src/styles.css`. The look is a parchment page with serif type, small-caps headings, double-ruled borders, underlined oxblood links and plain bevelled buttons, closer to an early-2000s personal site than a modern app. Keep it that way:

- Use the color variables on `:root` rather than new literal colors.
- Prefer plain HTML elements (`fieldset`, `hr`, links) styled globally over new wrapper components.
- Don't add a CSS framework or component library.
- The `.page` frame sizes to its content, is never narrower than 760px (except on small screens) and never wider than the viewport. Each page's content sets the width:
  - Text pages render a top-level `<article>`, which the stylesheet fixes at 690px, the width of the old 760px column inside the frame.
  - The game page's `article.game` is a two-column grid with one row. The `.game-main` column holds the terminal (100 columns wide, which sets the column's width) with the status bar below it, and the `.game-sidebar` takes the fixed-width second column, so together they set the frame's width. The sidebar stretches to the height of the first column: the minimap sits at its top and the inventory grows to fill the rest, lining its bottom edge up with the status bar's. The minimap has a fixed height, which sets the size of its rooms; the map is centered in the panel's width. On small screens the grid becomes one column.
- No page may scroll horizontally at phone width; the `@media (max-width: 600px)` block handles small screens.
