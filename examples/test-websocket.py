#!/usr/bin/env python3
"""
Simple WebSocket test client
Requires: pip install websocket-client
"""

import sys
import time
try:
    import websocket
except ImportError:
    print("Installing websocket-client...")
    import subprocess
    subprocess.check_call([sys.executable, "-m", "pip", "install", "websocket-client"])
    import websocket

def on_message(ws, message):
    print(f"📩 Received: {message}")

def on_error(ws, error):
    print(f"❌ Error: {error}")

def on_close(ws, close_status_code, close_msg):
    print("🔌 Connection closed")

def on_open(ws):
    print("✅ Connected to server")

    # Send test messages
    messages = [
        "Hello from Python client!",
        "This is a test message",
        "WebSocket is working!"
    ]

    for msg in messages:
        ws.send(msg)
        print(f"📤 Sent: {msg}")
        time.sleep(0.5)

    # Close connection
    time.sleep(1)
    ws.close()

if __name__ == "__main__":
    server_url = "ws://localhost:8096/ws"

    print(f"Connecting to {server_url}...")

    ws = websocket.WebSocketApp(server_url,
        on_open=on_open,
        on_message=on_message,
        on_error=on_error,
        on_close=on_close)

    # Run WebSocket
    ws.run_forever()
