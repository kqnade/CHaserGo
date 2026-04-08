package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

// Server represents the CHaser game server
type Server struct {
	config     ServerConfig
	Board      *Board
	DumpSystem *DumpSystem
	HotConn    *Connection
	CoolConn   *Connection
	snapshotCh chan BoardSnapshot
	revision   uint64
}

// ServerConfig holds server configuration
type ServerConfig struct {
	MapPath    string
	HotPort    int
	CoolPort   int
	DumpPath   string
	EnableDump bool
	SnapshotCh chan BoardSnapshot // nil = no-op
}

// NewServer creates a new CHaser server
func NewServer(config ServerConfig) (*Server, error) {
	board, err := NewBoard(config.MapPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load board: %w", err)
	}

	dumpSystem, err := NewDumpSystem(config.DumpPath, config.MapPath, config.EnableDump)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize dump system: %w", err)
	}

	s := &Server{
		config:     config,
		Board:      board,
		DumpSystem: dumpSystem,
		snapshotCh: config.SnapshotCh,
	}

	s.publishSnapshot(KindInitial, TurnStepFirst, PhaseWaiting, "", "")
	return s, nil
}

// Start starts the server and waits for connections
func (s *Server) Start(ctx context.Context) error {
	log.Printf("Starting CHaser server...")
	log.Printf("Hot port: %d, Cool port: %d", s.config.HotPort, s.config.CoolPort)
	log.Printf("Max turns: %d", s.Board.MaxTurns)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	errChan := make(chan error, 2)
	doneChan := make(chan struct{})

	wg.Add(2)

	go func() {
		defer wg.Done()
		conn, name, err := s.acceptConnectionWithContext(ctx, s.config.HotPort, "Hot")
		if err != nil {
			errChan <- fmt.Errorf("hot connection failed: %w", err)
			return
		}
		s.HotConn = conn
		s.Board.Hot.Name = name
		log.Printf("Hot player connected: %s", name)
	}()

	go func() {
		defer wg.Done()
		conn, name, err := s.acceptConnectionWithContext(ctx, s.config.CoolPort, "Cool")
		if err != nil {
			errChan <- fmt.Errorf("cool connection failed: %w", err)
			return
		}
		s.CoolConn = conn
		s.Board.Cool.Name = name
		log.Printf("Cool player connected: %s", name)
	}()

	go func() {
		wg.Wait()
		close(doneChan)
	}()

	select {
	case err := <-errChan:
		cancel() // stop the other goroutine
		<-doneChan
		s.publishSnapshot(KindError, TurnStepFirst, PhaseError, "", err.Error())
		return err
	case <-doneChan:
	}

	if err := s.DumpSystem.SetNames(s.Board.Hot.Name, s.Board.Cool.Name); err != nil {
		log.Printf("Warning: failed to write names to dump: %v", err)
	}

	log.Println("Both players connected. Starting game...")
	s.publishSnapshot(KindConnected, TurnStepFirst, PhaseRunning, "", "")

	return s.runGame(ctx)
}

// acceptConnectionWithContext accepts a connection on the specified port with context support
func (s *Server) acceptConnectionWithContext(ctx context.Context, port int, playerType string) (*Connection, string, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, "", fmt.Errorf("failed to listen on port %d: %w", port, err)
	}
	defer listener.Close()

	log.Printf("Waiting for %s player on port %d...", playerType, port)

	_ = listener.(*net.TCPListener).SetDeadline(time.Now().Add(60 * time.Second))

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
	}

	log.Printf("%s player connected from %s", playerType, conn.RemoteAddr())

	connection := NewConnection(conn)

	name, err := connection.ReceiveContext(ctx)
	if err != nil {
		connection.Close()
		return nil, "", fmt.Errorf("failed to receive player name: %w", err)
	}

	return connection, name, nil
}

// turnActor はターンの先攻/後攻情報をまとめた型
type turnActor struct {
	conn     *Connection
	self     *Character
	opponent *Character
	step     TurnStep
}

// actorsForTurn はターン番号に応じた先攻/後攻の順序を返す
func (s *Server) actorsForTurn(turn int) [2]turnActor {
	if turn%2 == 0 {
		return [2]turnActor{
			{s.HotConn, s.Board.Hot, s.Board.Cool, TurnStepFirst},
			{s.CoolConn, s.Board.Cool, s.Board.Hot, TurnStepSecond},
		}
	}
	return [2]turnActor{
		{s.CoolConn, s.Board.Cool, s.Board.Hot, TurnStepFirst},
		{s.HotConn, s.Board.Hot, s.Board.Cool, TurnStepSecond},
	}
}

// runGame runs the main game loop
func (s *Server) runGame(ctx context.Context) error {
	defer s.HotConn.Close()
	defer s.CoolConn.Close()
	defer s.DumpSystem.Close()

	for s.Board.Turn < s.Board.MaxTurns && !s.Board.GameOver {
		actors := s.actorsForTurn(s.Board.Turn)
		for _, a := range actors {
			err := s.processTurn(ctx, a.conn, a.self, a.opponent)
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return err
				}
				log.Printf("%s turn error: %v", a.self.Name, err)
				a.self.IsAlive = false
				s.Board.GameOver = true
			}

			// ActionEnd: GameOver経路を含め毎回発火
			s.publishSnapshot(KindActionEnd, a.step, PhaseRunning, "", "")
			if s.Board.GameOver {
				break
			}
		}

		if !s.Board.GameOver {
			s.Board.IncrementTurn()
			s.publishSnapshot(KindTurnEnd, TurnStepSecond, PhaseRunning, "", "")
			if err := s.DumpSystem.Action(s.Board); err != nil {
				log.Printf("Warning: failed to write action to dump: %v", err)
			}
		}
	}

	return s.endGame()
}

// processTurn processes one player's turn
func (s *Server) processTurn(ctx context.Context, conn *Connection, char *Character, opponent *Character) error {
	if err := conn.Send("Ready\n"); err != nil {
		return fmt.Errorf("failed to send ready: %w", err)
	}

	if err := conn.WaitForReadyContext(ctx); err != nil {
		return fmt.Errorf("failed to receive ready: %w", err)
	}

	// Ready レスポンス（周辺9マス）生成・送信
	var readyResponse [10]int
	if s.Board.GameOver {
		readyResponse[0] = 0
	} else {
		readyResponse[0] = 1
	}
	directions := []struct {
		dy, dx int
		index  int
	}{
		{-1, -1, 1}, {-1, 0, 2}, {-1, 1, 3},
		{0, -1, 4}, {0, 0, 5}, {0, 1, 6},
		{1, -1, 7}, {1, 0, 8}, {1, 1, 9},
	}
	for _, d := range directions {
		pos := Position{Y: char.Position.Y + d.dy, X: char.Position.X + d.dx}
		if pos == opponent.Position {
			readyResponse[d.index] = 1
		} else {
			readyResponse[d.index] = int(s.Board.GetCell(pos))
		}
	}

	if err := conn.SendResponse(readyResponse); err != nil {
		return fmt.Errorf("failed to send ready response: %w", err)
	}

	if s.Board.GameOver {
		return nil
	}

	// 行動受信
	actionStr, err := conn.ReceiveActionContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to receive action: %w", err)
	}

	action, direction, err := ParseAction(actionStr)
	if err != nil {
		return fmt.Errorf("failed to parse action: %w", err)
	}

	log.Printf("%s: %s %d (Turn %d)", char.Name, action, direction, s.Board.Turn)

	var response [10]int
	switch action {
	case "wk":
		if err := s.Board.Walk(char, direction); err != nil {
			log.Printf("%s walk failed: %v", char.Name, err)
		}
		response = BuildWalkResponse(char, opponent, s.Board)
	case "lk":
		response = BuildLookResponse(char, opponent, s.Board, direction)
	case "sc":
		response = BuildSearchResponse(char, opponent, s.Board, direction)
	case "pt":
		s.Board.Put(char.Position, direction)
		response = BuildPutResponse(char, opponent, s.Board)
	}

	if err := conn.SendResponse(response); err != nil {
		return fmt.Errorf("failed to send response: %w", err)
	}

	// '#' 確認応答受信
	ack, err := conn.ReceiveContext(ctx)
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

	winner, reason := s.Board.GetResult()

	var winnerName string
	if winner == nil {
		log.Printf("Result: Draw - %s", reason)
		log.Printf("Hot: %d items, Cool: %d items", s.Board.Hot.Items, s.Board.Cool.Items)
		if err := s.DumpSystem.Result(nil, nil, reason); err != nil {
			log.Printf("Warning: failed to write result to dump: %v", err)
		}
	} else {
		winnerName = winner.Name
		loser := s.Board.GetOpponent(winner)
		log.Printf("Result: %s wins! - %s", winner.Name, reason)
		log.Printf("Hot: %d items, Cool: %d items", s.Board.Hot.Items, s.Board.Cool.Items)
		if err := s.DumpSystem.Result(winner, loser, reason); err != nil {
			log.Printf("Warning: failed to write result to dump: %v", err)
		}
	}

	s.publishSnapshot(KindGameOver, TurnStepFirst, PhaseGameOver, winnerName, reason)

	if s.HotConn != nil {
		_ = s.HotConn.SendGameOver()
	}
	if s.CoolConn != nil {
		_ = s.CoolConn.SendGameOver()
	}

	return nil
}

// publishSnapshot はスナップショットを snapshotCh に non-blocking で送信する
// snapshotCh が nil の場合は no-op
func (s *Server) publishSnapshot(kind SnapshotKind, step TurnStep, phase SnapshotPublicPhase, winner, reason string) {
	if s.snapshotCh == nil {
		return
	}
	s.revision++
	snap := SnapshotFromBoard(s.Board, kind, step, phase, s.revision, winner, reason)
	select {
	case s.snapshotCh <- snap:
	default:
		// channel full: drain old snapshot and send new (overwrite semantics)
		select {
		case <-s.snapshotCh:
		default:
		}
		s.snapshotCh <- snap
	}
}
