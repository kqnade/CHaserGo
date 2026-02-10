package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

// Server represents the CHaser game server
type Server struct {
	HotPort    int
	CoolPort   int
	Board      *Board
	DumpSystem *DumpSystem
	HotConn    *Connection
	CoolConn   *Connection
}

// ServerConfig holds server configuration
type ServerConfig struct {
	MapPath   string
	HotPort   int
	CoolPort  int
	DumpPath  string
	EnableDump bool
}

// NewServer creates a new CHaser server
func NewServer(config ServerConfig) (*Server, error) {
	// ボードを読み込み
	board, err := NewBoard(config.MapPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load board: %w", err)
	}

	// ダンプシステムを初期化
	dumpSystem, err := NewDumpSystem(config.DumpPath, config.MapPath, config.EnableDump)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize dump system: %w", err)
	}

	return &Server{
		HotPort:    config.HotPort,
		CoolPort:   config.CoolPort,
		Board:      board,
		DumpSystem: dumpSystem,
	}, nil
}

// Start starts the server and waits for connections
func (s *Server) Start() error {
	log.Printf("Starting CHaser server...")
	log.Printf("Hot port: %d, Cool port: %d", s.HotPort, s.CoolPort)
	log.Printf("Max turns: %d", s.Board.MaxTurns)

	// 並行して両ポートで接続を待つ
	// エラー時に早期終了するためのcontext
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	errChan := make(chan error, 2)
	doneChan := make(chan struct{})

	wg.Add(2)

	// Hot接続待機
	go func() {
		defer wg.Done()
		conn, name, err := s.acceptConnectionWithContext(ctx, s.HotPort, "Hot")
		if err != nil {
			errChan <- fmt.Errorf("hot connection failed: %w", err)
			cancel() // もう片方をキャンセル
			return
		}
		s.HotConn = conn
		s.Board.Hot.Name = name
		log.Printf("Hot player connected: %s", name)
	}()

	// Cool接続待機
	go func() {
		defer wg.Done()
		conn, name, err := s.acceptConnectionWithContext(ctx, s.CoolPort, "Cool")
		if err != nil {
			errChan <- fmt.Errorf("cool connection failed: %w", err)
			cancel() // もう片方をキャンセル
			return
		}
		s.CoolConn = conn
		s.Board.Cool.Name = name
		log.Printf("Cool player connected: %s", name)
	}()

	// 完了待機
	go func() {
		wg.Wait()
		close(doneChan)
	}()

	// 最初のエラーまたは完了を待つ
	select {
	case err := <-errChan:
		cancel() // 残りをキャンセル
		<-doneChan // 全goroutineの終了を待つ
		return err
	case <-doneChan:
		// 正常完了
	}

	// プレイヤー名をダンプに記録
	if err := s.DumpSystem.SetNames(s.Board.Hot.Name, s.Board.Cool.Name); err != nil {
		log.Printf("Warning: failed to write names to dump: %v", err)
	}

	log.Println("Both players connected. Starting game...")

	// ゲームメインループ
	return s.runGame()
}

// acceptConnectionWithContext accepts a connection on the specified port with context support
func (s *Server) acceptConnectionWithContext(ctx context.Context, port int, playerType string) (*Connection, string, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, "", fmt.Errorf("failed to listen on port %d: %w", port, err)
	}
	defer listener.Close()

	log.Printf("Waiting for %s player on port %d...", playerType, port)

	// タイムアウト設定（60秒）
	_ = listener.(*net.TCPListener).SetDeadline(time.Now().Add(60 * time.Second))

	// contextのキャンセルを監視
	acceptChan := make(chan net.Conn, 1)
	acceptErrChan := make(chan error, 1)

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			acceptErrChan <- err
			return
		}
		acceptChan <- conn
	}()

	var conn net.Conn
	select {
	case <-ctx.Done():
		return nil, "", ctx.Err()
	case err := <-acceptErrChan:
		return nil, "", fmt.Errorf("failed to accept connection: %w", err)
	case conn = <-acceptChan:
		// 成功
	}

	log.Printf("%s player connected from %s", playerType, conn.RemoteAddr())

	connection := NewConnection(conn)

	// プレイヤー名を受信
	name, err := connection.Receive()
	if err != nil {
		connection.Close()
		return nil, "", fmt.Errorf("failed to receive player name: %w", err)
	}

	return connection, name, nil
}

// runGame runs the main game loop
func (s *Server) runGame() error {
	defer s.HotConn.Close()
	defer s.CoolConn.Close()
	defer s.DumpSystem.Close()

	for s.Board.Turn < s.Board.MaxTurns && !s.Board.GameOver {
		// ターン番号に応じて先後を決定（偶数: Hot先攻、奇数: Cool先攻）
		if s.Board.Turn%2 == 0 {
			// Hot先攻
			if err := s.processTurn(s.HotConn, s.Board.Hot, s.CoolConn, s.Board.Cool); err != nil {
				log.Printf("Hot turn error: %v", err)
				s.Board.Hot.IsAlive = false
				s.Board.GameOver = true
				break
			}

			if s.Board.GameOver {
				break
			}

			// Cool後攻
			if err := s.processTurn(s.CoolConn, s.Board.Cool, s.HotConn, s.Board.Hot); err != nil {
				log.Printf("Cool turn error: %v", err)
				s.Board.Cool.IsAlive = false
				s.Board.GameOver = true
				break
			}
		} else {
			// Cool先攻
			if err := s.processTurn(s.CoolConn, s.Board.Cool, s.HotConn, s.Board.Hot); err != nil {
				log.Printf("Cool turn error: %v", err)
				s.Board.Cool.IsAlive = false
				s.Board.GameOver = true
				break
			}

			if s.Board.GameOver {
				break
			}

			// Hot後攻
			if err := s.processTurn(s.HotConn, s.Board.Hot, s.CoolConn, s.Board.Cool); err != nil {
				log.Printf("Hot turn error: %v", err)
				s.Board.Hot.IsAlive = false
				s.Board.GameOver = true
				break
			}
		}

		// ターンを進める
		s.Board.IncrementTurn()

		// 状態をダンプに記録
		if err := s.DumpSystem.Action(s.Board); err != nil {
			log.Printf("Warning: failed to write action to dump: %v", err)
		}
	}

	// ゲーム終了処理
	return s.endGame()
}

// processTurn processes one player's turn
func (s *Server) processTurn(conn *Connection, char *Character, opponentConn *Connection, opponent *Character) error {
	// "Ready\n" 送信（CHaserプロトコル）
	if err := conn.Send("Ready\n"); err != nil {
		return fmt.Errorf("failed to send ready: %w", err)
	}

	// "gr" 受信待機（準備完了確認）
	if err := conn.WaitForReady(); err != nil {
		return fmt.Errorf("failed to receive ready: %w", err)
	}

	// レスポンスを生成して送信（Ready専用）
	var readyResponse [10]int
	if s.Board.GameOver {
		readyResponse[0] = 0
	} else {
		readyResponse[0] = 1
	}
	// 周囲9マスの情報
	directions := []struct {
		dy, dx int
		index  int
	}{
		{-1, -1, 1}, // 左上
		{-1, 0, 2},  // 上
		{-1, 1, 3},  // 右上
		{0, -1, 4},  // 左
		{0, 0, 5},   // 中央
		{0, 1, 6},   // 右
		{1, -1, 7},  // 左下
		{1, 0, 8},   // 下
		{1, 1, 9},   // 右下
	}
	for _, d := range directions {
		pos := Position{
			Y: char.Position.Y + d.dy,
			X: char.Position.X + d.dx,
		}
		if pos == opponent.Position {
			readyResponse[d.index] = 1 // 敵
		} else {
			readyResponse[d.index] = int(s.Board.GetCell(pos))
		}
	}

	if err := conn.SendResponse(readyResponse); err != nil {
		return fmt.Errorf("failed to send ready response: %w", err)
	}

	// ゲームオーバーなら終了
	if s.Board.GameOver {
		return nil
	}

	// 行動データ受信
	actionStr, err := conn.ReceiveAction()
	if err != nil {
		return fmt.Errorf("failed to receive action: %w", err)
	}

	// アクションをパース
	action, direction, err := ParseAction(actionStr)
	if err != nil {
		return fmt.Errorf("failed to parse action: %w", err)
	}

	log.Printf("%s: %s %d (Turn %d)", char.Name, action, direction, s.Board.Turn)

	// アクションを実行してレスポンスを生成
	var response [10]int

	switch action {
	case "wk": // walk
		if err := s.Board.Walk(char, direction); err != nil {
			log.Printf("%s walk failed: %v", char.Name, err)
		}
		response = BuildWalkResponse(char, opponent, s.Board)

	case "lk": // look
		response = BuildLookResponse(char, opponent, s.Board, direction)

	case "sc": // search
		response = BuildSearchResponse(char, opponent, s.Board, direction)

	case "pt": // put
		s.Board.Put(char.Position, direction)
		response = BuildPutResponse(char, opponent, s.Board)
	}

	// レスポンスを送信
	if err := conn.SendResponse(response); err != nil {
		return fmt.Errorf("failed to send response: %w", err)
	}

	// 確認応答("#\r\n")を受信
	ack, err := conn.Receive()
	if err != nil {
		return fmt.Errorf("failed to receive acknowledgment: %w", err)
	}
	if ack != "#" {
		log.Printf("Warning: expected '#' acknowledgment, got '%s'", ack)
	}

	return nil
}

// endGame handles game end
func (s *Server) endGame() error {
	log.Println("Game Over!")

	// 勝敗判定
	winner, reason := s.Board.GetResult()

	if winner == nil {
		log.Printf("Result: Draw - %s", reason)
		log.Printf("Hot: %d items, Cool: %d items", s.Board.Hot.Items, s.Board.Cool.Items)

		// 引き分けをダンプに記録
		if err := s.DumpSystem.Result(nil, nil, reason); err != nil {
			log.Printf("Warning: failed to write result to dump: %v", err)
		}
	} else {
		loser := s.Board.GetOpponent(winner)
		log.Printf("Result: %s wins! - %s", winner.Name, reason)
		log.Printf("Hot: %d items, Cool: %d items", s.Board.Hot.Items, s.Board.Cool.Items)

		// 勝者をダンプに記録
		if err := s.DumpSystem.Result(winner, loser, reason); err != nil {
			log.Printf("Warning: failed to write result to dump: %v", err)
		}
	}

	// ゲームオーバー信号を送信
	if s.HotConn != nil {
		_ = s.HotConn.SendGameOver()
	}
	if s.CoolConn != nil {
		_ = s.CoolConn.SendGameOver()
	}

	return nil
}
