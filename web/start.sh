#!/bin/sh
# 啟動網頁測試 client（純靜態檔，用 Python 內建 http.server）。
# 用法：sh web/start.sh [port]，預設 5173，也可用環境變數 PORT 指定。
# 回應加上 Cache-Control: no-store，避免瀏覽器快取舊的 JS 模組。

PORT="${1:-${PORT:-5173}}"
cd "$(dirname "$0")" || exit 1

echo "Web client: http://localhost:${PORT}"
exec python3 - "$PORT" <<'PY'
import sys
from functools import partial
from http.server import SimpleHTTPRequestHandler, ThreadingHTTPServer


class NoStoreHandler(SimpleHTTPRequestHandler):
    def end_headers(self):
        self.send_header("Cache-Control", "no-store")
        super().end_headers()


handler = partial(NoStoreHandler, directory=".")
with ThreadingHTTPServer(("", int(sys.argv[1])), handler) as server:
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        pass
PY
