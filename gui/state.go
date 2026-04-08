package gui

import (
	"sync/atomic"

	"github.com/kqnade/CHaserGo/server"
)

// GameState はサーバーgoroutineとEbitengineメインスレッド間で
// 共有されるロックフリーなゲーム状態
type GameState struct {
	ptr atomic.Pointer[server.BoardSnapshot]
}

// Run は ch からスナップショットを受信して ptr を更新するループ
// cmd/chaser-server-gui/main.go から goroutine として起動する
func (g *GameState) Run(ch <-chan server.BoardSnapshot) {
	for snap := range ch {
		s := snap
		g.ptr.Store(&s)
	}
}

// Load は現在のスナップショットを返す（nil = 未初期化）
func (g *GameState) Load() *server.BoardSnapshot {
	return g.ptr.Load()
}
