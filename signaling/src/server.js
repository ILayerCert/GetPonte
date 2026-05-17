/**
 * Ponte POC — Lightweight WebRTC Signaling Server
 *
 * This server handles WebRTC signaling for peer-to-peer mesh connections.
 * No SFU (mediasoup) — peers connect directly to each other.
 * Works great for small rooms (up to ~6 participants).
 *
 * Protocol:
 *   Client → Server:
 *     { type: "join", roomId, peerId, username }
 *     { type: "signal", targetPeerId, signal }
 *     { type: "leave" }
 *
 *   Server → Client:
 *     { type: "room-peers", peers: [{ peerId, username }] }
 *     { type: "peer-joined", peerId, username }
 *     { type: "signal", fromPeerId, signal }
 *     { type: "peer-left", peerId }
 */

const { WebSocketServer } = require('ws');
const { v4: uuidv4 } = require('uuid');
const http = require('http');

const PORT = parseInt(process.env.PORT) || 4000;

// ── Room & Peer tracking ───────────────────────────────────────────────

// rooms: Map<roomId, Map<peerId, { ws, username }>>
const rooms = new Map();

// ws → { roomId, peerId, username }
const peerInfo = new Map();

// ── HTTP server (health check) ─────────────────────────────────────────

const httpServer = http.createServer((req, res) => {
  if (req.url === '/health') {
    const roomCount = rooms.size;
    let peerCount = 0;
    rooms.forEach(r => { peerCount += r.size; });
    res.writeHead(200, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify({ status: 'ok', rooms: roomCount, peers: peerCount }));
  } else {
    res.writeHead(404);
    res.end();
  }
});

// ── WebSocket server ───────────────────────────────────────────────────

const wss = new WebSocketServer({ server: httpServer });

wss.on('connection', (ws) => {
  console.log('[connect] New WebSocket connection');

  ws.on('message', (raw) => {
    let msg;
    try {
      msg = JSON.parse(raw);
    } catch (e) {
      console.error('[error] Invalid JSON:', raw.toString().slice(0, 100));
      return;
    }

    switch (msg.type) {
      case 'join':
        handleJoin(ws, msg);
        break;
      case 'signal':
        handleSignal(ws, msg);
        break;
      case 'leave':
        handleLeave(ws);
        break;
      default:
        console.warn('[warn] Unknown message type:', msg.type);
    }
  });

  ws.on('close', () => {
    handleLeave(ws);
  });

  ws.on('error', (err) => {
    console.error('[error] WebSocket error:', err.message);
    handleLeave(ws);
  });
});

// ── Handlers ───────────────────────────────────────────────────────────

function handleJoin(ws, msg) {
  const { roomId, peerId, username } = msg;

  if (!roomId || !peerId) {
    ws.send(JSON.stringify({ type: 'error', message: 'roomId and peerId required' }));
    return;
  }

  // Leave any existing room first
  handleLeave(ws);

  // Get or create room
  if (!rooms.has(roomId)) {
    rooms.set(roomId, new Map());
    console.log(`[room] Created room ${roomId}`);
  }

  const room = rooms.get(roomId);

  // Store peer info
  const info = { roomId, peerId, username: username || 'Anonymous' };
  peerInfo.set(ws, info);
  room.set(peerId, { ws, username: info.username });

  console.log(`[join] ${info.username} (${peerId}) joined room ${roomId} (${room.size} peers)`);

  // Send existing peers to the new peer
  const existingPeers = [];
  room.forEach((peer, id) => {
    if (id !== peerId) {
      existingPeers.push({ peerId: id, username: peer.username });
    }
  });

  ws.send(JSON.stringify({
    type: 'room-peers',
    peers: existingPeers,
  }));

  // Notify existing peers about the new peer
  broadcast(roomId, peerId, {
    type: 'peer-joined',
    peerId,
    username: info.username,
  });
}

function handleSignal(ws, msg) {
  const info = peerInfo.get(ws);
  if (!info) return;

  const { targetPeerId, signal } = msg;
  if (!targetPeerId || !signal) return;

  const room = rooms.get(info.roomId);
  if (!room) return;

  const target = room.get(targetPeerId);
  if (!target) return;

  // Forward the signal to the target peer
  target.ws.send(JSON.stringify({
    type: 'signal',
    fromPeerId: info.peerId,
    signal,
  }));
}

function handleLeave(ws) {
  const info = peerInfo.get(ws);
  if (!info) return;

  const { roomId, peerId, username } = info;
  peerInfo.delete(ws);

  const room = rooms.get(roomId);
  if (!room) return;

  room.delete(peerId);
  console.log(`[leave] ${username} (${peerId}) left room ${roomId} (${room.size} peers remaining)`);

  // Notify remaining peers
  broadcast(roomId, peerId, {
    type: 'peer-left',
    peerId,
  });

  // Cleanup empty rooms
  if (room.size === 0) {
    rooms.delete(roomId);
    console.log(`[room] Deleted empty room ${roomId}`);
  }
}

// ── Utilities ──────────────────────────────────────────────────────────

function broadcast(roomId, excludePeerId, message) {
  const room = rooms.get(roomId);
  if (!room) return;

  const data = JSON.stringify(message);
  room.forEach((peer, id) => {
    if (id !== excludePeerId && peer.ws.readyState === 1) {
      peer.ws.send(data);
    }
  });
}

// ── Start ──────────────────────────────────────────────────────────────

httpServer.listen(PORT, '0.0.0.0', () => {
  console.log(`🌉 Ponte Signaling Server running on port ${PORT}`);
  console.log(`   Health check: http://localhost:${PORT}/health`);
});
