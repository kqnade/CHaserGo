package server

import (
	"reflect"
	"testing"
)

// newTestBoard は 5×5 のテスト用盤面を返す
//
//	W W W W W
//	W . . . W
//	W . I . W
//	W . . . W
//	W W W W W
//
// Hot=(y:1,x:1), Cool=(y:3,x:3), Item=(y:2,x:2)
func newTestBoard() *Board {
	board := &Board{
		Width:    5,
		Height:   5,
		MaxTurns: 100,
		Hot:      &Character{Name: "Hot", Position: Position{Y: 1, X: 1}, IsAlive: true},
		Cool:     &Character{Name: "Cool", Position: Position{Y: 3, X: 3}, IsAlive: true},
		Turn:     0,
	}
	board.MapData = make([][]CellType, board.Height)
	for i := range board.MapData {
		board.MapData[i] = make([]CellType, board.Width)
	}
	for x := 0; x < board.Width; x++ {
		board.MapData[0][x] = Wall
		board.MapData[board.Height-1][x] = Wall
	}
	for y := 0; y < board.Height; y++ {
		board.MapData[y][0] = Wall
		board.MapData[y][board.Width-1] = Wall
	}
	board.MapData[2][2] = Item
	return board
}

func TestNewBoard(t *testing.T) {
	b, err := NewBoard("testdata/test.map")
	if err != nil {
		t.Fatalf("NewBoard: %v", err)
	}
	if b.Width != 5 || b.Height != 5 {
		t.Errorf("size = %dx%d, want 5x5", b.Width, b.Height)
	}
	if b.MaxTurns != 100 {
		t.Errorf("MaxTurns = %d, want 100", b.MaxTurns)
	}
	if b.Hot.Position != (Position{Y: 1, X: 1}) {
		t.Errorf("Hot position = %v, want {1,1}", b.Hot.Position)
	}
	if b.Cool.Position != (Position{Y: 3, X: 3}) {
		t.Errorf("Cool position = %v, want {3,3}", b.Cool.Position)
	}
	if b.MapData[2][2] != Item {
		t.Errorf("MapData[2][2] = %v, want Item", b.MapData[2][2])
	}
}

func TestNewBoardFileNotFound(t *testing.T) {
	_, err := NewBoard("testdata/nonexistent.map")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestGetCell(t *testing.T) {
	b := newTestBoard()

	tests := []struct {
		pos  Position
		want CellType
	}{
		{Position{Y: 0, X: 0}, Wall},
		{Position{Y: 1, X: 1}, Empty},
		{Position{Y: 2, X: 2}, Item},
		// 境界外は Wall 扱い
		{Position{Y: -1, X: 0}, Wall},
		{Position{Y: 0, X: -1}, Wall},
		{Position{Y: 5, X: 0}, Wall},
		{Position{Y: 0, X: 5}, Wall},
	}

	for _, tt := range tests {
		if got := b.GetCell(tt.pos); got != tt.want {
			t.Errorf("GetCell(%v) = %v, want %v", tt.pos, got, tt.want)
		}
	}
}

func TestSetCell(t *testing.T) {
	b := newTestBoard()
	pos := Position{Y: 1, X: 2}

	b.SetCell(pos, Wall)
	if b.MapData[1][2] != Wall {
		t.Errorf("SetCell: expected Wall at (1,2)")
	}

	// 境界外は no-op
	before := make([][]CellType, len(b.MapData))
	for y := range b.MapData {
		before[y] = append([]CellType(nil), b.MapData[y]...)
	}

	for _, pos := range []Position{
		{Y: -1, X: 0},
		{Y: 0, X: -1},
		{Y: b.Height, X: 0},
		{Y: 0, X: b.Width},
	} {
		b.SetCell(pos, Item)
	}

	if !reflect.DeepEqual(b.MapData, before) {
		t.Fatal("SetCell should not modify the board for out-of-bounds positions")
	}
}

func TestMove(t *testing.T) {
	b := newTestBoard()
	start := Position{Y: 2, X: 2}

	tests := []struct {
		dir  Direction
		want Position
	}{
		{Up, Position{Y: 1, X: 2}},
		{Down, Position{Y: 3, X: 2}},
		{Left, Position{Y: 2, X: 1}},
		{Right, Position{Y: 2, X: 3}},
	}

	for _, tt := range tests {
		if got := b.Move(start, tt.dir); got != tt.want {
			t.Errorf("Move(%v, %v) = %v, want %v", start, tt.dir, got, tt.want)
		}
	}
}

func TestWalk(t *testing.T) {
	t.Run("通常移動", func(t *testing.T) {
		b := newTestBoard()
		char := b.Hot // (1,1)

		if err := b.Walk(char, Right); err != nil {
			t.Fatalf("Walk Right: %v", err)
		}
		if char.Position != (Position{Y: 1, X: 2}) {
			t.Errorf("position = %v, want {1,2}", char.Position)
		}
		if b.GameOver {
			t.Error("GameOver should be false")
		}
	})

	t.Run("アイテム収集", func(t *testing.T) {
		b := newTestBoard()
		char := b.Hot // (1,1) → right → (1,2) → down → (2,2)=Item

		_ = b.Walk(char, Right)
		if err := b.Walk(char, Down); err != nil {
			t.Fatalf("Walk Down: %v", err)
		}
		if char.Items != 1 {
			t.Errorf("Items = %d, want 1", char.Items)
		}
		if b.MapData[2][2] != Empty {
			t.Error("item should be consumed")
		}
	})

	t.Run("壁に衝突→ゲームオーバー", func(t *testing.T) {
		b := newTestBoard()
		char := b.Hot // (1,1) → Up → (0,1)=Wall

		err := b.Walk(char, Up)
		if err == nil {
			t.Error("expected error when hitting wall")
		}
		if char.IsAlive {
			t.Error("IsAlive should be false after hitting wall")
		}
		if !b.GameOver {
			t.Error("GameOver should be true")
		}
	})

	t.Run("移動後に囲まれない（移動元は常にEmpty）", func(t *testing.T) {
		// Walk 後に移動元セルは Empty のままなので IsSurrounded は false
		b := newTestBoard()
		// (1,2) 周囲のうち移動元(1,1)以外を壁にする
		b.MapData[0][2] = Wall // Up (already wall)
		b.MapData[2][2] = Wall // Down
		b.MapData[1][3] = Wall // Right

		char := b.Hot
		err := b.Walk(char, Right)
		if err != nil {
			t.Errorf("Walk should succeed: %v", err)
		}
		// (1,1) は Empty なので囲まれていない
		if !char.IsAlive {
			t.Error("IsAlive should be true (not actually surrounded)")
		}
	})
}

func TestLook(t *testing.T) {
	b := newTestBoard()
	// Hot=(1,1): Up 2 → (−1,1)=境界外=Wall, Right 2 → (1,3)=Empty, Down 2 → (3,1)=Empty

	tests := []struct {
		dir  Direction
		want CellType
	}{
		{Up, Wall},     // (1,1) → (-1,1) = 境界外
		{Right, Empty}, // (1,1) → (1,3) = Empty
		{Down, Empty},  // (1,1) → (3,1) = Empty
		{Left, Wall},   // (1,1) → (1,-1) = 境界外
	}

	for _, tt := range tests {
		if got := b.Look(b.Hot.Position, tt.dir); got != tt.want {
			t.Errorf("Look(%v) = %v, want %v", tt.dir, got, tt.want)
		}
	}
}

func TestSearch(t *testing.T) {
	b := newTestBoard()
	// Hot=(1,1), Right 方向: (1,2)=Empty, (1,3)=Empty, (1,4)=Wall, 以降は境界外=Wall

	result := b.Search(b.Hot.Position, Right)

	if result[0] != Empty {
		t.Errorf("result[0] = %v, want Empty", result[0])
	}
	if result[1] != Empty {
		t.Errorf("result[1] = %v, want Empty", result[1])
	}
	if result[2] != Wall {
		t.Errorf("result[2] = %v, want Wall", result[2])
	}
	for i := 3; i < 9; i++ {
		if result[i] != Wall {
			t.Errorf("result[%d] = %v, want Wall", i, result[i])
		}
	}
}

func TestPut(t *testing.T) {
	t.Run("空セルに壁を置く", func(t *testing.T) {
		b := newTestBoard()
		// Hot=(1,1), Right → (1,2)=Empty
		b.Put(b.Hot.Position, Right)
		if b.MapData[1][2] != Wall {
			t.Error("expected Wall at (1,2) after Put")
		}
	})

	t.Run("既存の壁には置けない", func(t *testing.T) {
		b := newTestBoard()
		// Hot=(1,1), Up → (0,1)=Wall
		b.Put(b.Hot.Position, Up)
		if b.MapData[0][1] != Wall {
			t.Error("wall should remain Wall")
		}
	})
}

func TestIsSurrounded(t *testing.T) {
	b := newTestBoard()

	// Hot=(1,1) は囲まれていない
	if b.IsSurrounded(b.Hot) {
		t.Error("Hot should not be surrounded initially")
	}

	// 四方を壁にする
	b.MapData[0][1] = Wall // Up (already wall)
	b.MapData[2][1] = Wall // Down
	b.MapData[1][0] = Wall // Left (already wall)
	b.MapData[1][2] = Wall // Right

	if !b.IsSurrounded(b.Hot) {
		t.Error("Hot should be surrounded")
	}
}

func TestGetResult(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*Board)
		wantNil    bool // winner が nil かどうか
		wantReason string
	}{
		{
			name: "Hot が多くアイテム保有",
			setup: func(b *Board) {
				b.Hot.Items = 5
				b.Cool.Items = 3
			},
			wantNil:    false,
			wantReason: "hot has more items",
		},
		{
			name: "Cool が多くアイテム保有",
			setup: func(b *Board) {
				b.Hot.Items = 2
				b.Cool.Items = 4
			},
			wantNil:    false,
			wantReason: "cool has more items",
		},
		{
			name: "引き分け",
			setup: func(b *Board) {
				b.Hot.Items = 3
				b.Cool.Items = 3
			},
			wantNil:    true,
			wantReason: "draw",
		},
		{
			name: "Hot が死亡",
			setup: func(b *Board) {
				b.Hot.IsAlive = false
			},
			wantNil:    false,
			wantReason: "hot died",
		},
		{
			name: "Cool が死亡",
			setup: func(b *Board) {
				b.Cool.IsAlive = false
			},
			wantNil:    false,
			wantReason: "cool died",
		},
		{
			name: "両者死亡・Hot が多い",
			setup: func(b *Board) {
				b.Hot.IsAlive = false
				b.Cool.IsAlive = false
				b.Hot.Items = 3
				b.Cool.Items = 1
			},
			wantNil:    false,
			wantReason: "both died, hot has more items",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newTestBoard()
			tt.setup(b)
			winner, reason := b.GetResult()
			if (winner == nil) != tt.wantNil {
				t.Errorf("winner nil = %v, want %v", winner == nil, tt.wantNil)
			}
			if reason != tt.wantReason {
				t.Errorf("reason = %q, want %q", reason, tt.wantReason)
			}
		})
	}
}

func TestIncrementTurn(t *testing.T) {
	b := newTestBoard()

	b.IncrementTurn()
	if b.Turn != 1 {
		t.Errorf("Turn = %d, want 1", b.Turn)
	}
	if b.GameOver {
		t.Error("GameOver should be false")
	}

	// MaxTurns まで進める
	b.Turn = b.MaxTurns - 1
	b.IncrementTurn()
	if !b.GameOver {
		t.Error("GameOver should be true at MaxTurns")
	}
}

func TestGetOpponent(t *testing.T) {
	b := newTestBoard()

	if b.GetOpponent(b.Hot) != b.Cool {
		t.Error("GetOpponent(Hot) should return Cool")
	}
	if b.GetOpponent(b.Cool) != b.Hot {
		t.Error("GetOpponent(Cool) should return Hot")
	}
}
