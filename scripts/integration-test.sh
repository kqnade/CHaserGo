#!/bin/bash
set -e

echo "=== CHaser Integration Test ==="

# テストマップのパスを確認
if [ -f "testdata/RandMap_1.map" ]; then
    MAP_FILE="testdata/RandMap_1.map"
elif [ -f "testdata/sample_map.txt" ]; then
    MAP_FILE="testdata/sample_map.txt"
else
    echo "Error: No test map found"
    exit 1
fi

echo "Using map: $MAP_FILE"

# サーバーをバックグラウンドで起動
echo "Starting server..."
./bin/chaser-server -nd "$MAP_FILE" &
SERVER_PID=$!

# サーバーの起動を待つ
sleep 2

# サーバーが起動しているか確認
if ! ps -p $SERVER_PID > /dev/null; then
    echo "Error: Server failed to start"
    exit 1
fi

echo "Server started (PID: $SERVER_PID)"

# test1とtest2を並行実行
echo "Starting test clients..."

# test1をポート2009で実行
(cd examples/test1 && CHASER_HOST=127.0.0.1 CHASER_PORT=2009 timeout 30s go run main.go) &
CLIENT1_PID=$!

# test2をポート2010で実行
(cd examples/test2 && CHASER_HOST=127.0.0.1 CHASER_PORT=2010 timeout 30s go run main.go) &
CLIENT2_PID=$!

# クライアントの終了を待つ
echo "Waiting for clients to finish..."
wait $CLIENT1_PID
CLIENT1_EXIT=$?

wait $CLIENT2_PID
CLIENT2_EXIT=$?

# サーバーの終了を待つ
echo "Waiting for server to finish..."
wait $SERVER_PID
SERVER_EXIT=$?

echo "Server exited with code: $SERVER_EXIT"
echo "Client 1 exited with code: $CLIENT1_EXIT"
echo "Client 2 exited with code: $CLIENT2_EXIT"

# 結果判定
if [ $SERVER_EXIT -eq 0 ] || [ $SERVER_EXIT -eq 124 ]; then
    # サーバーが正常終了またはタイムアウト（想定内）
    if [ $CLIENT1_EXIT -eq 0 ] || [ $CLIENT1_EXIT -eq 124 ]; then
        if [ $CLIENT2_EXIT -eq 0 ] || [ $CLIENT2_EXIT -eq 124 ]; then
            echo "=== Integration Test PASSED ==="
            exit 0
        fi
    fi
fi

echo "=== Integration Test FAILED ==="
exit 1
