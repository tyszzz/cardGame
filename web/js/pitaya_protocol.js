export const PACKET_HANDSHAKE = 1;
export const PACKET_HANDSHAKE_ACK = 2;
export const PACKET_HEARTBEAT = 3;
export const PACKET_DATA = 4;
export const PACKET_KICK = 5;

export const MESSAGE_REQUEST = 0;
export const MESSAGE_NOTIFY = 1;
export const MESSAGE_RESPONSE = 2;
export const MESSAGE_PUSH = 3;
const MESSAGE_TYPE_MASK = 0x07;
const MESSAGE_GZIP_MASK = 0x10;

const textEncoder = new TextEncoder();
const textDecoder = new TextDecoder();

export function encodePacket(packetType, payload = new Uint8Array(0)) {
  const length = payload.length;
  if (length > 0xffffff) {
    throw new Error("Pitaya packet payload is too large");
  }
  const packet = new Uint8Array(4 + length);
  packet[0] = packetType;
  packet[1] = (length >> 16) & 0xff;
  packet[2] = (length >> 8) & 0xff;
  packet[3] = length & 0xff;
  packet.set(payload, 4);
  return packet;
}

export function decodePacket(data) {
  if (data.length < 4) {
    throw new Error("Invalid Pitaya packet header");
  }
  const payloadLength = (data[1] << 16) | (data[2] << 8) | data[3];
  const payload = data.subarray(4);
  if (payload.length !== payloadLength) {
    throw new Error("Incomplete Pitaya packet");
  }
  return { packetType: data[0], payload };
}

export function encodeHandshake() {
  const payload = {
    sys: { type: "js-websocket", version: "1.0.0", heartbeat: 0 },
    user: {},
  };
  return encodePacket(PACKET_HANDSHAKE, textEncoder.encode(JSON.stringify(payload)));
}

function concat(parts) {
  const result = new Uint8Array(parts.reduce((sum, part) => sum + part.length, 0));
  let offset = 0;
  for (const part of parts) {
    result.set(part, offset);
    offset += part.length;
  }
  return result;
}

function encodeRoute(route) {
  const routeBytes = textEncoder.encode(route);
  if (routeBytes.length > 0xff) {
    throw new Error("Pitaya route is longer than 255 bytes");
  }
  return concat([Uint8Array.of(routeBytes.length), routeBytes]);
}

export function encodeMessage(requestId, route, body) {
  if (requestId <= 0) {
    throw new Error("requestId must be positive");
  }
  const payload = concat([
    Uint8Array.of(MESSAGE_REQUEST << 1),
    encodeVarint(requestId),
    encodeRoute(route),
    textEncoder.encode(JSON.stringify(body)),
  ]);
  return encodePacket(PACKET_DATA, payload);
}

export function encodeNotify(route, body) {
  const payload = concat([
    Uint8Array.of(MESSAGE_NOTIFY << 1),
    encodeRoute(route),
    textEncoder.encode(JSON.stringify(body)),
  ]);
  return encodePacket(PACKET_DATA, payload);
}

async function inflate(body) {
  try {
    const stream = new Blob([body]).stream().pipeThrough(new DecompressionStream("deflate"));
    return new Uint8Array(await new Response(stream).arrayBuffer());
  } catch (error) {
    throw new Error("Invalid compressed Pitaya message body", { cause: error });
  }
}

export async function decodeMessage(payload) {
  if (payload.length === 0) {
    throw new Error("Empty Pitaya message");
  }

  let offset = 0;
  const flag = payload[offset++];
  const messageType = (flag >> 1) & MESSAGE_TYPE_MASK;
  const compressedRoute = (flag & 1) === 1;
  const compressedBody = (flag & MESSAGE_GZIP_MASK) !== 0;

  let requestId = 0;
  if (messageType === MESSAGE_REQUEST || messageType === MESSAGE_RESPONSE) {
    [requestId, offset] = decodeVarint(payload, offset);
  }
  if (messageType === MESSAGE_RESPONSE) {
    let body = payload.subarray(offset);
    if (compressedBody) body = await inflate(body);
    return { messageType, requestId, route: "", body };
  }

  if (compressedRoute) {
    throw new Error("Compressed Pitaya routes are not supported");
  }
  if (messageType !== MESSAGE_REQUEST && messageType !== MESSAGE_PUSH) {
    throw new Error(`Unknown Pitaya message type ${messageType}`);
  }
  if (offset >= payload.length) {
    throw new Error("Missing Pitaya route length");
  }
  const routeLength = payload[offset++];
  const routeEnd = offset + routeLength;
  if (routeEnd > payload.length) {
    throw new Error("Invalid Pitaya route length");
  }
  const route = textDecoder.decode(payload.subarray(offset, routeEnd));
  let body = payload.subarray(routeEnd);
  if (compressedBody) body = await inflate(body);
  return { messageType, requestId, route, body };
}

export function encodeVarint(value) {
  const result = [];
  while (value >= 0x80) {
    result.push((value & 0x7f) | 0x80);
    value = Math.floor(value / 128);
  }
  result.push(value);
  return Uint8Array.from(result);
}

export function decodeVarint(data, offset) {
  let value = 0;
  let multiplier = 1;
  for (let i = 0; offset < data.length && i < 8; i++) {
    const current = data[offset++];
    value += (current & 0x7f) * multiplier;
    if ((current & 0x80) === 0) {
      return [value, offset];
    }
    multiplier *= 128;
  }
  throw new Error("Invalid Pitaya varint");
}

export function decodeJson(payload) {
  if (payload.length === 0) return null;
  return JSON.parse(textDecoder.decode(payload));
}
