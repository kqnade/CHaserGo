package server

import (
	"testing"
)

func TestParseAction(t *testing.T) {
	tests := []struct {
		input     string
		wantAct   string
		wantDir   Direction
		wantErr   bool
	}{
		{"wu", "wk", Up, false},
		{"wd", "wk", Down, false},
		{"wl", "wk", Left, false},
		{"wr", "wk", Right, false},
		{"lu", "lk", Up, false},
		{"ld", "lk", Down, false},
		{"ll", "lk", Left, false},
		{"lr", "lk", Right, false},
		{"su", "sc", Up, false},
		{"sd", "sc", Down, false},
		{"sl", "sc", Left, false},
		{"sr", "sc", Right, false},
		{"pu", "pt", Up, false},
		{"pd", "pt", Down, false},
		{"pl", "pt", Left, false},
		{"pr", "pt", Right, false},
		// 異常系
		{"", "", 0, true},
		{"x", "", 0, true},
		{"wu_extra", "wk", Up, false}, // 3文字目以降は無視
		{"zu", "", 0, true},           // 無効なアクション
		{"wx", "", 0, true},           // 無効な方向
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			act, dir, err := ParseAction(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAction(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if act != tt.wantAct {
				t.Errorf("action = %q, want %q", act, tt.wantAct)
			}
			if dir != tt.wantDir {
				t.Errorf("direction = %v, want %v", dir, tt.wantDir)
			}
		})
	}
}

func TestBuildWalkResponse(t *testing.T) {
	t.Run("ゲーム継続中・周囲9マス", func(t *testing.T) {
		b := newTestBoard()
		// Hot=(1,1): 周囲は(0,0)=W,(0,1)=W,(0,2)=W,(1,0)=W,(1,1)=E,(1,2)=E,(2,0)=W,(2,1)=E,(2,2)=I
		resp := BuildWalkResponse(b.Hot, b.Cool, b)

		if resp[0] != 1 {
			t.Errorf("resp[0] = %d, want 1 (game ongoing)", resp[0])
		}
		// index 1: 左上(0,0)=Wall=2
		if resp[1] != int(Wall) {
			t.Errorf("resp[1] = %d, want %d", resp[1], int(Wall))
		}
		// index 5: 中央(1,1)=Empty=0
		if resp[5] != int(Empty) {
			t.Errorf("resp[5] = %d, want %d (center)", resp[5], int(Empty))
		}
		// index 9: 右下(2,2)=Item=3
		if resp[9] != int(Item) {
			t.Errorf("resp[9] = %d, want %d (item)", resp[9], int(Item))
		}
	})

	t.Run("ゲームオーバー時は制御フラグ0", func(t *testing.T) {
		b := newTestBoard()
		b.GameOver = true
		resp := BuildWalkResponse(b.Hot, b.Cool, b)
		if resp[0] != 0 {
			t.Errorf("resp[0] = %d, want 0 (game over)", resp[0])
		}
	})

	t.Run("相手プレイヤーを Enemy として検出", func(t *testing.T) {
		b := newTestBoard()
		// Hot=(1,1), Cool=(3,3) → 周囲に Cool はいない
		resp := BuildWalkResponse(b.Hot, b.Cool, b)
		for i := 1; i <= 9; i++ {
			if resp[i] == 1 {
				t.Errorf("resp[%d] = 1 (enemy), but Cool should not be adjacent", i)
			}
		}

		// Cool を Hot の隣に移動させる
		b.Cool.Position = Position{Y: 1, X: 2}
		resp2 := BuildWalkResponse(b.Hot, b.Cool, b)
		// index 6: 右(1,2) = Cool = 1
		if resp2[6] != 1 {
			t.Errorf("resp2[6] = %d, want 1 (enemy at right)", resp2[6])
		}
	})
}

func TestBuildLookResponse(t *testing.T) {
	t.Run("2マス先のセルを返す", func(t *testing.T) {
		b := newTestBoard()
		// Hot=(1,1), Right 2 → (1,3)=Empty
		resp := BuildLookResponse(b.Hot, b.Cool, b, Right)
		if resp[0] != 1 {
			t.Errorf("resp[0] = %d, want 1", resp[0])
		}
		if resp[2] != int(Empty) {
			t.Errorf("resp[2] = %d, want %d (Empty)", resp[2], int(Empty))
		}
	})

	t.Run("2マス先に相手がいる場合 Enemy", func(t *testing.T) {
		b := newTestBoard()
		// Hot=(1,1), Cool を (1,3) に配置 → Right 2 で Enemy
		b.Cool.Position = Position{Y: 1, X: 3}
		resp := BuildLookResponse(b.Hot, b.Cool, b, Right)
		if resp[2] != 1 {
			t.Errorf("resp[2] = %d, want 1 (enemy)", resp[2])
		}
	})

	t.Run("ゲームオーバー時は制御フラグ0", func(t *testing.T) {
		b := newTestBoard()
		b.GameOver = true
		resp := BuildLookResponse(b.Hot, b.Cool, b, Up)
		if resp[0] != 0 {
			t.Errorf("resp[0] = %d, want 0", resp[0])
		}
	})
}

func TestBuildSearchResponse(t *testing.T) {
	t.Run("右方向9マスを返す", func(t *testing.T) {
		b := newTestBoard()
		// Hot=(1,1), Right: (1,2)=E,(1,3)=E,(1,4)=W,以降=W
		resp := BuildSearchResponse(b.Hot, b.Cool, b, Right)
		if resp[0] != 1 {
			t.Errorf("resp[0] = %d, want 1", resp[0])
		}
		if resp[1] != int(Empty) {
			t.Errorf("resp[1] = %d, want Empty", resp[1])
		}
		if resp[2] != int(Empty) {
			t.Errorf("resp[2] = %d, want Empty", resp[2])
		}
		if resp[3] != int(Wall) {
			t.Errorf("resp[3] = %d, want Wall", resp[3])
		}
	})

	t.Run("直線上に相手がいる場合 Enemy", func(t *testing.T) {
		b := newTestBoard()
		b.Cool.Position = Position{Y: 1, X: 3}
		resp := BuildSearchResponse(b.Hot, b.Cool, b, Right)
		// index 2: 2マス先(1,3) = Cool = 1
		if resp[2] != 1 {
			t.Errorf("resp[2] = %d, want 1 (enemy)", resp[2])
		}
	})

	t.Run("ゲームオーバー時は制御フラグ0", func(t *testing.T) {
		b := newTestBoard()
		b.GameOver = true
		resp := BuildSearchResponse(b.Hot, b.Cool, b, Up)
		if resp[0] != 0 {
			t.Errorf("resp[0] = %d, want 0", resp[0])
		}
	})
}

func TestBuildPutResponse(t *testing.T) {
	t.Run("WalkResponse と同じ形式", func(t *testing.T) {
		b := newTestBoard()
		walkResp := BuildWalkResponse(b.Hot, b.Cool, b)
		putResp := BuildPutResponse(b.Hot, b.Cool, b)

		for i := 0; i < 10; i++ {
			if walkResp[i] != putResp[i] {
				t.Errorf("resp[%d]: Walk=%d, Put=%d (should match)", i, walkResp[i], putResp[i])
			}
		}
	})
}
