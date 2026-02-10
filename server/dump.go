package server

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// DumpSystem records game history
type DumpSystem struct {
	enabled  bool
	filePath string
	file     *os.File
	writer   *bufio.Writer
	mapData  []string
}

// NewDumpSystem creates a new dump system
func NewDumpSystem(filePath string, mapPath string, enabled bool) (*DumpSystem, error) {
	if !enabled {
		return &DumpSystem{enabled: false}, nil
	}

	// マップデータを読み込む
	mapData, err := readMapData(mapPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read map data: %w", err)
	}

	// ダンプファイルを開く
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create dump file: %w", err)
	}

	return &DumpSystem{
		enabled:  true,
		filePath: filePath,
		file:     file,
		writer:   bufio.NewWriter(file),
		mapData:  mapData,
	}, nil
}

// readMapData reads map data from file
func readMapData(mapPath string) ([]string, error) {
	file, err := os.Open(mapPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) >= 2 {
			// 先頭2文字を削除（"D ", "S " などのプレフィックス）
			lines = append(lines, line[2:])
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

// SetNames writes the player names to the dump file
func (d *DumpSystem) SetNames(hotName, coolName string) error {
	if !d.enabled {
		return nil
	}

	// プレイヤー名を書き込み
	_, err := d.writer.WriteString(fmt.Sprintf("%s,%s\n", hotName, coolName))
	if err != nil {
		return err
	}

	// マップデータを書き込み
	for _, line := range d.mapData {
		_, err := d.writer.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}

	// 初期スコア
	_, err = d.writer.WriteString("0,0\n")
	if err != nil {
		return err
	}

	return d.writer.Flush()
}

// Action records the current game state
func (d *DumpSystem) Action(board *Board) error {
	if !d.enabled {
		return nil
	}

	// 盤面データを書き込み
	for y := 0; y < board.Height; y++ {
		var row []string
		for x := 0; x < board.Width; x++ {
			row = append(row, fmt.Sprintf("%d", board.MapData[y][x]))
		}
		_, err := d.writer.WriteString(strings.Join(row, ",") + "\n")
		if err != nil {
			return err
		}
	}

	// Hotの位置
	_, err := d.writer.WriteString(fmt.Sprintf("%d,%d\n", board.Hot.Position.Y, board.Hot.Position.X))
	if err != nil {
		return err
	}

	// Coolの位置
	_, err = d.writer.WriteString(fmt.Sprintf("%d,%d\n", board.Cool.Position.Y, board.Cool.Position.X))
	if err != nil {
		return err
	}

	// アイテム数
	_, err = d.writer.WriteString(fmt.Sprintf("%d,%d\n", board.Hot.Items, board.Cool.Items))
	if err != nil {
		return err
	}

	return d.writer.Flush()
}

// Result records the game result
func (d *DumpSystem) Result(winner *Character, loser *Character, reason string) error {
	if !d.enabled {
		return nil
	}

	// ゲーム終了マーカー
	_, err := d.writer.WriteString("gameend\n")
	if err != nil {
		return err
	}

	// 勝敗情報
	if winner == nil {
		// 引き分け
		_, err = d.writer.WriteString(fmt.Sprintf("draw,draw,%s\n", reason))
	} else {
		_, err = d.writer.WriteString(fmt.Sprintf("%s,win,%s\n", winner.Name, reason))
	}

	if err != nil {
		return err
	}

	return d.writer.Flush()
}

// Close closes the dump file
func (d *DumpSystem) Close() error {
	if !d.enabled || d.file == nil {
		return nil
	}

	if err := d.writer.Flush(); err != nil {
		return err
	}

	return d.file.Close()
}
