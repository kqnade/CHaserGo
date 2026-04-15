package server

// SnapshotKind はスナップショットの発火理由
type SnapshotKind int

const (
	KindInitial   SnapshotKind = iota // NewServer 直後の初期盤面
	KindConnected                     // 両者接続完了・名前確定後
	KindActionEnd                     // 各 processTurn() 後（半ターン単位）
	KindTurnEnd                       // IncrementTurn() 後（両者完了後のみ）
	KindGameOver                      // endGame() 後
	KindError                         // 通信エラー等
)

// TurnStep は半ターン単位の先攻/後攻を示す
// KindActionEnd と組み合わせて「誰のアクション後か」を表す
type TurnStep int

const (
	TurnStepFirst  TurnStep = iota // 先攻アクション後
	TurnStepSecond                 // 後攻アクション後
)

// SnapshotPublicPhase は公開ライフサイクルのフェーズ
// Board.GameOver とは別管理（ルール状態は Board が持つ）
type SnapshotPublicPhase int

const (
	PhaseWaiting  SnapshotPublicPhase = iota // プレイヤー接続待ち
	PhaseRunning                             // 対戦中
	PhaseGameOver                            // 終了
	PhaseError                               // エラー
)

// BoardSnapshot はスレッド間で共有するための immutable な盤面スナップショット
// Board への直接参照は一切持たない（全フィールドが値コピー）
type BoardSnapshot struct {
	Kind     SnapshotKind
	Step     TurnStep            // KindActionEnd 時のみ意味を持つ
	Phase    SnapshotPublicPhase
	Revision uint64              // 単調増加、取りこぼし検知用

	// 盤面（1次元平坦化 deep copy）
	// MapData[y][x] = MapFlat[y*Width+x]
	MapFlat  []int
	Width    int
	Height   int
	MaxTurns int
	Turn     int

	WinnerName string
	Reason     string

	HotName  string
	HotX     int
	HotY     int
	HotItems int
	HotAlive bool

	CoolName  string
	CoolX     int
	CoolY     int
	CoolItems int
	CoolAlive bool
}

// SnapshotFromBoard は Board から BoardSnapshot を生成する
// MapData を1次元に平坦化することでスライス参照問題を回避する
func SnapshotFromBoard(b *Board, kind SnapshotKind, step TurnStep, phase SnapshotPublicPhase, rev uint64, winner, reason string) BoardSnapshot {
	flat := make([]int, b.Height*b.Width)
	for y := 0; y < b.Height; y++ {
		for x := 0; x < b.Width; x++ {
			flat[y*b.Width+x] = int(b.MapData[y][x])
		}
	}
	return BoardSnapshot{
		Kind:     kind,
		Step:     step,
		Phase:    phase,
		Revision: rev,
		MapFlat:  flat,
		Width:    b.Width,
		Height:   b.Height,
		MaxTurns: b.MaxTurns,
		Turn:     b.Turn,
		WinnerName: winner,
		Reason:     reason,
		HotName:  b.Hot.Name,
		HotX:     b.Hot.Position.X,
		HotY:     b.Hot.Position.Y,
		HotItems: b.Hot.Items,
		HotAlive: b.Hot.IsAlive,
		CoolName:  b.Cool.Name,
		CoolX:     b.Cool.Position.X,
		CoolY:     b.Cool.Position.Y,
		CoolItems: b.Cool.Items,
		CoolAlive: b.Cool.IsAlive,
	}
}
