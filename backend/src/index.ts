import { createServer, IncomingMessage } from 'http'; // TODO: https
import { WebSocketServer, WebSocket } from 'ws';
import express from 'express';
import cors from 'cors';
import { handlePostAuthToken } from './auth';
import * as net from 'net';

const PORT = 3000;
const MUD_PORT = 7272;

const app = express();

app.use(cors({
  origin: 'http://localhost:5173'
}));
app.use(express.json());

app.post('/api/auth/token', handlePostAuthToken);
const server = createServer(app);

const webSocketServer = new WebSocketServer({ noServer: true });

// Listen for raw HTTP 'upgrade' event
server.on('upgrade', async (request, socket, head) => {
  const pathname = new URL(request.url || '', `http://${request.headers.host}`).pathname;
  if (pathname !== '/ws-telnet') {
    socket.destroy();
    return;
  }

  const token = request.headers['sec-websocket-protocol'];
  if (!token) {
    socket.write('HTTP/1.1 401 Unauthorized\r\n\r\n');
    socket.destroy();
    return;
  }

  try {
    const response = await fetch(`${process.env.RC_API_URL}/profiles/me`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`
      }
    });

    if (!response.ok) {
      console.error(`Recurse API responded with status ${response.status}`);
      const responseText = await response.text();
      console.error(responseText);
      socket.write('HTTP/1.1 500 Internal Server Error\r\n\r\n');
      socket.destroy();
      return;
    }

    const profileData = await response.json();
    console.log(`Got profile data from RC API for user ${profileData.id}`);

    webSocketServer.handleUpgrade(request, socket, head, (webSocket) => {
      webSocketServer.emit('connection', webSocket, request, profileData.id);
    });
  } catch (err) {
    if (err instanceof Error) {
      console.error('OAuth token validation failed:', err.message);
    }
    socket.write('HTTP/1.1 401 Unauthorized\r\n\r\n');
    socket.destroy();
  }
});

webSocketServer.on('connection', (webSocket: WebSocket, request: IncomingMessage, userId: number) => {
  console.log(`Setting up Telnet bridge for user ID ${userId}`);

  const telnetSocket = new net.Socket();
  telnetSocket.connect(MUD_PORT, '127.0.0.1');

  telnetSocket.on('data', (data) => webSocket.send(data.toString()));
  webSocket.on('message', (message) => telnetSocket.write(message.toString()));
  telnetSocket.on('close', () => webSocket.close());
  telnetSocket.on('error', () => webSocket.close());
  webSocket.on('error', () => telnetSocket.end());
});

server.listen(PORT, () => {
  console.log(`RC Disco MUD backend is listening on port ${PORT}...`);
});
