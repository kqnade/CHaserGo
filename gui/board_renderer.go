package gui

import (
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
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
	tiles    tiles
	loadOnce sync.Once
}

// NewBoardRenderer creates a new BoardRenderer
func NewBoardRenderer() *BoardRenderer {
	return &BoardRenderer{}
}

// load は最初の Draw 呼び出し時にテクスチャを初期化する（GPU コンテキストが必要）
func (r *BoardRenderer) load() {
	r.loadOnce.Do(func() {
		r.tiles.floor = mustLoadImage(assetFloor)
		r.tiles.wall = mustLoadImage(assetBlock)
		r.tiles.item = mustLoadImage(assetItem)
		r.tiles.hot = mustLoadImage(assetHot)
		r.tiles.cool = mustLoadImage(assetCool)
	})
}

// Draw はゲームボードを描画する
func (r *BoardRenderer) Draw(screen *ebiten.Image, snap *server.BoardSnapshot) {
	r.load()

	boardAreaH := ScreenHeight - HUDHeight
	tileSize := min(ScreenWidth/snap.Width, boardAreaH/snap.Height, 40)
	offsetX := (ScreenWidth - tileSize*snap.Width) / 2
	offsetY := (boardAreaH - tileSize*snap.Height) / 2

	for y := 0; y < snap.Height; y++ {
		for x := 0; x < snap.Width; x++ {
			cell := snap.MapFlat[y*snap.Width+x]
			r.drawTile(screen, r.cellTile(cell), offsetX+x*tileSize, offsetY+y*tileSize, tileSize)
		}
	}

	r.drawCharacter(screen, r.tiles.hot, snap.HotAlive, offsetX+snap.HotX*tileSize, offsetY+snap.HotY*tileSize, tileSize)
	r.drawCharacter(screen, r.tiles.cool, snap.CoolAlive, offsetX+snap.CoolX*tileSize, offsetY+snap.CoolY*tileSize, tileSize)
}

func (r *BoardRenderer) cellTile(cell int) *ebiten.Image {
	switch cell {
	case 2:
		return r.tiles.wall
	case 3:
		return r.tiles.item
	default:
		return r.tiles.floor
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
