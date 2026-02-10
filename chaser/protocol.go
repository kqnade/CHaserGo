package chaser

import (
	"errors"
	"fmt"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// Direction は移動・観察方向を表す型
type Direction int

const (
	Up Direction = iota
	Down
	Left
	Right
)

// String はDirection型の文字列表現を返す
func (d Direction) String() string {
	switch d {
	case Up:
		return "Up"
	case Down:
		return "Down"
	case Left:
		return "Left"
	case Right:
		return "Right"
	default:
		return fmt.Sprintf("Direction(%d)", d)
	}
}

// CellType はマスの状態を表す型
type CellType int

const (
	Empty CellType = 0 // 空白またはゲームオーバー
	Enemy CellType = 1 // 敵
	Wall  CellType = 2 // 壁
	Item  CellType = 3 // アイテム
)

// String はCellType型の文字列表現を返す
func (c CellType) String() string {
	switch c {
	case Empty:
		return "Empty"
	case Enemy:
		return "Enemy"
	case Wall:
		return "Wall"
	case Item:
		return "Item"
	default:
		return fmt.Sprintf("CellType(%d)", c)
	}
}

// Response はサーバーからのレスポンスを表す構造体
type Response struct {
	GameOver bool        // values[0] == 0の場合true
	Values   [10]CellType // 周囲9マス+制御フラグ
}

// エラー定義
var (
	ErrInvalidResponseLength = errors.New("invalid response length: expected 10 characters")
	ErrInvalidResponseChar   = errors.New("invalid response character: must be 0-3")
)

// parseResponse は10桁の文字列をResponseに変換する
// レスポンス形式: 10桁の数字（0-3）
// [0]: 制御フラグ（0=ゲームオーバー）
// [1-9]: 周囲のマス情報（0=空, 1=敵, 2=壁, 3=アイテム）
func parseResponse(line string) (*Response, error) {
	// 改行文字を除去
	if len(line) > 0 && (line[len(line)-1] == '\n' || line[len(line)-1] == '\r') {
		line = line[:len(line)-1]
	}
	if len(line) > 0 && (line[len(line)-1] == '\n' || line[len(line)-1] == '\r') {
		line = line[:len(line)-1]
	}

	// 長さチェック
	if len(line) != 10 {
		return nil, fmt.Errorf("%w: got %d", ErrInvalidResponseLength, len(line))
	}

	resp := &Response{}

	// 各文字をパース
	for i := 0; i < 10; i++ {
		char := line[i]
		if char < '0' || char > '3' {
			return nil, fmt.Errorf("%w at position %d: '%c'", ErrInvalidResponseChar, i, char)
		}
		resp.Values[i] = CellType(char - '0')
	}

	// 制御フラグチェック（values[0] == 0の場合はゲームオーバー）
	resp.GameOver = (resp.Values[0] == Empty)

	return resp, nil
}

// encodeNameForPort はポート番号に応じて名前を適切にエンコードする
// ポート40000, 50000: CP932（なでしこサーバー用）
// その他のポート: UTF-8
func encodeNameForPort(name, port string) ([]byte, error) {
	if port == "40000" || port == "50000" {
		// CP932エンコード（ShiftJISエンコーダーはWindows-31J=CP932を使用）
		encoder := japanese.ShiftJIS.NewEncoder()
		encoded, _, err := transform.Bytes(encoder, []byte(name))
		if err != nil {
			return nil, fmt.Errorf("failed to encode name to CP932: %w", err)
		}
		return encoded, nil
	}
	// UTF-8（そのまま）
	return []byte(name), nil
}

// directionToCommand は方向を2文字のコマンド文字列に変換する
func directionToCommand(dir Direction, prefix byte) string {
	suffix := byte('u')
	switch dir {
	case Up:
		suffix = 'u'
	case Down:
		suffix = 'd'
	case Left:
		suffix = 'l'
	case Right:
		suffix = 'r'
	}
	return string([]byte{prefix, suffix})
}
