
import websocket

from pitaya_protocol import (
    MESSAGE_RESPONSE,
    PACKET_DATA,
    PACKET_HANDSHAKE_ACK,
    PACKET_HANDSHAKE,
    PACKET_HEARTBEAT,
    decode_json,
    decode_message,
    decode_packet,
    encode_handshake,
    encode_message,
    encode_packet,
)


class PitayaClient:
    def __init__(self, url: str):
        self.url = url
        self.ws = None
        self.request_id = 0

    def connect(self):
        print(f"Connecting to server at {self.url}...")
        self.ws = websocket.create_connection(self.url)
        self.ws.send(encode_handshake(), opcode=websocket.ABNF.OPCODE_BINARY)

        packet_type, payload = self._receive_packet()
        if packet_type != PACKET_HANDSHAKE:
            raise ConnectionError(
                f"Unexpected Pitaya handshake packet type: {packet_type}"
            )
        decode_json(payload)
        self.ws.send(
            encode_packet(PACKET_HANDSHAKE_ACK),
            opcode=websocket.ABNF.OPCODE_BINARY,
        )
        print("Connected to server.")

    def close(self):
        if self.ws:
            self.ws.close()
            self.ws = None
            print("Connection closed.")

    def login(self, account_id: str, password: str) -> object:
        return self.request(
            "AccountHandler.Login",
            {
                "accountId": account_id,
                "password": password,
            },
        )

    def request(self, route: str, body: object) -> object:
        if not self.ws:
            raise Exception("WebSocket connection is not established.")

        self.request_id += 1
        request_id = self.request_id
        request = encode_message(request_id, route, body)
        self.ws.send(request, opcode=websocket.ABNF.OPCODE_BINARY)

        while True:
            packet_type, payload = self._receive_packet()
            if packet_type == PACKET_HEARTBEAT:
                self.ws.send(
                    encode_packet(PACKET_HEARTBEAT),
                    opcode=websocket.ABNF.OPCODE_BINARY,
                )
                continue
            if packet_type != PACKET_DATA:
                raise ConnectionError(
                    f"Unexpected Pitaya response packet type: {packet_type}"
                )

            try:
                message = decode_message(payload)
            except ValueError as error:
                raise ValueError(
                    f"Failed to decode Pitaya DATA payload: {payload.hex()}"
                ) from error
            if message.message_type != MESSAGE_RESPONSE:
                continue
            if message.request_id != request_id:
                continue
            return decode_json(message.body)

    def _receive_packet(self) -> tuple[int, bytes]:
        if not self.ws:
            raise Exception("WebSocket connection is not established.")

        data = self.ws.recv()
        if isinstance(data, str):
            raise ValueError("Expected a binary Pitaya packet")
        return decode_packet(data)