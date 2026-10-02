# Castle Recurse

![Castle Recurse Welcome](./readme/welcome.png)

Castle Recurse is a [Multi-User Dungeon](https://en.wikipedia.org/wiki/Multi-user_dungeon) (a multiplayer online text adventure game) developed for the Recurse Center. The game is still in active development.

## Dependencies
 - Go
 - NodeJS
 - NPM

 ## Running the Project
 
1. Clone the repo
2. Put the `env.json` file into the root of the `backend/` folder
3. Run the backend in one terminal
```
cd backend
go run .
```

4. In another terminal, install frontend dependencies and run the frontend
```
cd frontend
npm install

npm run dev
```

5. Access the frontend by navigating to `localhost:5173` in your browser.

## Project Architecture

The game runs on a Go backend server and players connect using a companion React frontend. The frontend contains no game logic and is just a plain terminal which sends and receives messages.

Users login from the frontend via Recurse OAuth. Once logged in, their client establishes a websocket connection with the server, and the rest of gameplay happens through this websocket connection.

Since this is a realtime multiplayer game, good concurrency is an important part of keeping the game running. Pictured below is a flow of the game's concurrency scheme.

![Concurrency Diagram](./readme/concurrency.png)

- The main goroutine handles gameplay
  - Gameplay happens on a 3 second server tick
  - All game logic is single-threaded 
- A separate goroutine runs an HTTP server and listens for requests
  - Each new HTTP request spawns a separate goroutine to handle the request (this is the default behavior for http in Go)
- When a websocket connection is made between client and server, two goroutines are established to run the connection
  - The WebSocket Reader routine listens for messages from the client and sends them to the game
  - The WebSocket Writer routine listens for messages from the game and sends them to the client
  - Go's channels feature is used to facilitate this communication. Messages can be sent between the game and the client's web socket connection without explicit synchronization on the application side
