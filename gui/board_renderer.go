package gui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/kqnade/CHaserGo/server"
)

var (
	colorEmpty = color.RGBA{R: 200, G: 200, B: 200, A: 255} // 薄グレー（床）
	colorWall  = color.RGBA{R: 60, G: 60, B: 60, A: 255}    // 濃グレー（壁）
	colorItem  = color.RGBA{R: 255, G: 215, B: 0, A: 255}   // 金色（アイテム）
	colorHot   = color.RGBA{R: 220, G: 50, B: 50, A: 255}   // 赤（Hot）
	colorCool  = color.RGBA{R: 50, G: 50, B: 220, A: 255}   // 青（Cool）
	colorDead  = color.RGBA{R: 120, G: 120, B: 120, A: 180} // グレー（死亡）
)

// BoardRenderer はゲームフィールドを描画する
type BoardRenderer struct{}

// NewBoardRenderer creates a new BoardRenderer
func NewBoardRenderer() *BoardRenderer {
	return &BoardRenderer{}
}

// Draw はゲームボードを描画する
func (r *BoardRenderer) Draw(screen *ebiten.Image, snap *server.BoardSnapshot) {
	boardAreaH := ScreenHeight - HUDHeight

	// タイルサイズをボードサイズに合わせて動的計算（上限40px）
	tileW := ScreenWidth / snap.Width
	tileH := boardAreaH / snap.Height
	tileSize := min(tileW, tileH, 40)

	offsetX := (ScreenWidth - tileSize*snap.Width) / 2
	offsetY := (boardAreaH - tileSize*snap.Height) / 2

	// セルを描画
	for y := 0; y < snap.Height; y++ {
		for x := 0; x < snap.Width; x++ {
			cell := snap.MapFlat[y*snap.Width+x]
			px := float64(offsetX + x*tileSize)
			py := float64(offsetY + y*tileSize)
			ebitenutil.DrawRect(screen, px, py, float64(tileSize-1), float64(tileSize-1), cellColor(cell))
		}
	}

	// Hot マーカー描画
	drawMarker(screen, offsetX+snap.HotX*tileSize, offsetY+snap.HotY*tileSize, tileSize, colorHot, snap.HotAlive)
	// Cool マーカー描画
	drawMarker(screen, offsetX+snap.CoolX*tileSize, offsetY+snap.CoolY*tileSize, tileSize, colorCool, snap.CoolAlive)
}

// cellColor はセル種別に対応する色を返す
func cellColor(cell int) color.Color {
	switch cell {
	case 2: // Wall
		return colorWall
	case 3: // Item
		return colorItem
	default: // Empty
		return colorEmpty
	}
}

// drawMarker はキャラクターを矩形マーカーで描画する
func drawMarker(screen *ebiten.Image, px, py, tileSize int, col color.Color, alive bool) {
	if !alive {
		col = colorDead
	}
	margin := tileSize / 5
	x := float64(px + margin)
	y := float64(py + margin)
	size := float64(tileSize-1) - float64(margin*2)
	ebitenutil.DrawRect(screen, x, y, size, size, col)
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
