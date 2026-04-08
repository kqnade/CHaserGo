package gui

import (
	"bytes"
	_ "embed"
	"image"
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
)

// タイル・キャラクター画像（Light テーマ）

//go:embed assets/images/Light/Floor.png
var assetFloor []byte

//go:embed assets/images/Light/Block.png
var assetBlock []byte

//go:embed assets/images/Light/Item.png
var assetItem []byte

//go:embed assets/images/Light/Hot.png
var assetHot []byte

//go:embed assets/images/Light/Cool.png
var assetCool []byte

// BGM

//go:embed assets/sounds/ji_023.wav
var assetBGMWAV []byte

func mustLoadImage(data []byte) *ebiten.Image {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		panic("failed to load embedded image: " + err.Error())
	}
	return ebiten.NewImageFromImage(img)
}
