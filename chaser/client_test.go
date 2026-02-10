package chaser

import (
	"context"
	"testing"
	"time"

	"CHaserGo/chaser/testserver"
)

// TestNewClient はクライアントの作成をテスト
func TestNewClient(t *testing.T) {
	config := ClientConfig{
		Host: "127.0.0.1",
		Port: "12345",
		Name: "test",
	}

	client := NewClient(config)
	if client == nil {
		t.Fatal("NewClient() returned nil")
	}

	if client.config.Name != "test" {
		t.Errorf("Expected name 'test', got %q", client.config.Name)
	}
}

// TestConnectAndDisconnect はサーバーへの接続と切断をテスト
func TestConnectAndDisconnect(t *testing.T) {
	// モックサーバー起動
	server := testserver.NewMockServer("0")
	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start mock server: %v", err)
	}
	defer server.Stop()

	time.Sleep(50 * time.Millisecond)

	// クライアント作成と接続
	config := ClientConfig{
		Host: "127.0.0.1",
		Port: server.Port(),
		Name: "testclient",
	}

	client := NewClient(config)
	ctx := context.Background()

	err = client.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// 切断
	err = client.Disconnect()
	if err != nil {
		t.Errorf("Failed to disconnect: %v", err)
	}
}

// TestReady は準備信号の送受信をテスト
func TestReady(t *testing.T) {
	// モックサーバー起動
	server := testserver.NewMockServer("0")
	server.SetResponses([]string{"1000000000"})
	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start mock server: %v", err)
	}
	defer server.Stop()

	time.Sleep(50 * time.Millisecond)

	// クライアント接続
	config := ClientConfig{
		Host: "127.0.0.1",
		Port: server.Port(),
		Name: "testclient",
	}

	client := NewClient(config)
	ctx := context.Background()

	err = client.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Disconnect()

	// Ready呼び出し
	resp, err := client.Ready(ctx)
	if err != nil {
		t.Fatalf("Ready() failed: %v", err)
	}

	// レスポンス確認
	if resp.GameOver {
		t.Error("Expected GameOver=false, got true")
	}

	if resp.Values[0] != Enemy {
		t.Errorf("Expected Values[0]=Enemy, got %v", resp.Values[0])
	}
}

// TestReadyGameOver はゲームオーバー時のReadyをテスト
func TestReadyGameOver(t *testing.T) {
	// モックサーバー起動（ゲームオーバーレスポンス）
	server := testserver.NewMockServer("0")
	server.SetResponses([]string{"0000000000"})
	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start mock server: %v", err)
	}
	defer server.Stop()

	time.Sleep(50 * time.Millisecond)

	// クライアント接続
	config := ClientConfig{
		Host: "127.0.0.1",
		Port: server.Port(),
		Name: "testclient",
	}

	client := NewClient(config)
	ctx := context.Background()

	err = client.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Disconnect()

	// Ready呼び出し
	resp, err := client.Ready(ctx)
	if err != nil {
		t.Fatalf("Ready() failed: %v", err)
	}

	// ゲームオーバー確認
	if !resp.GameOver {
		t.Error("Expected GameOver=true, got false")
	}
}

// TestConnectWithTimeout はタイムアウト付き接続をテスト
func TestConnectWithTimeout(t *testing.T) {
	// 存在しないポートに接続を試みる
	config := ClientConfig{
		Host: "127.0.0.1",
		Port: "54321", // 使用されていないポート
		Name: "testclient",
	}

	client := NewClient(config)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := client.Connect(ctx)
	if err == nil {
		t.Error("Expected connection error, got nil")
		client.Disconnect()
	}
}

// TestMultipleCommands は複数のコマンド送信をテスト
func TestMultipleCommands(t *testing.T) {
	// モックサーバー起動
	server := testserver.NewMockServer("0")
	server.SetResponses([]string{
		"1000000000",
		"1222000000",
		"1000222000",
	})
	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start mock server: %v", err)
	}
	defer server.Stop()

	time.Sleep(50 * time.Millisecond)

	// クライアント接続
	config := ClientConfig{
		Host: "127.0.0.1",
		Port: server.Port(),
		Name: "testclient",
	}

	client := NewClient(config)
	ctx := context.Background()

	err = client.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Disconnect()

	// 3回Readyを呼び出す
	for i := 0; i < 3; i++ {
		resp, err := client.Ready(ctx)
		if err != nil {
			t.Fatalf("Ready() call %d failed: %v", i+1, err)
		}
		if resp.GameOver {
			t.Errorf("Call %d: unexpected GameOver", i+1)
		}
	}
}

// TestEncodeNameForPortInConfig は接続時の名前エンコーディングをテスト
func TestEncodeNameForPortInConfig(t *testing.T) {
	tests := []struct {
		name     string
		port     string
		userName string
	}{
		{
			name:     "UTF-8 port",
			port:     "2001",
			userName: "テスト1",
		},
		{
			name:     "CP932 port 40000",
			port:     "40000",
			userName: "テスト2",
		},
		{
			name:     "CP932 port 50000",
			port:     "50000",
			userName: "テスト3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックサーバー起動
			server := testserver.NewMockServer(tt.port)
			err := server.Start()
			if err != nil {
				t.Fatalf("Failed to start mock server: %v", err)
			}
			defer server.Stop()

			time.Sleep(50 * time.Millisecond)

			// クライアント接続（エンコーディングテスト）
			config := ClientConfig{
				Host: "127.0.0.1",
				Port: server.Port(), // 実際のポートを使用
				Name: tt.userName,
			}

			client := NewClient(config)
			ctx := context.Background()

			// 接続できればエンコーディングは正常
			err = client.Connect(ctx)
			if err != nil {
				t.Errorf("Failed to connect with %s encoding: %v", tt.name, err)
			} else {
				client.Disconnect()
			}
		})
	}
}
