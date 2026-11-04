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
  game/game.tsx         game page: websocket connection and terminal state
  game/terminal/        terminal component, color-tag parser and colors
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

The game page renders the terminal inside an `article.game`.

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
  - The game page's `article.game` is a flex row. Its panels (the terminal, which is 100 columns wide, and later the minimap and status bars) set its width, so a new panel widens the frame.
- No page may scroll horizontally at phone width; the `@media (max-width: 600px)` block handles small screens.
