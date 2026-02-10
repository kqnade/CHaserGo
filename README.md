# CHaserGo

CHaserGo は、プログラミング競技会「U-16プログラミングコンテスト」で使用されるCHaserゲーム用の**完全なGoエコシステム**です。クライアントライブラリ、ゲームサーバー、マップジェネレーターを含み、CHaser開発に必要なすべてをGoで提供します。

## 特徴

### クライアントライブラリ
- **Go言語らしい設計**: 型安全性、Context対応、構造化されたエラーハンドリング
- **学習用途に最適**: Go言語の特性を学べる設計
- **完全なプロトコル互換**: Ruby版CHaserConnectと同じプロトコルをサポート
- **文字エンコーディング対応**: UTF-8とCP932（Shift_JIS）の両方をサポート
- **テスト駆動開発**: 80%以上のテストカバレッジ
- **サンプルプログラム**: 3つの動作サンプルを同梱

### ゲームサーバー
- **CUI対戦サーバー**: ローカルでのAIテストに最適
- **ダンプ機能**: CHaserViewer互換のゲーム記録
- **カスタマイズ可能**: ポート番号、ターン数の調整が可能
- **完全互換**: 既存のCHaserクライアントとも動作

### マップジェネレーター
- **ランダムマップ生成**: 多様なマップを自動生成
- **カスタマイズ可能**: ブロック数、アイテム数を指定可能
- **バッチ生成**: 複数マップを一度に生成
- **完全互換**: CHaser標準マップフォーマットに対応

## Goエコシステムの利点

すべてをGoで実装することで、以下のメリットがあります：

- **簡単なセットアップ**: `go install`だけで全ツールがインストール可能
- **クロスプラットフォーム**: Windows、macOS、Linuxで同じコードが動作
- **CI/CD統合**: GitHub Actionsで自動テストが簡単
- **単一言語**: 学習コストが低く、メンテナンスが容易
- **高速**: Go言語の高速な実行速度を活用

## インストール

### ライブラリのみ使用（AI開発）

```bash
go get github.com/kqnade/CHaserGo
```

または、プロジェクトのgo.modに以下を追加:

```go
require github.com/kqnade/CHaserGo v0.2.0
```

### フルツールセット（開発・テスト環境）

```bash
# リポジトリをクローン
git clone https://github.com/kqnade/CHaserGo.git
cd CHaserGo

# すべてのツールをインストール
make install

# または個別にインストール
go install ./cmd/chaser-server      # ゲームサーバー
go install ./cmd/chaser-mapgen      # マップジェネレーター
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

## ゲームサーバー

ローカルでAIをテストするための簡易サーバー。

### 基本的な使い方

```bash
# マップファイルを指定してサーバー起動
chaser-server map.txt

# ポート番号をカスタマイズ
chaser-server -f 3000 -s 3001 map.txt

# ダンプファイルを指定
chaser-server -d game.dump map.txt

# ダンプを無効化
chaser-server -nd map.txt
```

### オプション

- `-f, --first-port`: 先攻プレイヤーのポート（デフォルト: 2009）
- `-s, --second-port`: 後攻プレイヤーのポート（デフォルト: 2010）
- `-d, --dump-path`: ダンプファイルの出力先（デフォルト: ./chaser.dump）
- `-nd, --non-dump`: ダンプ出力を無効化

### 実行例

```bash
# 1. サーバーを起動
chaser-server map.txt

# 2. 別のターミナルでクライアントを起動
cd examples/test1
go run main.go  # ポート2009に接続

# 3. さらに別のターミナルで2つ目のクライアントを起動
cd examples/test2
go run main.go  # ポート2010に接続
```

## マップジェネレーター

ランダムなゲームマップを生成するツール。

### 基本的な使い方

```bash
# 10個のマップを生成
chaser-mapgen 10

# ブロック数とアイテム数を指定
chaser-mapgen -b 15 -i 20 5

# 出力先を指定
chaser-mapgen -o ./my_maps 3

# シードを指定（再現可能な生成）
chaser-mapgen -s 12345 5
```

### オプション

- `-b, --blockNum`: 小マップ内の最大ブロック数（デフォルト: 9）
- `-i, --itemNum`: 小マップ内の最大アイテム数（デフォルト: 10）
- `-o, --output`: 出力ディレクトリ（デフォルト: ./generated_map）
- `-s, --seed`: ランダムシード（0で現在時刻を使用）

### マップ仕様

- サイズ: 15×17
- 生成アルゴリズム: 7×8の小マップを4回転させて結合
- エージェント配置: 対角配置
- 出力形式: CHaser標準フォーマット

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

## テスト・CI/CD

### ユニットテスト

```bash
# 全テスト実行
go test ./...

# カバレッジ付き実行
go test -cover ./...

# 詳細出力
go test -v ./...
```

### 統合テスト

サーバーを使った実際の対戦テスト。

```bash
# Makeを使った統合テスト
make integration-test

# 手動で実行
make build
make mapgen
bash scripts/integration-test.sh  # Linux/macOS
pwsh scripts/integration-test.ps1 # Windows
```

### Makefile

開発を便利にするMakefileを提供しています。

```bash
make help              # ヘルプ表示
make build             # 全ツールをビルド
make test              # ユニットテスト実行
make integration-test  # 統合テスト実行
make mapgen            # サンプルマップ生成
make fmt               # コードフォーマット
make lint              # リンター実行
make clean             # ビルド成果物を削除
make install           # 全ツールをインストール
```

### GitHub Actions

CI/CDパイプラインが設定済みです：

- **ユニットテスト**: すべてのプッシュで自動実行
- **統合テスト**: サーバー・クライアント間の対戦テスト
- **Lint**: コード品質チェック
- **マルチプラットフォームビルド**: Windows、macOS、Linuxでビルド検証

## プロジェクト構造

```
CHaserGo/
├── chaser/              # クライアントライブラリ
│   ├── client.go        # クライアント実装
│   ├── client_test.go   # クライアントテスト
│   ├── protocol.go      # プロトコル処理
│   ├── protocol_test.go # プロトコルテスト
│   └── testserver/      # テスト用モックサーバー
│       └── mock.go
├── server/              # ゲームサーバー
│   ├── server.go        # サーバー本体
│   ├── board.go         # ボード管理
│   ├── protocol.go      # プロトコル処理
│   └── dump.go          # ダンプシステム
├── mapgen/              # マップジェネレーター
│   ├── generator.go     # マップ生成ロジック
│   └── generator_test.go
├── cmd/                 # コマンドラインツール
│   ├── chaser-server/   # サーバーCLI
│   │   └── main.go
│   └── chaser-mapgen/   # マップ生成CLI
│       └── main.go
├── examples/            # サンプルプログラム
│   ├── test1/           # 基本探索
│   ├── test2/           # 壁沿い移動
│   └── test3/           # 複雑なAI
├── scripts/             # テストスクリプト
│   ├── integration-test.sh
│   └── integration-test.ps1
├── .github/workflows/   # CI/CD設定
│   └── ci.yml
├── docs/                # ドキュメント
│   ├── API.md           # API詳細
│   └── DEVELOPMENT.md   # 開発ガイド
├── Makefile             # ビルド自動化
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

## クイックスタート

完全な開発環境をセットアップして、すぐに開発を始められます。

```bash
# 1. リポジトリをクローン
git clone https://github.com/kqnade/CHaserGo.git
cd CHaserGo

# 2. ツールをビルド
make build

# 3. マップを生成
make mapgen

# 4. サーバーを起動（別ターミナル）
./bin/chaser-server testdata/RandMap_1.map

# 5. サンプルAIを実行（別ターミナル×2）
./bin/test1  # ポート2009
./bin/test2  # ポート2010

# 6. 統合テストを実行
make integration-test
```

## 互換性

### CHaser公式ツールとの互換性

- **クライアントライブラリ**: Ruby版CHaserConnectと完全互換
- **ゲームサーバー**: Python版compactCHaserServerと互換
- **マップジェネレーター**: Python版MapGeneratorと互換
- **マップフォーマット**: CHaser標準フォーマット
- **ダンプフォーマット**: CHaserViewer互換

### 動作確認済み環境

- Go 1.24以降
- Windows 10/11
- macOS 12以降
- Ubuntu 20.04以降

## 関連リンク

- [PortableEditor（公式エディタ）](https://github.com/KPC-U16/PortableEditor-Pub)
- [CHaser情報サイト](https://procon.946oss.net/)
- [compactCHaserServer（Python版）](https://github.com/yugu0202/compactCHaserServer)
- [MapGenerator（Python版）](https://github.com/KPC-U16/MapGenerator)

## 貢献

Issue報告やPull Requestを歓迎します。詳細は [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) を参照してください。
