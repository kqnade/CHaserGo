package testserver

import (
	"bufio"
	"net"
	"testing"
	"time"
)

// TestMockServerStartStop はモックサーバーの起動と停止をテスト
func TestMockServerStartStop(t *testing.T) {
	server := NewMockServer("0")

	// 起動
	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// ポートが割り当てられているか確認
	port := server.Port()
	if port == "" || port == "0" {
		t.Errorf("Port not assigned: got %q", port)
	}

	// 停止
	err = server.Stop()
	if err != nil {
		t.Errorf("Failed to stop server: %v", err)
	}
}

// TestMockServerConnection はモックサーバーへの接続をテスト
func TestMockServerConnection(t *testing.T) {
	server := NewMockServer("0")
	server.SetResponses([]string{"1000000000"})

	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	// サーバーが起動するまで少し待つ
	time.Sleep(50 * time.Millisecond)

	// 接続
	conn, err := net.Dial("tcp", "127.0.0.1:"+server.Port())
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	// 名前送信
	_, err = writer.WriteString("test\n")
	if err != nil {
		t.Fatalf("Failed to send name: %v", err)
	}
	writer.Flush()

	// getReadyコマンド送信
	_, err = writer.WriteString("gr\r\n")
	if err != nil {
		t.Fatalf("Failed to send command: %v", err)
	}
	writer.Flush()

	// 初期行受信
	line, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("Failed to read initial line: %v", err)
	}
	if line != "Ready\n" {
		t.Errorf("Expected 'Ready\\n', got %q", line)
	}

	// レスポンス受信
	response, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	// 改行を除去して確認
	response = response[:len(response)-1]
	if response != "1000000000" {
		t.Errorf("Expected '1000000000', got %q", response)
	}
}

// TestMockServerMultipleResponses は複数のレスポンスをテスト
func TestMockServerMultipleResponses(t *testing.T) {
	server := NewMockServer("0")
	responses := []string{
		"1000000000",
		"1222222222",
		"0000000000", // ゲームオーバー
	}
	server.SetResponses(responses)

	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	time.Sleep(50 * time.Millisecond)

	conn, err := net.Dial("tcp", "127.0.0.1:"+server.Port())
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	// 名前送信
	writer.WriteString("test\n")
	writer.Flush()

	// 各レスポンスを確認
	for i, expected := range responses {
		// getReadyコマンド
		writer.WriteString("gr\r\n")
		writer.Flush()

		// 初期行
		_, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("Command %d: failed to read initial line: %v", i, err)
		}

		// レスポンス
		response, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("Command %d: failed to read response: %v", i, err)
		}

		response = response[:len(response)-1]
		if response != expected {
			t.Errorf("Command %d: expected %q, got %q", i, expected, response)
		}
	}
}
