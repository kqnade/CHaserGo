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
	defer func() {
		_ = server.Stop()
	}()

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
	defer func() {
		_ = server.Stop()
	}()

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
	defer func() {
		_ = server.Stop()
	}()

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

// TestMultipleCommands は複数のコマンド送信をテスト（Ready → Walk パターン）
func TestMultipleCommands(t *testing.T) {
	// モックサーバー起動
	server := testserver.NewMockServer("0")
	server.SetResponses([]string{
		"1000000000", // Ready #1
		"1222000000", // Walk #1
		"1000222000", // Ready #2
		"1000000222", // Walk #2
	})
	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start mock server: %v", err)
	}
	defer func() {
		_ = server.Stop()
	}()

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

	// Ready → Walk を2回繰り返す
	for i := 0; i < 2; i++ {
		// Ready呼び出し
		resp, err := client.Ready(ctx)
		if err != nil {
			t.Fatalf("Ready() call %d failed: %v", i+1, err)
		}
		if resp.GameOver {
			t.Errorf("Ready call %d: unexpected GameOver", i+1)
		}

		// Walk呼び出し
		resp, err = client.Walk(ctx, Up)
		if err != nil {
			t.Fatalf("Walk() call %d failed: %v", i+1, err)
		}
		if resp.GameOver {
			t.Errorf("Walk call %d: unexpected GameOver", i+1)
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
			defer func() {
		_ = server.Stop()
	}()

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

// TestWalk はWalkメソッドをテスト
func TestWalk(t *testing.T) {
	tests := []struct {
		name      string
		direction Direction
		response  string
		wantValue CellType
	}{
		{name: "Walk Up", direction: Up, response: "1000000000", wantValue: Enemy},
		{name: "Walk Down", direction: Down, response: "1222000000", wantValue: Enemy},
		{name: "Walk Left", direction: Left, response: "1000222000", wantValue: Enemy},
		{name: "Walk Right", direction: Right, response: "1000000222", wantValue: Enemy},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := testserver.NewMockServer("0")
			server.SetResponses([]string{"1000000000", tt.response})
			err := server.Start()
			if err != nil {
				t.Fatalf("Failed to start server: %v", err)
			}
			defer func() {
		_ = server.Stop()
	}()

			time.Sleep(50 * time.Millisecond)

			config := ClientConfig{Host: "127.0.0.1", Port: server.Port(), Name: "test"}
			client := NewClient(config)
			ctx := context.Background()

			err = client.Connect(ctx)
			if err != nil {
				t.Fatalf("Failed to connect: %v", err)
			}
			defer client.Disconnect()

			// Readyを先に呼び出して初期メッセージを消費
			_, err = client.Ready(ctx)
			if err != nil {
				t.Fatalf("Ready() failed: %v", err)
			}

			resp, err := client.Walk(ctx, tt.direction)
			if err != nil {
				t.Fatalf("Walk() failed: %v", err)
			}

			if resp.Values[0] != tt.wantValue {
				t.Errorf("Expected Values[0]=%v, got %v", tt.wantValue, resp.Values[0])
			}
		})
	}
}

// TestLook はLookメソッドをテスト
func TestLook(t *testing.T) {
	tests := []struct {
		name      string
		direction Direction
		response  string
	}{
		{name: "Look Up", direction: Up, response: "1000200000"},
		{name: "Look Down", direction: Down, response: "1000000020"},
		{name: "Look Left", direction: Left, response: "1020000000"},
		{name: "Look Right", direction: Right, response: "1000020000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := testserver.NewMockServer("0")
			server.SetResponses([]string{"1000000000", tt.response})
			err := server.Start()
			if err != nil {
				t.Fatalf("Failed to start server: %v", err)
			}
			defer func() {
		_ = server.Stop()
	}()

			time.Sleep(50 * time.Millisecond)

			config := ClientConfig{Host: "127.0.0.1", Port: server.Port(), Name: "test"}
			client := NewClient(config)
			ctx := context.Background()

			err = client.Connect(ctx)
			if err != nil {
				t.Fatalf("Failed to connect: %v", err)
			}
			defer client.Disconnect()

			// Readyを先に呼び出して初期メッセージを消費
			_, err = client.Ready(ctx)
			if err != nil {
				t.Fatalf("Ready() failed: %v", err)
			}

			resp, err := client.Look(ctx, tt.direction)
			if err != nil {
				t.Fatalf("Look() failed: %v", err)
			}

			if resp == nil {
				t.Fatal("Response is nil")
			}
		})
	}
}

// TestSearch はSearchメソッドをテスト
func TestSearch(t *testing.T) {
	tests := []struct {
		name      string
		direction Direction
		response  string
	}{
		{name: "Search Up", direction: Up, response: "1000300000"},
		{name: "Search Down", direction: Down, response: "1000000030"},
		{name: "Search Left", direction: Left, response: "1030000000"},
		{name: "Search Right", direction: Right, response: "1000030000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := testserver.NewMockServer("0")
			server.SetResponses([]string{"1000000000", tt.response})
			err := server.Start()
			if err != nil {
				t.Fatalf("Failed to start server: %v", err)
			}
			defer func() {
		_ = server.Stop()
	}()

			time.Sleep(50 * time.Millisecond)

			config := ClientConfig{Host: "127.0.0.1", Port: server.Port(), Name: "test"}
			client := NewClient(config)
			ctx := context.Background()

			err = client.Connect(ctx)
			if err != nil {
				t.Fatalf("Failed to connect: %v", err)
			}
			defer client.Disconnect()

			// Readyを先に呼び出して初期メッセージを消費
			_, err = client.Ready(ctx)
			if err != nil {
				t.Fatalf("Ready() failed: %v", err)
			}

			resp, err := client.Search(ctx, tt.direction)
			if err != nil {
				t.Fatalf("Search() failed: %v", err)
			}

			if resp == nil {
				t.Fatal("Response is nil")
			}
		})
	}
}

// TestPut はPutメソッドをテスト
func TestPut(t *testing.T) {
	tests := []struct {
		name      string
		direction Direction
		response  string
	}{
		{name: "Put Up", direction: Up, response: "1000200000"},
		{name: "Put Down", direction: Down, response: "1000000020"},
		{name: "Put Left", direction: Left, response: "1020000000"},
		{name: "Put Right", direction: Right, response: "1000020000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := testserver.NewMockServer("0")
			server.SetResponses([]string{"1000000000", tt.response})
			err := server.Start()
			if err != nil {
				t.Fatalf("Failed to start server: %v", err)
			}
			defer func() {
		_ = server.Stop()
	}()

			time.Sleep(50 * time.Millisecond)

			config := ClientConfig{Host: "127.0.0.1", Port: server.Port(), Name: "test"}
			client := NewClient(config)
			ctx := context.Background()

			err = client.Connect(ctx)
			if err != nil {
				t.Fatalf("Failed to connect: %v", err)
			}
			defer client.Disconnect()

			// Readyを先に呼び出して初期メッセージを消費
			_, err = client.Ready(ctx)
			if err != nil {
				t.Fatalf("Ready() failed: %v", err)
			}

			resp, err := client.Put(ctx, tt.direction)
			if err != nil {
				t.Fatalf("Put() failed: %v", err)
			}

			if resp == nil {
				t.Fatal("Response is nil")
			}
		})
	}
}

// TestSetDeadline はSetDeadlineメソッドをテスト
func TestSetDeadline(t *testing.T) {
	server := testserver.NewMockServer("0")
	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer func() {
		_ = server.Stop()
	}()

	time.Sleep(50 * time.Millisecond)

	config := ClientConfig{Host: "127.0.0.1", Port: server.Port(), Name: "test"}
	client := NewClient(config)
	ctx := context.Background()

	// 未接続時のSetDeadline
	err = client.SetDeadline(time.Now().Add(1 * time.Second))
	if err != ErrNotConnected {
		t.Errorf("Expected ErrNotConnected, got %v", err)
	}

	// 接続後のSetDeadline
	err = client.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Disconnect()

	err = client.SetDeadline(time.Now().Add(1 * time.Second))
	if err != nil {
		t.Errorf("SetDeadline() failed: %v", err)
	}
}

// TestConnectErrors はConnect時のエラーケースをテスト
func TestConnectErrors(t *testing.T) {
	// すでに接続済みの場合
	server := testserver.NewMockServer("0")
	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer func() {
		_ = server.Stop()
	}()

	time.Sleep(50 * time.Millisecond)

	config := ClientConfig{Host: "127.0.0.1", Port: server.Port(), Name: "test"}
	client := NewClient(config)
	ctx := context.Background()

	err = client.Connect(ctx)
	if err != nil {
		t.Fatalf("First connect failed: %v", err)
	}
	defer client.Disconnect()

	// 2回目の接続
	err = client.Connect(ctx)
	if err != ErrAlreadyConnected {
		t.Errorf("Expected ErrAlreadyConnected, got %v", err)
	}
}

// TestReadyErrors はReady時のエラーケースをテスト
func TestReadyErrors(t *testing.T) {
	// 未接続時のReady
	config := ClientConfig{Host: "127.0.0.1", Port: "12345", Name: "test"}
	client := NewClient(config)
	ctx := context.Background()

	_, err := client.Ready(ctx)
	if err != ErrNotConnected {
		t.Errorf("Expected ErrNotConnected, got %v", err)
	}
}

// TestDisconnectWhenNotConnected は未接続時のDisconnectをテスト
func TestDisconnectWhenNotConnected(t *testing.T) {
	config := ClientConfig{Host: "127.0.0.1", Port: "12345", Name: "test"}
	client := NewClient(config)

	// 未接続時のDisconnect（エラーにならない）
	err := client.Disconnect()
	if err != nil {
		t.Errorf("Disconnect() on unconnected client returned error: %v", err)
	}
}

// TestDirectionStringDefault はDirection.String()のデフォルトケースをテスト
func TestDirectionStringDefault(t *testing.T) {
	invalid := Direction(99)
	str := invalid.String()
	if str != "Direction(99)" {
		t.Errorf("Expected 'Direction(99)', got %q", str)
	}
}

// TestCellTypeStringDefault はCellType.String()のデフォルトケースをテスト
func TestCellTypeStringDefault(t *testing.T) {
	invalid := CellType(99)
	str := invalid.String()
	if str != "CellType(99)" {
		t.Errorf("Expected 'CellType(99)', got %q", str)
	}
}

// TestWalkNotConnected は未接続時のWalkをテスト
func TestWalkNotConnected(t *testing.T) {
	config := ClientConfig{Host: "127.0.0.1", Port: "12345", Name: "test"}
	client := NewClient(config)
	ctx := context.Background()

	_, err := client.Walk(ctx, Up)
	if err != ErrNotConnected {
		t.Errorf("Expected ErrNotConnected, got %v", err)
	}
}

// TestLookNotConnected は未接続時のLookをテスト
func TestLookNotConnected(t *testing.T) {
	config := ClientConfig{Host: "127.0.0.1", Port: "12345", Name: "test"}
	client := NewClient(config)
	ctx := context.Background()

	_, err := client.Look(ctx, Up)
	if err != ErrNotConnected {
		t.Errorf("Expected ErrNotConnected, got %v", err)
	}
}
