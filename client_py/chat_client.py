import threading, logging
import grpc
from tenacity import retry, wait_exponential, stop_after_attempt

import chat_pb2, chat_pb2_grpc  # sinh từ .proto

log = logging.getLogger("chat.client")

class ChatClient:
    def __init__(self, host="127.0.0.1:7070", user_id="user"):
        self.host = host
        self.user_id = user_id
        self._chan = grpc.insecure_channel(host)
        self.stub = chat_pb2_grpc.ChatServiceStub(self._chan)

    def join(self, room_id, since_seq=0, on_event=None):
        def run():
            stream = self.stub.Join(chat_pb2.JoinRequest(room_id=room_id, user_id=self.user_id, since_seq=since_seq))
            for ev in stream:
                if on_event: on_event(ev)
        threading.Thread(target=run, daemon=True).start()

    @retry(wait=wait_exponential(min=1, max=8), stop=stop_after_attempt(5))
    def send_text(self, room_id, text, typ="user_message", payload_json=""):
        r = chat_pb2.SendRequest(room_id=room_id, user_id=self.user_id, type=typ, text=text, payload_json=payload_json)
        return self.stub.Send(r)
