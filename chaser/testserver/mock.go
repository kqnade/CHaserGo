package testserver

import (
	"bufio"
	"fmt"
	"net"
	"sync"
)

// MockServer はテスト用のモックCHaserサーバー
type MockServer struct {
	listener   net.Listener
	port       string
	responses  []string // レスポンスのキュー
	mu         sync.Mutex
	stopChan   chan struct{}
	clientConn net.Conn
	running    bool
}

// NewMockServer は新しいモックサーバーを作成する
// port: リスニングポート（例: "0" で自動割り当て）
func NewMockServer(port string) *MockServer {
	return &MockServer{
		port:      port,
		responses: []string{},
		stopChan:  make(chan struct{}),
	}
}

// Start はモックサーバーを起動する
func (ms *MockServer) Start() error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.running {
		return fmt.Errorf("server already running")
	}

	listener, err := net.Listen("tcp", "127.0.0.1:"+ms.port)
	if err != nil {
		return fmt.Errorf("failed to start mock server: %w", err)
	}

	ms.listener = listener
	ms.running = true

	// 実際のポートを取得（port="0"の場合に自動割り当て）
	ms.port = fmt.Sprintf("%d", listener.Addr().(*net.TCPAddr).Port)

	// 接続を受け付けるgoroutine
	go ms.acceptConnections()

	return nil
}

// Stop はモックサーバーを停止する
func (ms *MockServer) Stop() error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if !ms.running {
		return nil
	}

	close(ms.stopChan)
	ms.running = false

	if ms.clientConn != nil {
		ms.clientConn.Close()
	}

	if ms.listener != nil {
		return ms.listener.Close()
	}

	return nil
}

// Port は現在リスニング中のポート番号を返す
func (ms *MockServer) Port() string {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	return ms.port
}

// SetResponses はサーバーが返すレスポンスのリストを設定する
// 各レスポンスは順番に使用される
func (ms *MockServer) SetResponses(responses []string) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.responses = make([]string, len(responses))
	copy(ms.responses, responses)
}

// acceptConnections はクライアント接続を受け付ける
func (ms *MockServer) acceptConnections() {
	for {
		select {
		case <-ms.stopChan:
			return
		default:
		}

		conn, err := ms.listener.Accept()
		if err != nil {
			select {
			case <-ms.stopChan:
				return
			default:
				continue
			}
		}

		ms.mu.Lock()
		ms.clientConn = conn
		ms.mu.Unlock()

		go ms.handleClient(conn)
	}
}

// handleClient はクライアントとの通信を処理する
func (ms *MockServer) handleClient(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	// 1. 名前を受信
	name, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	_ = name // 名前は使用しないが、プロトコルに従って受信

	responseIndex := 0

	for {
		select {
		case <-ms.stopChan:
			return
		default:
		}

		// コマンドを受信
		cmd, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		// コマンドの種類を判定
		isGetReady := len(cmd) >= 2 && cmd[0:2] == "gr"

		ms.mu.Lock()
		var response string
		if responseIndex < len(ms.responses) {
			response = ms.responses[responseIndex]
			responseIndex++
		} else {
			// デフォルトレスポンス（ゲーム継続中）
			response = "1000000000"
		}
		ms.mu.Unlock()

		if isGetReady {
			// getReadyの場合: 初期行送信 + レスポンス
			writer.WriteString("Ready\n")
			writer.Flush()
		}

		// レスポンス送信
		writer.WriteString(response + "\n")
		writer.Flush()

		if !isGetReady {
			// getReady以外: 確認応答を受信
			_, err = reader.ReadString('\n')
			if err != nil {
				return
			}
		}
	}
}
