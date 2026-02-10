package mapgen

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

// CellType represents a cell on the map
type CellType int

const (
	Empty CellType = 0
	Block CellType = 2
	Item  CellType = 3
)

// Position represents a coordinate
type Position struct {
	X int
	Y int
}

// Map represents a game map
type Map struct {
	Width  int
	Height int
	Data   [][]CellType
	Hot    Position
	Cool   Position
	Turns  int
}

// Generator generates CHaser maps
type Generator struct {
	rng *rand.Rand
}

// NewGenerator creates a new map generator
func NewGenerator() *Generator {
	return &Generator{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// NewGeneratorWithSeed creates a generator with a specific seed
func NewGeneratorWithSeed(seed int64) *Generator {
	return &Generator{
		rng: rand.New(rand.NewSource(seed)),
	}
}

// GenerateMap generates a random map
func (g *Generator) GenerateMap(maxBlocks, maxItems int) *Map {
	// 小マップ（7×8）を生成
	smallMap := g.generateSmallMap(7, 8, maxBlocks, maxItems)

	// 4つの回転バリエーションを作成
	map0 := smallMap
	map1 := g.rotateMap(map0)
	map2 := g.rotateMap(map1)
	map3 := g.rotateMap(map2)

	// 隙間用マップ（8×2）を生成
	gapMap1 := g.generateGapMap(8, 2)
	gapMap2 := g.generateGapMap(8, 2)

	// 大マップ（15×17）に結合
	largeMap := g.jointMaps(map0, map1, map2, map3, gapMap1, gapMap2)

	// エージェントを対角配置
	g.placeAgents(largeMap)

	largeMap.Turns = 120

	return largeMap
}

// generateSmallMap generates a small map with random blocks and items
func (g *Generator) generateSmallMap(width, height, maxBlocks, maxItems int) *Map {
	m := &Map{
		Width:  width,
		Height: height,
		Data:   make([][]CellType, height),
	}

	for y := 0; y < height; y++ {
		m.Data[y] = make([]CellType, width)
	}

	// ブロックをランダム配置（最外周を除く）
	blockCount := g.rng.Intn(maxBlocks + 1)
	for i := 0; i < blockCount; i++ {
		for attempts := 0; attempts < 100; attempts++ {
			x := g.rng.Intn(width-2) + 1
			y := g.rng.Intn(height-2) + 1

			if m.Data[y][x] == Empty {
				m.Data[y][x] = Block
				break
			}
		}
	}

	// アイテムをランダム配置
	itemCount := g.rng.Intn(maxItems + 1)
	for i := 0; i < itemCount; i++ {
		for attempts := 0; attempts < 100; attempts++ {
			x := g.rng.Intn(width)
			y := g.rng.Intn(height)

			if m.Data[y][x] == Empty {
				m.Data[y][x] = Item
				break
			}
		}
	}

	return m
}

// generateGapMap generates a gap map
func (g *Generator) generateGapMap(width, height int) *Map {
	m := &Map{
		Width:  width,
		Height: height,
		Data:   make([][]CellType, height),
	}

	for y := 0; y < height; y++ {
		m.Data[y] = make([]CellType, width)
	}

	return m
}

// rotateMap rotates a map 90 degrees clockwise
func (g *Generator) rotateMap(m *Map) *Map {
	// 転置（transpose）
	transposed := &Map{
		Width:  m.Height,
		Height: m.Width,
		Data:   make([][]CellType, m.Width),
	}

	for y := 0; y < transposed.Height; y++ {
		transposed.Data[y] = make([]CellType, transposed.Width)
		for x := 0; x < transposed.Width; x++ {
			transposed.Data[y][x] = m.Data[x][y]
		}
	}

	// 上下反転（vertical flip）
	rotated := &Map{
		Width:  transposed.Width,
		Height: transposed.Height,
		Data:   make([][]CellType, transposed.Height),
	}

	for y := 0; y < rotated.Height; y++ {
		rotated.Data[y] = make([]CellType, rotated.Width)
		for x := 0; x < rotated.Width; x++ {
			rotated.Data[y][x] = transposed.Data[transposed.Height-1-y][x]
		}
	}

	return rotated
}

// jointMaps combines 4 small maps and 2 gap maps into a large map
func (g *Generator) jointMaps(m0, m1, m2, m3, gap1, gap2 *Map) *Map {
	// 15×17の大マップを作成
	largeMap := &Map{
		Width:  15,
		Height: 17,
		Data:   make([][]CellType, 17),
	}

	for y := 0; y < 17; y++ {
		largeMap.Data[y] = make([]CellType, 15)
	}

	// 左上（m0: 7×8）
	g.copyRegion(largeMap, m0, 0, 0)

	// 右上（m1: 8×7）
	g.copyRegion(largeMap, m1, 7, 0)

	// 中央隙間1（8×2）
	g.copyRegion(largeMap, gap1, 0, 7)

	// 中央隙間2（8×2）
	g.copyRegion(largeMap, gap2, 7, 7)

	// 左下（m2: 7×8）
	g.copyRegion(largeMap, m2, 0, 9)

	// 右下（m3: 8×7）
	g.copyRegion(largeMap, m3, 7, 9)

	return largeMap
}

// copyRegion copies a source map to a region in the destination map
func (g *Generator) copyRegion(dst, src *Map, offsetX, offsetY int) {
	for y := 0; y < src.Height; y++ {
		for x := 0; x < src.Width; x++ {
			dstY := offsetY + y
			dstX := offsetX + x
			if dstY < dst.Height && dstX < dst.Width {
				dst.Data[dstY][dstX] = src.Data[y][x]
			}
		}
	}
}

// placeAgents places agents diagonally
func (g *Generator) placeAgents(m *Map) {
	// Coolをランダムに配置
	for attempts := 0; attempts < 1000; attempts++ {
		x := g.rng.Intn(m.Width)
		y := g.rng.Intn(m.Height)

		if m.Data[y][x] == Empty {
			m.Cool = Position{X: x, Y: y}

			// Hotを対角に配置
			m.Hot = Position{
				X: m.Width - 1 - x,
				Y: m.Height - 1 - y,
			}

			// Hotの位置が空いているか確認
			if m.Data[m.Hot.Y][m.Hot.X] == Empty {
				return
			}
		}
	}

	// フォールバック: 対角の角に配置
	m.Cool = Position{X: 0, Y: 0}
	m.Hot = Position{X: m.Width - 1, Y: m.Height - 1}
}

// SaveToFile saves the map to a file in CHaser format
func (m *Map) SaveToFile(filename string) error {
	// ディレクトリを作成
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	// エラーハンドリング用のヘルパー関数
	writeErr := func() error {
		// ヘッダー情報
		basename := filepath.Base(filename)
		if _, err := fmt.Fprintf(file, "N generated%s\n", basename); err != nil {
			return fmt.Errorf("failed to write name: %w", err)
		}
		if _, err := fmt.Fprintf(file, "T %d\n", m.Turns); err != nil {
			return fmt.Errorf("failed to write turns: %w", err)
		}
		if _, err := fmt.Fprintf(file, "S %d,%d\n", m.Height, m.Width); err != nil {
			return fmt.Errorf("failed to write size: %w", err)
		}

		// マップデータ
		for y := 0; y < m.Height; y++ {
			for x := 0; x < m.Width; x++ {
				if _, err := fmt.Fprintf(file, "D %d,%d,%d\n", y, x, m.Data[y][x]); err != nil {
					return fmt.Errorf("failed to write map data at (%d,%d): %w", y, x, err)
				}
			}
		}

		// エージェント位置
		if _, err := fmt.Fprintf(file, "H %d,%d\n", m.Hot.Y, m.Hot.X); err != nil {
			return fmt.Errorf("failed to write hot position: %w", err)
		}
		if _, err := fmt.Fprintf(file, "C %d,%d\n", m.Cool.Y, m.Cool.X); err != nil {
			return fmt.Errorf("failed to write cool position: %w", err)
		}

		return nil
	}

	// 書き込み実行
	if err := writeErr(); err != nil {
		file.Close() // エラー時も明示的にクローズ
		return err
	}

	// ファイルを明示的にクローズしてエラーチェック
	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}

	return nil
}

// String returns a string representation of the map
func (m *Map) String() string {
	result := fmt.Sprintf("Map %dx%d (Turns: %d)\n", m.Width, m.Height, m.Turns)
	result += fmt.Sprintf("Hot: (%d,%d), Cool: (%d,%d)\n", m.Hot.X, m.Hot.Y, m.Cool.X, m.Cool.Y)

	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			if m.Hot.Y == y && m.Hot.X == x {
				result += "H"
			} else if m.Cool.Y == y && m.Cool.X == x {
				result += "C"
			} else {
				switch m.Data[y][x] {
				case Empty:
					result += "."
				case Block:
					result += "#"
				case Item:
					result += "*"
				}
			}
		}
		result += "\n"
	}

	return result
}
