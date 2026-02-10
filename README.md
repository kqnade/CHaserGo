# CHaserGo

CHaserGo は、プログラミング競技会「U-16プログラミングコンテスト」で使用されるCHaserゲームサーバー用のGoクライアントライブラリです。学生がGo言語でゲームAIを開発できるようにするために作成されました。

## 特徴

- **Go言語らしい設計**: 型安全性、Context対応、構造化されたエラーハンドリング
- **学習用途に最適**: Go言語の特性を学べる設計
- **完全なプロトコル互換**: Ruby版CHaserConnectと同じプロトコルをサポート
- **文字エンコーディング対応**: UTF-8とCP932（Shift_JIS）の両方をサポート
- **テスト駆動開発**: 80%以上のテストカバレッジ
- **サンプルプログラム**: 3つの動作サンプルを同梱

## インストール

```bash
go get github.com/kqnade/CHaserGo
```

または、リポジトリをクローン:

```bash
git clone https://github.com/kqnade/CHaserGo.git
cd CHaserGo
```

## 使い方

### 基本的な使用例

```go
package main

import (
    "context"
    "fmt"
    "log"

    "CHaserGo/chaser"
)

func main() {
    // クライアント作成
    client := chaser.NewClient(chaser.ClientConfig{
        Host: "127.0.0.1",
        Port: "2001",
        Name: "テストAI",
    })

    // 接続
    ctx := context.Background()
    if err := client.Connect(ctx); err != nil {
        log.Fatalf("接続エラー: %v", err)
    }
    defer client.Disconnect()

    // ゲームループ
    for {
        // Ready - ゲーム開始準備
        resp, err := client.Ready(ctx)
        if err != nil {
            log.Fatalf("エラー: %v", err)
        }
        if resp.GameOver {
            break
        }

        // 上方向を探索
        resp, err = client.Search(ctx, chaser.Up)
        if err != nil {
            log.Fatalf("エラー: %v", err)
        }

        // レスポンス内容を確認
        fmt.Printf("Values: %v\n", resp.Values)

        // 上が空いていれば移動
        if resp.Values[2] == chaser.Empty {
            resp, err = client.Walk(ctx, chaser.Up)
            if err != nil {
                log.Fatalf("エラー: %v", err)
            }
        }

        if resp.GameOver {
            break
        }
    }

    fmt.Println("ゲーム終了")
}
```

### Direction（方向）

```go
chaser.Up     // 上
chaser.Down   // 下
chaser.Left   // 左
chaser.Right  // 右
```

### CellType（マスの状態）

```go
chaser.Empty  // 空白またはゲームオーバー
chaser.Enemy  // 敵
chaser.Wall   // 壁
chaser.Item   // アイテム
```

### Response（レスポンス）

```go
type Response struct {
    GameOver bool          // ゲーム終了フラグ
    Values   [10]CellType  // 周囲9マス+制御フラグ
}

// Values配列のインデックス:
// [0] 制御フラグ（0=ゲーム終了）
// [1] 左上斜め
// [2] 上
// [3] 右上斜め
// [4] 左
// [5] 中央（自分）
// [6] 右
// [7] 左下斜め
// [8] 下
// [9] 右下斜め
```

### APIメソッド

```go
// 接続・切断
func (c *Client) Connect(ctx context.Context) error
func (c *Client) Disconnect() error

// ゲーム制御
func (c *Client) Ready(ctx context.Context) (*Response, error)

// 移動
func (c *Client) Walk(ctx context.Context, dir Direction) (*Response, error)

// 観察（隣接マス）
func (c *Client) Look(ctx context.Context, dir Direction) (*Response, error)

// 探索（直線9マス）
func (c *Client) Search(ctx context.Context, dir Direction) (*Response, error)

// ブロック設置
func (c *Client) Put(ctx context.Context, dir Direction) (*Response, error)
```

## サンプルプログラム

### test1: 基本的な探索ループ

4方向を順番に探索するシンプルなAI。

```bash
cd examples/test1
go run main.go
```

### test2: 壁沿い移動

壁に沿って移動するアルゴリズム。状態管理の例として有用。

```bash
cd examples/test2
go run main.go
```

### test3: 複雑なAI

敵検出・攻撃、アイテム収集、壁回避を組み合わせた高度なAI。

```bash
cd examples/test3
go run main.go
```

## テスト

```bash
# 全テスト実行
go test ./...

# カバレッジ付き実行
go test -cover ./chaser

# 詳細出力
go test -v ./chaser
```

## プロジェクト構造

```
CHaserGo/
├── chaser/              # メインパッケージ
│   ├── client.go        # クライアント実装
│   ├── client_test.go   # クライアントテスト
│   ├── protocol.go      # プロトコル処理
│   ├── protocol_test.go # プロトコルテスト
│   └── testserver/      # テスト用モックサーバー
│       └── mock.go
├── examples/            # サンプルプログラム
│   ├── test1/           # 基本探索
│   ├── test2/           # 壁沿い移動
│   └── test3/           # 複雑なAI
├── docs/                # ドキュメント
│   ├── API.md           # API詳細
│   └── DEVELOPMENT.md   # 開発ガイド
├── go.mod
├── go.sum
├── README.md
└── LICENSE
```

## 文字エンコーディング

CHaserGoは、ポート番号に応じて自動的に文字エンコーディングを切り替えます:

- **ポート40000/50000**: CP932（Shift_JIS）エンコード（なでしこサーバー用）
- **その他のポート**: UTF-8

## 開発

詳細な開発ガイドは [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) を参照してください。

## ライセンス

MIT License

## 関連リンク

- [CHaserプロジェクト](http://www.procon.gr.jp/)
- [U-16プログラミングコンテスト](http://www.u16procon.info/)
- [PortableEditor（公式エディタ）](https://github.com/KPC-U16/PortableEditor-Pub)
- [CHaser情報サイト](https://procon.946oss.net/)

## 貢献

Issue報告やPull Requestを歓迎します。詳細は [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) を参照してください。
