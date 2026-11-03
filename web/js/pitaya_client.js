import {
  MESSAGE_PUSH,
  MESSAGE_RESPONSE,
  PACKET_DATA,
  PACKET_HANDSHAKE,
  PACKET_HANDSHAKE_ACK,
  PACKET_HEARTBEAT,
  decodeJson,
  decodeMessage,
  decodePacket,
  encodeHandshake,
  encodeMessage,
  encodeNotify,
  encodePacket,
} from "./pitaya_protocol.js";

export class PitayaClient {
  constructor(url) {
    this.url = url;
    this.ws = null;
    this.requestId = 0;
    this.pendingRequests = new Map();
    this.routeHandlers = new Map();
  }

  connect() {
    return new Promise((resolve, reject) => {
      const ws = new WebSocket(this.url);
      ws.binaryType = "arraybuffer";
      let handshaken = false;

      ws.onopen = () => ws.send(encodeHandshake());
      ws.onerror = () => {
        if (!handshaken) reject(new Error(`Cannot connect to ${this.url}`));
      };
      ws.onclose = () => {
        if (this.ws === ws) this.ws = null;
        this._failPendingRequests(new Error("Connection closed"));
        if (!handshaken) reject(new Error(`Cannot connect to ${this.url}`));
      };
      ws.onmessage = async (event) => {
        try {
          const { packetType, payload } = this._decodeEvent(event);
          if (!handshaken) {
            if (packetType !== PACKET_HANDSHAKE) {
              throw new Error(`Unexpected Pitaya handshake packet type: ${packetType}`);
            }
            decodeJson(payload);
            ws.send(encodePacket(PACKET_HANDSHAKE_ACK));
            handshaken = true;
            this.ws = ws;
            resolve();
            return;
          }
          await this._handlePacket(ws, packetType, payload);
        } catch (error) {
          if (!handshaken) {
            ws.close();
            reject(error);
          } else {
            this._failPendingRequests(error);
          }
        }
      };
    });
  }

  registerRoute(route, callback) {
    this.routeHandlers.set(route, callback);
  }

  unregisterRoute(route) {
    this.routeHandlers.delete(route);
  }

  close() {
    if (this.ws) {
      const ws = this.ws;
      this.ws = null;
      ws.close();
    }
    this._failPendingRequests(new Error("Connection closed"));
  }

  login(accountId, password) {
    return this.request("AccountHandler.Login", { accountId, password });
  }

  createAccount(accountId, password, nickname) {
    return this.request("AccountHandler.CreateNewAccount", { accountId, password, nickname });
  }

  summon(poolId, times) {
    return this.request("SummonHandler.SummonCard", { poolId, times });
  }

  getInfo() {
    return this.request("PlayerHandler.GetInfo", {});
  }

  getBackpack() {
    return this.request("BackpackHandler.GetBackpack", {});
  }

  getCards() {
    return this.request("CardHandler.GetCards", {});
  }

  ascendCard(cardId) {
    return this.request("CardHandler.AscendCard", { cardId });
  }

  starUpCard(cardId) {
    return this.request("CardHandler.StarUpCard", { cardId });
  }

  upgradeCard(cardId, levels = 1) {
    return this.request("CardHandler.UpgradeCard", { cardId, levels });
  }

  publish(route, body) {
    if (!this.ws) throw new Error("WebSocket connection is not established.");
    this.ws.send(encodeNotify(route, body));
  }

  request(route, body) {
    if (!this.ws) return Promise.reject(new Error("WebSocket connection is not established."));

    const requestId = ++this.requestId;
    return new Promise((resolve, reject) => {
      this.pendingRequests.set(requestId, { resolve, reject });
      try {
        this.ws.send(encodeMessage(requestId, route, body));
      } catch (error) {
        this.pendingRequests.delete(requestId);
        reject(error);
      }
    });
  }

  _decodeEvent(event) {
    if (typeof event.data === "string") {
      throw new Error("Expected a binary Pitaya packet");
    }
    return decodePacket(new Uint8Array(event.data));
  }

  async _handlePacket(ws, packetType, payload) {
    if (packetType === PACKET_HEARTBEAT) {
      ws.send(encodePacket(PACKET_HEARTBEAT));
      return;
    }
    if (packetType !== PACKET_DATA) return;

    const message = await decodeMessage(payload);
    if (message.messageType === MESSAGE_RESPONSE) {
      const pending = this.pendingRequests.get(message.requestId);
      if (pending) {
        this.pendingRequests.delete(message.requestId);
        pending.resolve(decodeJson(message.body));
      }
    } else if (message.messageType === MESSAGE_PUSH) {
      const callback = this.routeHandlers.get(message.route);
      if (callback) callback(decodeJson(message.body));
    }
  }

  _failPendingRequests(error) {
    const pending = [...this.pendingRequests.values()];
    this.pendingRequests.clear();
    for (const { reject } of pending) reject(error);
  }
}
