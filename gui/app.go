package gui

import (
	"bytes"
	"context"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/kqnade/CHaserGo/server"
)

const (
	ScreenWidth  = 800
	ScreenHeight = 600
	HUDHeight    = 100
	sampleRate   = 44100
)

// App は Ebitengine のゲームループを管理する
type App struct {
	state     *GameState
	renderer  *BoardRenderer
	hud       *HUD
	cancel    context.CancelFunc
	bgmPlayer *audio.Player
}

// compile-time check
var _ ebiten.Game = (*App)(nil)

// NewApp creates a new App
func NewApp(state *GameState, cancel context.CancelFunc) *App {
	app := &App{
		state:    state,
		renderer: NewBoardRenderer(),
		hud:      NewHUD(),
		cancel:   cancel,
	}

	// BGM 初期化
	audioCtx := audio.NewContext(sampleRate)
	stream, err := wav.DecodeWithSampleRate(sampleRate, bytes.NewReader(assetBGMWAV))
	if err != nil {
		log.Printf("BGM decode error: %v", err)
		return app
	}
	loop := audio.NewInfiniteLoop(stream, stream.Length())
	player, err := audioCtx.NewPlayer(loop)
	if err != nil {
		log.Printf("BGM player error: %v", err)
		return app
	}
	app.bgmPlayer = player
	return app
}

// Update is called every tick (60fps)
func (a *App) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyT) {
		a.renderer.NextTheme()
	}

	if ebiten.IsKeyPressed(ebiten.KeyEscape) || ebiten.IsKeyPressed(ebiten.KeyQ) {
		if a.bgmPlayer != nil {
			a.bgmPlayer.Close()
		}
		a.cancel()
		return ebiten.Termination
	}

	// BGM 制御: ゲーム開始で再生、終了で停止
	if a.bgmPlayer != nil {
		snap := a.state.Load()
		switch {
		case snap != nil && snap.Phase == server.PhaseRunning && !a.bgmPlayer.IsPlaying():
			a.bgmPlayer.Play()
		case snap != nil && snap.Phase == server.PhaseGameOver && a.bgmPlayer.IsPlaying():
			a.bgmPlayer.Pause()
		}
	}

	return nil
}

// Draw is called every frame
func (a *App) Draw(screen *ebiten.Image) {
	snap := a.state.Load()

	if snap == nil || snap.Phase == server.PhaseWaiting {
		msg := "Waiting for players..."
		if snap != nil {
			msg = "Waiting for players to connect..."
		}
		drawWaitingScreen(screen, msg)
		return
	}

	a.renderer.Draw(screen, snap)
	a.hud.Draw(screen, snap)
}

// Layout returns the logical screen size
func (a *App) Layout(_, _ int) (int, int) {
	return ScreenWidth, ScreenHeight
}

// drawWaitingScreen は接続待ち画面を描画する
func drawWaitingScreen(screen *ebiten.Image, msg string) {
	screen.Fill(color.RGBA{R: 30, G: 30, B: 30, A: 255})
	ebitenutil.DebugPrintAt(screen, "CHaser Server GUI", ScreenWidth/2-80, ScreenHeight/2-20)
	ebitenutil.DebugPrintAt(screen, msg, ScreenWidth/2-len(msg)*3, ScreenHeight/2)
	ebitenutil.DebugPrintAt(screen, "Press ESC to quit", ScreenWidth/2-80, ScreenHeight/2+20)
}
