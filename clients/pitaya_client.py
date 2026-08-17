
import queue
import threading

import websocket

from pitaya_protocol import (
    MESSAGE_PUSH,
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
    encode_notify,
    encode_packet,
)


class PitayaClient:
    def __init__(self, url: str):
        self.url = url
        self.ws = None
        self.request_id = 0
        self.pending_requests = {}
        self.pending_lock = threading.Lock()
        self.send_lock = threading.Lock()
        self.route_handlers = {}
        self.stop_event = threading.Event()
        self.receiver_thread = None

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
        self.stop_event.clear()
        self.receiver_thread = threading.Thread(
            target=self._receive_loop,
            daemon=True,
        )
        self.receiver_thread.start()
        print("Connected to server.")

    def register_route(self, route: str, callback):
        self.route_handlers[route] = callback

    def unregister_route(self, route: str):
        self.route_handlers.pop(route, None)

    def close(self):
        if self.ws:
            self.stop_event.set()
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

    def create_account(self, account_id: str, password: str) -> object:
        return self.request(
            "AccountHandler.CreateNewAccount",
            {
                "accountId": account_id,
                "password": password,
            },
        )

    def summon(self, uid: str , times: int , summon_type: str) -> object:
        return self.request(
            "SummonHandler.SummonCard",
            {
                "uid": uid,
                "times": times,
                "summonType": summon_type,
            },
        )

    def publish(self, route: str, body: object) -> None:
        if not self.ws:
            raise Exception("WebSocket connection is not established.")

        message = encode_notify(route, body)
        with self.send_lock:
            self.ws.send(message, opcode=websocket.ABNF.OPCODE_BINARY)

    def request(self, route: str, body: object) -> object:
        if not self.ws:
            raise Exception("WebSocket connection is not established.")

        self.request_id += 1
        request_id = self.request_id
        response_queue = queue.Queue(maxsize=1)
        with self.pending_lock:
            self.pending_requests[request_id] = response_queue

        request = encode_message(request_id, route, body)
        try:
            with self.send_lock:
                self.ws.send(request, opcode=websocket.ABNF.OPCODE_BINARY)
            result = response_queue.get()
            if isinstance(result, Exception):
                raise result
            return result
        finally:
            with self.pending_lock:
                self.pending_requests.pop(request_id, None)

    def _receive_loop(self):
        while not self.stop_event.is_set():
            try:
                packet_type, payload = self._receive_packet()
                if packet_type == PACKET_HEARTBEAT:
                    with self.send_lock:
                        self.ws.send(
                            encode_packet(PACKET_HEARTBEAT),
                            opcode=websocket.ABNF.OPCODE_BINARY,
                        )
                    continue
                if packet_type != PACKET_DATA:
                    continue

                message = decode_message(payload)
                if message.message_type == MESSAGE_RESPONSE:
                    with self.pending_lock:
                        response_queue = self.pending_requests.get(message.request_id)
                    if response_queue:
                        response_queue.put(decode_json(message.body))
                elif message.message_type == MESSAGE_PUSH:
                    callback = self.route_handlers.get(message.route)
                    if callback:
                        callback(decode_json(message.body))
            except Exception as error:
                if not self.stop_event.is_set():
                    self._fail_pending_requests(error)
                    return

    def _fail_pending_requests(self, error: Exception):
        with self.pending_lock:
            pending_requests = list(self.pending_requests.values())
        for response_queue in pending_requests:
            response_queue.put(error)

    def _receive_packet(self) -> tuple[int, bytes]:
        if not self.ws:
            raise Exception("WebSocket connection is not established.")

        data = self.ws.recv()
        if isinstance(data, str):
            raise ValueError("Expected a binary Pitaya packet")
        return decode_packet(data)