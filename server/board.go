package server

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// CellType represents a cell on the board
type CellType int

const (
	Empty CellType = 0
	Wall  CellType = 2
	Item  CellType = 3
)

// Direction represents movement direction
type Direction int

const (
	Up Direction = iota
	Down
	Left
	Right
)

// Position represents a coordinate on the board
type Position struct {
	X int
	Y int
}

// Character represents a player character
type Character struct {
	Name     string
	Position Position
	Items    int
	IsAlive  bool
}

// Board manages the game state
type Board struct {
	MapData   [][]CellType
	Width     int
	Height    int
	MaxTurns  int
	Hot       *Character
	Cool      *Character
	Turn      int
	GameOver  bool
	mapPath   string
}

// NewBoard creates a new board from a map file
func NewBoard(mapPath string) (*Board, error) {
	file, err := os.Open(mapPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open map file: %w", err)
	}
	defer file.Close()

	board := &Board{
		mapPath: mapPath,
		Hot:     &Character{IsAlive: true},
		Cool:    &Character{IsAlive: true},
		Turn:    0,
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 2 {
			continue
		}

		prefix := line[:2]
		data := strings.TrimSpace(line[2:])

		switch prefix {
		case "N ":
			// N行は未使用
			continue
		case "T ":
			// ターン数
			board.MaxTurns, err = strconv.Atoi(data)
			if err != nil {
				return nil, fmt.Errorf("invalid turn count: %w", err)
			}
		case "S ":
			// サイズ（座標系を反転）
			parts := strings.Split(data, ",")
			if len(parts) != 2 {
				return nil, errors.New("invalid size format")
			}
			board.Height, err = strconv.Atoi(parts[0])
			if err != nil {
				return nil, fmt.Errorf("invalid height: %w", err)
			}
			board.Width, err = strconv.Atoi(parts[1])
			if err != nil {
				return nil, fmt.Errorf("invalid width: %w", err)
			}
			// マップデータ初期化
			board.MapData = make([][]CellType, board.Height)
			for i := range board.MapData {
				board.MapData[i] = make([]CellType, board.Width)
			}
		case "D ":
			// マップデータ
			parts := strings.Split(data, ",")
			if len(parts) < 3 {
				return nil, errors.New("invalid map data format")
			}
			y, err := strconv.Atoi(parts[0])
			if err != nil {
				return nil, fmt.Errorf("invalid y coordinate: %w", err)
			}
			x, err := strconv.Atoi(parts[1])
			if err != nil {
				return nil, fmt.Errorf("invalid x coordinate: %w", err)
			}
			value, err := strconv.Atoi(parts[2])
			if err != nil {
				return nil, fmt.Errorf("invalid cell value: %w", err)
			}
			if y >= 0 && y < board.Height && x >= 0 && x < board.Width {
				board.MapData[y][x] = CellType(value)
			}
		case "H ":
			// Hot(先攻)の初期位置
			parts := strings.Split(data, ",")
			if len(parts) != 2 {
				return nil, errors.New("invalid hot position format")
			}
			board.Hot.Position.Y, err = strconv.Atoi(parts[0])
			if err != nil {
				return nil, fmt.Errorf("invalid hot y: %w", err)
			}
			board.Hot.Position.X, err = strconv.Atoi(parts[1])
			if err != nil {
				return nil, fmt.Errorf("invalid hot x: %w", err)
			}
		case "C ":
			// Cool(後攻)の初期位置
			parts := strings.Split(data, ",")
			if len(parts) != 2 {
				return nil, errors.New("invalid cool position format")
			}
			board.Cool.Position.Y, err = strconv.Atoi(parts[0])
			if err != nil {
				return nil, fmt.Errorf("invalid cool y: %w", err)
			}
			board.Cool.Position.X, err = strconv.Atoi(parts[1])
			if err != nil {
				return nil, fmt.Errorf("invalid cool x: %w", err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading map file: %w", err)
	}

	return board, nil
}

// GetCell returns the cell type at the given position
func (b *Board) GetCell(pos Position) CellType {
	if pos.Y < 0 || pos.Y >= b.Height || pos.X < 0 || pos.X >= b.Width {
		return Wall // 境界外は壁扱い
	}
	return b.MapData[pos.Y][pos.X]
}

// SetCell sets the cell type at the given position
func (b *Board) SetCell(pos Position, cell CellType) {
	if pos.Y >= 0 && pos.Y < b.Height && pos.X >= 0 && pos.X < b.Width {
		b.MapData[pos.Y][pos.X] = cell
	}
}

// GetOpponent returns the opponent character
func (b *Board) GetOpponent(char *Character) *Character {
	if char == b.Hot {
		return b.Cool
	}
	return b.Hot
}

// Move calculates the new position based on direction
func (b *Board) Move(pos Position, dir Direction) Position {
	newPos := pos
	switch dir {
	case Up:
		newPos.Y--
	case Down:
		newPos.Y++
	case Left:
		newPos.X--
	case Right:
		newPos.X++
	}
	return newPos
}

// Walk moves the character and collects items
func (b *Board) Walk(char *Character, dir Direction) error {
	newPos := b.Move(char.Position, dir)

	// 壁または境界外チェック
	if b.GetCell(newPos) == Wall {
		char.IsAlive = false
		b.GameOver = true
		return errors.New("hit wall")
	}

	// アイテム収集
	if b.GetCell(newPos) == Item {
		char.Items++
		b.SetCell(newPos, Empty)
	}

	// 移動
	char.Position = newPos

	// 四方を壁で囲まれたかチェック
	if b.IsSurrounded(char) {
		char.IsAlive = false
		b.GameOver = true
		return errors.New("surrounded by walls")
	}

	return nil
}

// Look returns the cell 2 steps ahead in the given direction
func (b *Board) Look(pos Position, dir Direction) CellType {
	newPos := b.Move(pos, dir)
	newPos = b.Move(newPos, dir)
	return b.GetCell(newPos)
}

// Search returns cells in a straight line (up to 9 cells)
func (b *Board) Search(pos Position, dir Direction) [9]CellType {
	var result [9]CellType
	currentPos := pos

	for i := 0; i < 9; i++ {
		currentPos = b.Move(currentPos, dir)
		result[i] = b.GetCell(currentPos)
	}

	return result
}

// Put places a wall in the given direction
func (b *Board) Put(pos Position, dir Direction) {
	newPos := b.Move(pos, dir)
	if b.GetCell(newPos) == Empty {
		b.SetCell(newPos, Wall)
	}
}

// IsSurrounded checks if the character is surrounded by walls
func (b *Board) IsSurrounded(char *Character) bool {
	directions := []Direction{Up, Down, Left, Right}
	for _, dir := range directions {
		newPos := b.Move(char.Position, dir)
		if b.GetCell(newPos) != Wall {
			return false
		}
	}
	return true
}

// GetResult determines the winner based on items collected
func (b *Board) GetResult() (winner *Character, reason string) {
	if !b.Hot.IsAlive && !b.Cool.IsAlive {
		if b.Hot.Items > b.Cool.Items {
			return b.Hot, "both died, hot has more items"
		} else if b.Cool.Items > b.Hot.Items {
			return b.Cool, "both died, cool has more items"
		}
		return nil, "draw - both died with same items"
	}

	if !b.Hot.IsAlive {
		return b.Cool, "hot died"
	}
	if !b.Cool.IsAlive {
		return b.Hot, "cool died"
	}

	// 通常の勝敗判定（アイテム数）
	if b.Hot.Items > b.Cool.Items {
		return b.Hot, "hot has more items"
	} else if b.Cool.Items > b.Hot.Items {
		return b.Cool, "cool has more items"
	}

	return nil, "draw"
}

// IncrementTurn increments the turn counter
func (b *Board) IncrementTurn() {
	b.Turn++
	if b.Turn >= b.MaxTurns {
		b.GameOver = true
	}
}
