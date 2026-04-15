package gui

import (
	"bytes"
	_ "embed"
	"image"
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
)

// Theme はタイルテーマの種別
type Theme int

const (
	ThemeLight Theme = iota
	ThemeHeavy
	ThemeJewel
	ThemeCount
)

func (t Theme) String() string {
	switch t {
	case ThemeLight:
		return "Light"
	case ThemeHeavy:
		return "Heavy"
	case ThemeJewel:
		return "Jewel"
	default:
		return "Unknown"
	}
}

// Light テーマ

//go:embed assets/images/Light/Floor.png
var assetLightFloor []byte

//go:embed assets/images/Light/Block.png
var assetLightBlock []byte

//go:embed assets/images/Light/Item.png
var assetLightItem []byte

//go:embed assets/images/Light/Hot.png
var assetLightHot []byte

//go:embed assets/images/Light/Cool.png
var assetLightCool []byte

// Heavy テーマ

//go:embed assets/images/Heavy/Floor.png
var assetHeavyFloor []byte

//go:embed assets/images/Heavy/Block.png
var assetHeavyBlock []byte

//go:embed assets/images/Heavy/Item.png
var assetHeavyItem []byte

//go:embed assets/images/Heavy/Hot.png
var assetHeavyHot []byte

//go:embed assets/images/Heavy/Cool.png
var assetHeavyCool []byte

// Jewel テーマ

//go:embed assets/images/Jewel/Floor.png
var assetJewelFloor []byte

//go:embed assets/images/Jewel/Block.png
var assetJewelBlock []byte

//go:embed assets/images/Jewel/Item.png
var assetJewelItem []byte

//go:embed assets/images/Jewel/Hot.png
var assetJewelHot []byte

//go:embed assets/images/Jewel/Cool.png
var assetJewelCool []byte

// BGM

//go:embed assets/sounds/ji_023.wav
var assetBGMWAV []byte

// themeImageData はテーマごとの PNG バイト列をまとめた型
type themeImageData struct {
	floor, block, item, hot, cool []byte
}

// allThemeData は全テーマの画像データ
var allThemeData = [ThemeCount]themeImageData{
	ThemeLight: {floor: assetLightFloor, block: assetLightBlock, item: assetLightItem, hot: assetLightHot, cool: assetLightCool},
	ThemeHeavy: {floor: assetHeavyFloor, block: assetHeavyBlock, item: assetHeavyItem, hot: assetHeavyHot, cool: assetHeavyCool},
	ThemeJewel: {floor: assetJewelFloor, block: assetJewelBlock, item: assetJewelItem, hot: assetJewelHot, cool: assetJewelCool},
}

func mustLoadImage(data []byte) *ebiten.Image {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		panic("failed to load embedded image: " + err.Error())
	}
	return ebiten.NewImageFromImage(img)
}
