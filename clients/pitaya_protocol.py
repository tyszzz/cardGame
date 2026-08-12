import json
from dataclasses import dataclass


PACKET_HANDSHAKE = 1
PACKET_HANDSHAKE_ACK = 2
PACKET_HEARTBEAT = 3
PACKET_DATA = 4
PACKET_KICK = 5

MESSAGE_REQUEST = 0
MESSAGE_RESPONSE = 1
MESSAGE_PUSH = 2


@dataclass
class Message:
    message_type: int
    request_id: int
    route: str
    body: bytes


def encode_packet(packet_type: int, payload: bytes = b"") -> bytes:
    payload_length = len(payload)
    if payload_length > 0xFFFFFF:
        raise ValueError("Pitaya packet payload is too large")

    return bytes((packet_type,)) + payload_length.to_bytes(3, "big") + payload


def decode_packet(data: bytes) -> tuple[int, bytes]:
    if len(data) < 4:
        raise ValueError("Invalid Pitaya packet header")

    packet_type = data[0]
    payload_length = int.from_bytes(data[1:4], "big")
    payload = data[4:]
    if len(payload) != payload_length:
        raise ValueError("Incomplete Pitaya packet")

    return packet_type, payload


def encode_handshake() -> bytes:
    payload = {
        "sys": {
            "type": "python",
            "version": "1.0.0",
            "heartbeat": 0,
        },
        "user": {},
    }
    return encode_packet(PACKET_HANDSHAKE, json.dumps(payload).encode("utf-8"))


def encode_message(request_id: int, route: str, body: object) -> bytes:
    if request_id <= 0:
        raise ValueError("request_id must be positive")

    route_bytes = route.encode("utf-8")
    body_bytes = json.dumps(body, separators=(",", ":")).encode("utf-8")
    message_type = (MESSAGE_REQUEST << 1) | 0

    payload = bytes((message_type,))
    payload += encode_varint(request_id)
    payload += encode_varint(len(route_bytes))
    payload += route_bytes
    payload += body_bytes
    return encode_packet(PACKET_DATA, payload)


def decode_message(payload: bytes) -> Message:
    if not payload:
        raise ValueError("Empty Pitaya message")

    # Pitaya's JSON response frame uses 0x04 followed by the request id.
    if payload[0] == 0x04:
        if len(payload) < 2:
            raise ValueError("Incomplete Pitaya response message")
        return Message(MESSAGE_RESPONSE, payload[1], "", payload[2:])

    offset = 0
    message_type = payload[offset] >> 1
    compressed_route = payload[offset] & 1
    offset += 1

    request_id, offset = decode_varint(payload, offset)
    if message_type == MESSAGE_RESPONSE:
        return Message(message_type, request_id, "", payload[offset:])

    if compressed_route:
        raise ValueError("Compressed Pitaya routes are not supported")

    if message_type not in (0, MESSAGE_PUSH):
        raise ValueError(
            f"Unknown Pitaya message type {message_type}; "
            f"payload={payload.hex()}"
        )

    route_length, offset = decode_varint(payload, offset)
    route_end = offset + route_length
    if route_end > len(payload):
        raise ValueError("Invalid Pitaya route length")

    route = payload[offset:route_end].decode("utf-8")
    return Message(message_type, request_id, route, payload[route_end:])


def encode_varint(value: int) -> bytes:
    result = bytearray()
    while value >= 0x80:
        result.append((value & 0x7F) | 0x80)
        value >>= 7
    result.append(value)
    return bytes(result)


def decode_varint(data: bytes, offset: int) -> tuple[int, int]:
    value = 0
    shift = 0
    while offset < len(data):
        current = data[offset]
        offset += 1
        value |= (current & 0x7F) << shift
        if current & 0x80 == 0:
            return value, offset
        shift += 7
        if shift >= 64:
            break

    raise ValueError("Invalid Pitaya varint")


def decode_json(payload: bytes) -> object:
    if not payload:
        return None
    return json.loads(payload.decode("utf-8"))
