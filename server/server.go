package server

import (
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
	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	wg.Add(2)

	// Hot接続待機
	go func() {
		defer wg.Done()
		conn, name, err := s.acceptConnection(s.HotPort, "Hot")
		if err != nil {
			errChan <- fmt.Errorf("hot connection failed: %w", err)
			return
		}
		s.HotConn = conn
		s.Board.Hot.Name = name
		log.Printf("Hot player connected: %s", name)
	}()

	// Cool接続待機
	go func() {
		defer wg.Done()
		conn, name, err := s.acceptConnection(s.CoolPort, "Cool")
		if err != nil {
			errChan <- fmt.Errorf("cool connection failed: %w", err)
			return
		}
		s.CoolConn = conn
		s.Board.Cool.Name = name
		log.Printf("Cool player connected: %s", name)
	}()

	wg.Wait()
	close(errChan)

	// エラーチェック
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	// プレイヤー名をダンプに記録
	if err := s.DumpSystem.SetNames(s.Board.Hot.Name, s.Board.Cool.Name); err != nil {
		log.Printf("Warning: failed to write names to dump: %v", err)
	}

	log.Println("Both players connected. Starting game...")

	// ゲームメインループ
	return s.runGame()
}

// acceptConnection accepts a connection on the specified port
func (s *Server) acceptConnection(port int, playerType string) (*Connection, string, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, "", fmt.Errorf("failed to listen on port %d: %w", port, err)
	}
	defer listener.Close()

	log.Printf("Waiting for %s player on port %d...", playerType, port)

	// タイムアウト設定（60秒）
	_ = listener.(*net.TCPListener).SetDeadline(time.Now().Add(60 * time.Second))

	conn, err := listener.Accept()
	if err != nil {
		return nil, "", fmt.Errorf("failed to accept connection: %w", err)
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
		// ターン番号に応じて先後を決定（奇数: Hot先攻、偶数: Cool先攻）
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
	// "@" 送信（開始合図）
	if err := conn.SendStart(); err != nil {
		return fmt.Errorf("failed to send start: %w", err)
	}

	// "gr" 受信待機（準備完了確認）
	if err := conn.WaitForReady(); err != nil {
		return fmt.Errorf("failed to receive ready: %w", err)
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
	} else {
		loser := s.Board.GetOpponent(winner)
		log.Printf("Result: %s wins! - %s", winner.Name, reason)
		log.Printf("Hot: %d items, Cool: %d items", s.Board.Hot.Items, s.Board.Cool.Items)

		// ダンプに記録
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
