package mapgen

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewGenerator(t *testing.T) {
	gen := NewGenerator()
	if gen == nil {
		t.Fatal("NewGenerator returned nil")
	}
	if gen.rng == nil {
		t.Fatal("Generator rng is nil")
	}
}

func TestNewGeneratorWithSeed(t *testing.T) {
	seed := int64(12345)
	gen := NewGeneratorWithSeed(seed)
	if gen == nil {
		t.Fatal("NewGeneratorWithSeed returned nil")
	}

	// 同じシードで同じマップが生成されることを確認
	gen1 := NewGeneratorWithSeed(seed)
	gen2 := NewGeneratorWithSeed(seed)

	map1 := gen1.GenerateMap(9, 10)
	map2 := gen2.GenerateMap(9, 10)

	// サイズが同じであることを確認
	if map1.Width != map2.Width || map1.Height != map2.Height {
		t.Errorf("Maps have different sizes: %dx%d vs %dx%d",
			map1.Width, map1.Height, map2.Width, map2.Height)
	}

	// エージェント位置が同じであることを確認
	if map1.Hot != map2.Hot || map1.Cool != map2.Cool {
		t.Errorf("Agent positions differ: Hot(%v vs %v), Cool(%v vs %v)",
			map1.Hot, map2.Hot, map1.Cool, map2.Cool)
	}
}

func TestGenerateMap(t *testing.T) {
	gen := NewGenerator()
	m := gen.GenerateMap(9, 10)

	// サイズが15×17であることを確認
	if m.Width != 15 || m.Height != 17 {
		t.Errorf("Expected map size 15x17, got %dx%d", m.Width, m.Height)
	}

	// ターン数が120であることを確認
	if m.Turns != 120 {
		t.Errorf("Expected 120 turns, got %d", m.Turns)
	}

	// エージェントが配置されていることを確認
	if m.Hot.X < 0 || m.Hot.X >= m.Width || m.Hot.Y < 0 || m.Hot.Y >= m.Height {
		t.Errorf("Hot agent position out of bounds: %v", m.Hot)
	}

	if m.Cool.X < 0 || m.Cool.X >= m.Width || m.Cool.Y < 0 || m.Cool.Y >= m.Height {
		t.Errorf("Cool agent position out of bounds: %v", m.Cool)
	}

	// エージェントの位置にアイテムやブロックがないことを確認
	if m.Data[m.Hot.Y][m.Hot.X] != Empty {
		t.Errorf("Hot agent position is not empty: %v", m.Data[m.Hot.Y][m.Hot.X])
	}

	if m.Data[m.Cool.Y][m.Cool.X] != Empty {
		t.Errorf("Cool agent position is not empty: %v", m.Data[m.Cool.Y][m.Cool.X])
	}
}

func TestGenerateSmallMap(t *testing.T) {
	gen := NewGenerator()
	m := gen.generateSmallMap(7, 8, 9, 10)

	// サイズが正しいことを確認
	if m.Width != 7 || m.Height != 8 {
		t.Errorf("Expected small map size 7x8, got %dx%d", m.Width, m.Height)
	}

	// ブロックとアイテムの数をカウント
	blockCount := 0
	itemCount := 0

	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			switch m.Data[y][x] {
			case Block:
				blockCount++
				// 最外周にブロックがないことを確認
				if x == 0 || x == m.Width-1 || y == 0 || y == m.Height-1 {
					t.Errorf("Block found on border at (%d,%d)", x, y)
				}
			case Item:
				itemCount++
			}
		}
	}

	t.Logf("Generated small map with %d blocks and %d items", blockCount, itemCount)

	// ブロックとアイテムが最大値以下であることを確認
	if blockCount > 9 {
		t.Errorf("Too many blocks: %d (max 9)", blockCount)
	}

	if itemCount > 10 {
		t.Errorf("Too many items: %d (max 10)", itemCount)
	}
}

func TestRotateMap(t *testing.T) {
	gen := NewGenerator()

	// 3×3のテストマップを作成
	m := &Map{
		Width:  3,
		Height: 3,
		Data: [][]CellType{
			{Empty, Block, Empty},
			{Item, Empty, Empty},
			{Empty, Empty, Block},
		},
	}

	// 90度回転
	rotated := gen.rotateMap(m)

	// サイズが入れ替わることを確認
	if rotated.Width != m.Height || rotated.Height != m.Width {
		t.Errorf("Rotated map has wrong size: %dx%d (expected %dx%d)",
			rotated.Width, rotated.Height, m.Height, m.Width)
	}

	// 4回回転すると元に戻ることを確認
	m1 := gen.rotateMap(m)
	m2 := gen.rotateMap(m1)
	m3 := gen.rotateMap(m2)
	m4 := gen.rotateMap(m3)

	if m4.Width != m.Width || m4.Height != m.Height {
		t.Errorf("Map size changed after 4 rotations")
	}

	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			if m4.Data[y][x] != m.Data[y][x] {
				t.Errorf("Map data changed after 4 rotations at (%d,%d)", x, y)
			}
		}
	}
}

func TestSaveToFile(t *testing.T) {
	gen := NewGenerator()
	m := gen.GenerateMap(5, 5)

	// 一時ディレクトリにファイル保存
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.map")

	if err := m.SaveToFile(filename); err != nil {
		t.Fatalf("Failed to save map: %v", err)
	}

	// ファイルが存在することを確認
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Fatalf("Map file was not created")
	}

	// ファイルを読み込んで基本的な内容を確認
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read map file: %v", err)
	}

	contentStr := string(content)

	// 必須項目が含まれていることを確認
	requiredPrefixes := []string{"N ", "T ", "S ", "D ", "H ", "C "}
	for _, prefix := range requiredPrefixes {
		if !containsPrefix(contentStr, prefix) {
			t.Errorf("Map file missing required prefix: %s", prefix)
		}
	}
}

func containsPrefix(s, prefix string) bool {
	lines := []byte(s)
	target := []byte(prefix)
	for i := 0; i < len(lines)-len(target); i++ {
		if string(lines[i:i+len(target)]) == prefix {
			return true
		}
	}
	return false
}

func TestMapString(t *testing.T) {
	m := &Map{
		Width:  5,
		Height: 5,
		Data: [][]CellType{
			{Empty, Empty, Empty, Empty, Empty},
			{Empty, Block, Empty, Item, Empty},
			{Empty, Empty, Empty, Empty, Empty},
			{Empty, Item, Empty, Block, Empty},
			{Empty, Empty, Empty, Empty, Empty},
		},
		Hot:   Position{X: 0, Y: 0},
		Cool:  Position{X: 4, Y: 4},
		Turns: 120,
	}

	str := m.String()

	// 基本情報が含まれていることを確認
	if str == "" {
		t.Error("Map.String() returned empty string")
	}

	t.Log(str) // デバッグ用に出力
}
