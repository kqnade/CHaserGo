package gui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/kqnade/CHaserGo/server"
)

var hudBgColor = color.RGBA{R: 20, G: 20, B: 20, A: 220}

// HUD はスコア・ターン・勝者情報を描画する
type HUD struct{}

// NewHUD creates a new HUD
func NewHUD() *HUD {
	return &HUD{}
}

// Draw はHUD領域（画面下部 HUDHeight px）に情報を描画する
func (h *HUD) Draw(screen *ebiten.Image, snap *server.BoardSnapshot) {
	if snap == nil {
		return
	}
	y := ScreenHeight - HUDHeight

	// 背景
	vector.FillRect(screen, 0, float32(y), float32(ScreenWidth), float32(HUDHeight), hudBgColor, false)

	// ターン表示
	turnStr := fmt.Sprintf("Turn: %d / %d", snap.Turn, snap.MaxTurns)
	ebitenutil.DebugPrintAt(screen, turnStr, 10, y+8)

	// Hot スコア
	hotStr := fmt.Sprintf("[HOT]  %s: %d items", snap.HotName, snap.HotItems)
	if !snap.HotAlive {
		hotStr += " (DEAD)"
	}
	ebitenutil.DebugPrintAt(screen, hotStr, 10, y+28)

	// Cool スコア
	coolStr := fmt.Sprintf("[COOL] %s: %d items", snap.CoolName, snap.CoolItems)
	if !snap.CoolAlive {
		coolStr += " (DEAD)"
	}
	ebitenutil.DebugPrintAt(screen, coolStr, 10, y+48)

	// 勝者表示
	if snap.Phase == server.PhaseGameOver {
		var result string
		if snap.WinnerName != "" {
			result = fmt.Sprintf("WINNER: %s  (%s)", snap.WinnerName, snap.Reason)
		} else {
			result = fmt.Sprintf("DRAW  (%s)", snap.Reason)
		}
		ebitenutil.DebugPrintAt(screen, result, 10, y+68)
	}
}
