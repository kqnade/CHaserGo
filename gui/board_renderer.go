package gui

import (
	"fmt"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/kqnade/CHaserGo/server"
)

type tiles struct {
	floor *ebiten.Image
	wall  *ebiten.Image
	item  *ebiten.Image
	hot   *ebiten.Image
	cool  *ebiten.Image
}

// BoardRenderer はゲームフィールドを描画する
type BoardRenderer struct {
	allTiles     [ThemeCount]tiles
	loadOnce     sync.Once
	currentTheme Theme
}

// NewBoardRenderer creates a new BoardRenderer
func NewBoardRenderer() *BoardRenderer {
	return &BoardRenderer{currentTheme: ThemeLight}
}

// load は最初の Draw 呼び出し時に全テーマのテクスチャを初期化する
func (r *BoardRenderer) load() {
	r.loadOnce.Do(func() {
		for t := Theme(0); t < ThemeCount; t++ {
			d := allThemeData[t]
			r.allTiles[t] = tiles{
				floor: mustLoadImage(d.floor),
				wall:  mustLoadImage(d.block),
				item:  mustLoadImage(d.item),
				hot:   mustLoadImage(d.hot),
				cool:  mustLoadImage(d.cool),
			}
		}
	})
}

// NextTheme は次のテーマに切り替える
func (r *BoardRenderer) NextTheme() {
	r.currentTheme = (r.currentTheme + 1) % ThemeCount
}

// CurrentTheme は現在のテーマを返す
func (r *BoardRenderer) CurrentTheme() Theme {
	return r.currentTheme
}

// Draw はゲームボードを描画する
func (r *BoardRenderer) Draw(screen *ebiten.Image, snap *server.BoardSnapshot) {
	r.load()
	if snap == nil || snap.Width <= 0 || snap.Height <= 0 {
		return
	}

	tx := r.allTiles[r.currentTheme]
	boardAreaH := ScreenHeight - HUDHeight
	tileSize := min(ScreenWidth/snap.Width, boardAreaH/snap.Height, 40)
	offsetX := (ScreenWidth - tileSize*snap.Width) / 2
	offsetY := (boardAreaH - tileSize*snap.Height) / 2

	for y := 0; y < snap.Height; y++ {
		for x := 0; x < snap.Width; x++ {
			cell := snap.MapFlat[y*snap.Width+x]
			r.drawTile(screen, r.cellTile(tx, cell), offsetX+x*tileSize, offsetY+y*tileSize, tileSize)
		}
	}

	r.drawCharacter(screen, tx.hot, snap.HotAlive, offsetX+snap.HotX*tileSize, offsetY+snap.HotY*tileSize, tileSize)
	r.drawCharacter(screen, tx.cool, snap.CoolAlive, offsetX+snap.CoolX*tileSize, offsetY+snap.CoolY*tileSize, tileSize)

	// 現在のテーマ名を右上に表示
	label := fmt.Sprintf("Theme: %s [T]", r.currentTheme)
	ebitenutil.DebugPrintAt(screen, label, ScreenWidth-len(label)*7, 4)
}

func (r *BoardRenderer) cellTile(tx tiles, cell int) *ebiten.Image {
	switch cell {
	case int(server.Wall):
		return tx.wall
	case int(server.Item):
		return tx.item
	default:
		return tx.floor
	}
}

func (r *BoardRenderer) drawTile(screen, src *ebiten.Image, px, py, tileSize int) {
	sw, sh := src.Bounds().Dx(), src.Bounds().Dy()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(tileSize)/float64(sw), float64(tileSize)/float64(sh))
	op.GeoM.Translate(float64(px), float64(py))
	screen.DrawImage(src, op)
}

func (r *BoardRenderer) drawCharacter(screen, src *ebiten.Image, alive bool, px, py, tileSize int) {
	sw, sh := src.Bounds().Dx(), src.Bounds().Dy()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(tileSize)/float64(sw), float64(tileSize)/float64(sh))
	op.GeoM.Translate(float64(px), float64(py))
	if !alive {
		op.ColorScale.ScaleAlpha(0.4)
	}
	screen.DrawImage(src, op)
}
