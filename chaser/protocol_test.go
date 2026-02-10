package chaser

import (
	"testing"
)

// TestParseResponse は10桁の文字列をResponseに変換するテスト
func TestParseResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		expected Response
	}{
		{
			name:    "正常なレスポンス（すべて空）",
			input:   "0000000000",
			wantErr: false,
			expected: Response{
				GameOver: true,
				Values:   [10]CellType{Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty, Empty},
			},
		},
		{
			name:    "正常なレスポンス（ゲーム継続中）",
			input:   "1231231230",
			wantErr: false,
			expected: Response{
				GameOver: false,
				Values:   [10]CellType{Enemy, Wall, Item, Enemy, Wall, Item, Enemy, Wall, Item, Empty},
			},
		},
		{
			name:    "正常なレスポンス（制御フラグ0でゲームオーバー）",
			input:   "0123012301",
			wantErr: false,
			expected: Response{
				GameOver: true,
				Values:   [10]CellType{Empty, Enemy, Wall, Item, Empty, Enemy, Wall, Item, Empty, Enemy},
			},
		},
		{
			name:    "異常なレスポンス（長さが短い）",
			input:   "012345",
			wantErr: true,
		},
		{
			name:    "異常なレスポンス（長さが長い）",
			input:   "01234567890",
			wantErr: true,
		},
		{
			name:    "異常なレスポンス（数字以外）",
			input:   "012345678a",
			wantErr: true,
		},
		{
			name:    "異常なレスポンス（範囲外の数字）",
			input:   "0123456784",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseResponse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got.GameOver != tt.expected.GameOver {
				t.Errorf("parseResponse() GameOver = %v, want %v", got.GameOver, tt.expected.GameOver)
			}
			if got.Values != tt.expected.Values {
				t.Errorf("parseResponse() Values = %v, want %v", got.Values, tt.expected.Values)
			}
		})
	}
}

// TestEncodeNameForPort はポート番号に応じた名前エンコーディングのテスト
func TestEncodeNameForPort(t *testing.T) {
	tests := []struct {
		name     string
		userName string
		port     string
		wantErr  bool
		checkLen bool // バイト長をチェックするか（文字エンコーディングの違い）
	}{
		{
			name:     "UTF-8エンコーディング（通常ポート）",
			userName: "テスト1",
			port:     "2001",
			wantErr:  false,
			checkLen: false, // UTF-8なのでバイト長は可変
		},
		{
			name:     "CP932エンコーディング（ポート40000）",
			userName: "テスト2",
			port:     "40000",
			wantErr:  false,
			checkLen: true, // CP932なので特定のバイト長
		},
		{
			name:     "CP932エンコーディング（ポート50000）",
			userName: "テスト3",
			port:     "50000",
			wantErr:  false,
			checkLen: true,
		},
		{
			name:     "ASCII文字（通常ポート）",
			userName: "test",
			port:     "3000",
			wantErr:  false,
			checkLen: false,
		},
		{
			name:     "ASCII文字（ポート40000）",
			userName: "test",
			port:     "40000",
			wantErr:  false,
			checkLen: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := encodeNameForPort(tt.userName, tt.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("encodeNameForPort() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if len(got) == 0 {
				t.Errorf("encodeNameForPort() returned empty bytes")
			}
			// エンコーディングが正常に行われたことを確認
			// （詳細な検証は統合テストで行う）
			if tt.checkLen && (tt.port == "40000" || tt.port == "50000") {
				// CP932の場合、日本語は2バイト/文字程度
				// "テスト2" = 6バイト（3文字×2）+ "2"の1バイト = 7バイト程度
				if len(got) < 4 {
					t.Errorf("encodeNameForPort() CP932 encoding seems incorrect, len = %d", len(got))
				}
			}
		})
	}
}

// TestDirectionString はDirection型のString()メソッドのテスト
func TestDirectionString(t *testing.T) {
	tests := []struct {
		dir  Direction
		want string
	}{
		{Up, "Up"},
		{Down, "Down"},
		{Left, "Left"},
		{Right, "Right"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.dir.String(); got != tt.want {
				t.Errorf("Direction.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestCellTypeString はCellType型のString()メソッドのテスト
func TestCellTypeString(t *testing.T) {
	tests := []struct {
		cell CellType
		want string
	}{
		{Empty, "Empty"},
		{Enemy, "Enemy"},
		{Wall, "Wall"},
		{Item, "Item"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.cell.String(); got != tt.want {
				t.Errorf("CellType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDirectionToCommand は方向とプレフィックスからコマンド文字列を生成するテスト
func TestDirectionToCommand(t *testing.T) {
	tests := []struct {
		name   string
		dir    Direction
		prefix byte
		want   string
	}{
		{name: "Walk Up", dir: Up, prefix: 'w', want: "wu"},
		{name: "Walk Down", dir: Down, prefix: 'w', want: "wd"},
		{name: "Walk Left", dir: Left, prefix: 'w', want: "wl"},
		{name: "Walk Right", dir: Right, prefix: 'w', want: "wr"},
		{name: "Look Up", dir: Up, prefix: 'l', want: "lu"},
		{name: "Look Down", dir: Down, prefix: 'l', want: "ld"},
		{name: "Look Left", dir: Left, prefix: 'l', want: "ll"},
		{name: "Look Right", dir: Right, prefix: 'l', want: "lr"},
		{name: "Search Up", dir: Up, prefix: 's', want: "su"},
		{name: "Search Down", dir: Down, prefix: 's', want: "sd"},
		{name: "Search Left", dir: Left, prefix: 's', want: "sl"},
		{name: "Search Right", dir: Right, prefix: 's', want: "sr"},
		{name: "Put Up", dir: Up, prefix: 'p', want: "pu"},
		{name: "Put Down", dir: Down, prefix: 'p', want: "pd"},
		{name: "Put Left", dir: Left, prefix: 'p', want: "pl"},
		{name: "Put Right", dir: Right, prefix: 'p', want: "pr"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := directionToCommand(tt.dir, tt.prefix)
			if got != tt.want {
				t.Errorf("directionToCommand(%v, '%c') = %v, want %v", tt.dir, tt.prefix, got, tt.want)
			}
		})
	}
}
