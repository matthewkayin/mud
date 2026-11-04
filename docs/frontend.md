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
  pages/                one component per route
```

File names are snake_case; components are PascalCase named exports.

## Routes

| Path              | Page                                   | Notes                                              |
|-------------------|----------------------------------------|----------------------------------------------------|
| `/`               | `home_page.tsx`                        | What the game is, link to play                     |
| `/signin`         | `sign_in_page.tsx`                     | Redirects to `/game` when already signed in        |
| `/create-account` | `create_account_page.tsx`              | Redirects to `/game` when already signed in        |
| `/game`           | `game_page.tsx`                        | Redirects to `/signin` when signed out             |
| `*`               | `not_found_page.tsx`                   |                                                    |

The game page renders an empty `.terminal-slot` div. The terminal component goes inside it.

## Auth

`AuthProvider` asks `GET /api/auth/me` once on load and exposes `{ user, loading, signIn, createAccount, signOut }` through `useAuth()`. Pages that depend on login state render nothing while `loading` is true, so they don't redirect before the check finishes. Errors thrown by `auth_api.ts` carry a message meant to be shown to the user.

The server must implement these routes. Sessions are carried by a cookie the server sets; the client never handles tokens.

| Route                     | Body                     | Success                              | Failure            |
|---------------------------|--------------------------|--------------------------------------|--------------------|
| `GET /api/auth/me`        |                          | `200 { "username": string }`         | `401`              |
| `POST /api/auth/register` | `{ username, password }` | `200 { "username": string }` + cookie | `4xx { "error": string }` |
| `POST /api/auth/login`    | `{ username, password }` | `200 { "username": string }` + cookie | `4xx { "error": string }` |
| `POST /api/auth/logout`   |                          | `204`, clears the cookie             | `4xx { "error": string }` |

The `error` string is shown to the user as is.

## Styling

All styling lives in `src/styles.css`. The look is a parchment page with serif type, small-caps headings, double-ruled borders, underlined oxblood links and plain bevelled buttons, closer to an early-2000s personal site than a modern app. Keep it that way:

- Use the color variables on `:root` rather than new literal colors.
- Prefer plain HTML elements (`fieldset`, `hr`, links) styled globally over new wrapper components.
- Don't add a CSS framework or component library.
- The column is at most 760px wide and must not scroll horizontally at phone width; the `@media (max-width: 600px)` block handles small screens.
