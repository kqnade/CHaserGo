package server

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

// Connection represents a client connection
type Connection struct {
	conn   net.Conn
	reader *bufio.Reader
}

// NewConnection creates a new connection wrapper
func NewConnection(conn net.Conn) *Connection {
	return &Connection{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}
}

// Send sends a string message to the client
func (c *Connection) Send(message string) error {
	if c.conn == nil {
		return fmt.Errorf("connection is closed")
	}

	_, err := c.conn.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// Receive receives a message from the client
func (c *Connection) Receive() (string, error) {
	if c.conn == nil {
		return "", fmt.Errorf("connection is closed")
	}

	// タイムアウト設定（10秒）
	_ = c.conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	data, err := c.reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to receive message: %w", err)
	}

	return strings.TrimSpace(data), nil
}

// Close closes the connection
func (c *Connection) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// WaitForReady waits for "gr" (get ready) command
func (c *Connection) WaitForReady() error {
	msg, err := c.Receive()
	if err != nil {
		return err
	}

	if msg != "gr" {
		return fmt.Errorf("expected 'gr', got '%s'", msg)
	}

	return nil
}

// ReceiveAction receives an action command from the client
func (c *Connection) ReceiveAction() (string, error) {
	msg, err := c.Receive()
	if err != nil {
		return "", err
	}

	return msg, nil
}

// SendStart sends the start signal "@"
func (c *Connection) SendStart() error {
	return c.Send("@\n")
}

// SendGameOver sends the game over signal
func (c *Connection) SendGameOver() error {
	return c.Send("#\n")
}

// SendResponse sends the 10-value response
func (c *Connection) SendResponse(values [10]int) error {
	// CHaserプロトコル: スペースなしで10桁連続
	var message string
	for _, v := range values {
		message += fmt.Sprintf("%d", v)
	}
	message += "\n"
	return c.Send(message)
}

// ParseAction parses an action command
func ParseAction(command string) (action string, direction Direction, err error) {
	parts := strings.Split(command, " ")
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("invalid command format: %s", command)
	}

	action = parts[0]

	switch parts[1] {
	case "0":
		direction = Up
	case "1":
		direction = Down
	case "2":
		direction = Left
	case "3":
		direction = Right
	default:
		return "", 0, fmt.Errorf("invalid direction: %s", parts[1])
	}

	// アクションの妥当性チェック
	validActions := map[string]bool{
		"wk": true, // walk
		"lk": true, // look
		"sc": true, // search
		"pt": true, // put
	}

	if !validActions[action] {
		return "", 0, fmt.Errorf("invalid action: %s", action)
	}

	return action, direction, nil
}

// BuildLookResponse builds a response for look command
func BuildLookResponse(char *Character, opponent *Character, board *Board, dir Direction) [10]int {
	var values [10]int

	// 値0: 制御フラグ（1=継続、0=ゲームオーバー）
	if board.GameOver {
		values[0] = 0
	} else {
		values[0] = 1
	}

	// Look: 2マス先の情報
	cell := board.Look(char.Position, dir)

	// 2マス先に相手がいるかチェック
	targetPos := board.Move(char.Position, dir)
	targetPos = board.Move(targetPos, dir)

	if opponent.Position == targetPos {
		values[2] = 1 // 敵
	} else {
		values[2] = int(cell)
	}

	return values
}

// BuildSearchResponse builds a response for search command
func BuildSearchResponse(char *Character, opponent *Character, board *Board, dir Direction) [10]int {
	var values [10]int

	// 値0: 制御フラグ
	if board.GameOver {
		values[0] = 0
	} else {
		values[0] = 1
	}

	// Search: 直線9マスの情報
	cells := board.Search(char.Position, dir)

	for i := 0; i < 9; i++ {
		values[i+1] = int(cells[i])

		// 相手の位置チェック
		checkPos := char.Position
		for j := 0; j <= i; j++ {
			checkPos = board.Move(checkPos, dir)
		}

		if opponent.Position == checkPos {
			values[i+1] = 1 // 敵
		}
	}

	return values
}

// BuildWalkResponse builds a response for walk command
func BuildWalkResponse(char *Character, opponent *Character, board *Board) [10]int {
	var values [10]int

	// 値0: 制御フラグ
	if board.GameOver {
		values[0] = 0
	} else {
		values[0] = 1
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
			values[d.index] = 1 // 敵
		} else {
			values[d.index] = int(board.GetCell(pos))
		}
	}

	return values
}

// BuildPutResponse builds a response for put command
func BuildPutResponse(char *Character, opponent *Character, board *Board) [10]int {
	var values [10]int

	// 値0: 制御フラグ
	if board.GameOver {
		values[0] = 0
	} else {
		values[0] = 1
	}

	// 周囲9マスの情報（walkと同じ）
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
			values[d.index] = 1 // 敵
		} else {
			values[d.index] = int(board.GetCell(pos))
		}
	}

	return values
}
