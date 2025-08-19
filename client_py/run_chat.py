import argparse, logging, json, sys
from chat_client import ChatClient

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")

def on_event(ev):
    print(f"[{ev.room_id}] #{ev.seq} {ev.type} {ev.user_id}: {ev.text}")

def main():
    p = argparse.ArgumentParser()
    p.add_argument("--host", default="127.0.0.1:7070")
    p.add_argument("--room", default="general")
    p.add_argument("--user", default="alice")
    args = p.parse_args()

    cli = ChatClient(args.host, args.user)
    cli.join(args.room, on_event=on_event)

    print("Type messages. Prefix '!err ' to send system_error with JSON payload.")
    for line in sys.stdin:
        msg = line.rstrip("\n")
        if msg.startswith("!err "):
            payload = {"app": "MediaHelper", "info": msg[5:]}
            cli.send_text(args.room, msg[5:], typ="system_error", payload_json=json.dumps(payload))
        else:
            cli.send_text(args.room, msg)

if __name__ == "__main__":
    main()
