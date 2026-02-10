# CHaserGo API リファレンス

このドキュメントでは、CHaserGoライブラリのAPI詳細について説明します。

## 目次

- [型定義](#型定義)
- [クライアント設定](#クライアント設定)
- [クライアント操作](#クライアント操作)
- [ゲーム操作](#ゲーム操作)
- [エラーハンドリング](#エラーハンドリング)
- [ベストプラクティス](#ベストプラクティス)

---

## 型定義

### Direction

移動・観察・探索・設置の方向を表す型です。

```go
type Direction int

const (
    Up    Direction = iota  // 上
    Down                     // 下
    Left                     // 左
    Right                    // 右
)
```

**使用例:**

```go
// 上方向に移動
resp, err := client.Walk(ctx, chaser.Up)

// 右方向を探索
resp, err := client.Search(ctx, chaser.Right)
```

**メソッド:**

```go
func (d Direction) String() string
```

方向を文字列表現に変換します。

- `Up` → `"Up"`
- `Down` → `"Down"`
- `Left` → `"Left"`
- `Right` → `"Right"`
- その他 → `"Unknown"`

---

### CellType

マスの状態を表す型です。

```go
type CellType int

const (
    Empty  CellType = 0  // 空白またはゲームオーバー
    Enemy  CellType = 1  // 敵
    Wall   CellType = 2  // 壁
    Item   CellType = 3  // アイテム
)
```

**使用例:**

```go
// レスポンスのValues配列を確認
if resp.Values[2] == chaser.Enemy {
    // 上に敵がいる場合の処理
    resp, err = client.Put(ctx, chaser.Up)
}

if resp.Values[8] == chaser.Item {
    // 下にアイテムがある場合の処理
    resp, err = client.Walk(ctx, chaser.Down)
}

if resp.Values[6] == chaser.Wall {
    // 右が壁の場合の処理
}
```

**メソッド:**

```go
func (c CellType) String() string
```

セルタイプを文字列表現に変換します。

- `Empty` → `"Empty"`
- `Enemy` → `"Enemy"`
- `Wall` → `"Wall"`
- `Item` → `"Item"`
- その他 → `"Unknown"`

---

### Response

サーバーからのレスポンスを表す構造体です。

```go
type Response struct {
    GameOver bool          // ゲーム終了フラグ
    Values   [10]CellType  // 周囲9マス+制御フラグ
}
```

**フィールド:**

- `GameOver` (bool): `true`の場合、ゲームが終了しています（`Values[0] == 0`に相当）
- `Values` ([10]CellType): 周囲9マスの状態と制御フラグを格納する配列

**Values配列のインデックス:**

```
[1] [2] [3]    左上  上  右上
[4] [5] [6]  →  左  自分  右
[7] [8] [9]    左下  下  右下

[0] - 制御フラグ（0=ゲーム終了、その他=ゲーム継続中）
```

**使用例:**

```go
resp, err := client.Ready(ctx)
if err != nil {
    log.Fatalf("エラー: %v", err)
}

// ゲーム終了判定
if resp.GameOver {
    fmt.Println("ゲーム終了")
    return
}

// 上方向の確認
if resp.Values[2] == chaser.Empty {
    // 上が空いている
}

// 右方向の確認
if resp.Values[6] == chaser.Wall {
    // 右が壁
}

// 左下斜めの確認
if resp.Values[7] == chaser.Enemy {
    // 左下に敵がいる
}
```

---

## クライアント設定

### ClientConfig

クライアントの接続設定を表す構造体です。

```go
type ClientConfig struct {
    Host string  // サーバーホスト（例: "127.0.0.1"）
    Port string  // サーバーポート（例: "2001"）
    Name string  // プレイヤー名
}
```

**フィールド:**

- `Host` (string): 接続先サーバーのIPアドレスまたはホスト名
- `Port` (string): 接続先サーバーのポート番号（文字列形式）
- `Name` (string): プレイヤー名（サーバーに送信される）

**文字エンコーディング:**

`Port`の値に応じて、`Name`のエンコーディングが自動的に切り替わります:

- ポート`"40000"`または`"50000"`: CP932（Shift_JIS）でエンコード（なでしこサーバー用）
- その他のポート: UTF-8でエンコード

**使用例:**

```go
// UTF-8サーバーへの接続（通常）
config := chaser.ClientConfig{
    Host: "127.0.0.1",
    Port: "2001",
    Name: "テストAI",
}

// CP932サーバーへの接続（なでしこサーバー）
config := chaser.ClientConfig{
    Host: "127.0.0.1",
    Port: "40000",
    Name: "テストAI",
}
```

---

### Client

CHaserサーバーへの接続を管理する構造体です。

```go
type Client struct {
    // 非公開フィールド
}
```

---

## クライアント操作

### NewClient

新しいクライアントを作成します（接続は行いません）。

```go
func NewClient(config ClientConfig) *Client
```

**パラメータ:**

- `config` (ClientConfig): クライアント設定

**戻り値:**

- `*Client`: 新しいクライアントインスタンス

**使用例:**

```go
client := chaser.NewClient(chaser.ClientConfig{
    Host: "127.0.0.1",
    Port: "2001",
    Name: "テストAI",
})
```

---

### Connect

サーバーに接続します。

```go
func (c *Client) Connect(ctx context.Context) error
```

**パラメータ:**

- `ctx` (context.Context): コンテキスト（タイムアウト・キャンセル制御用）

**戻り値:**

- `error`: エラーが発生した場合はエラー、成功時は`nil`

**エラー:**

- `ErrAlreadyConnected`: 既に接続済みの場合
- その他: TCP接続エラー、名前送信エラーなど

**使用例:**

```go
// 基本的な接続
ctx := context.Background()
if err := client.Connect(ctx); err != nil {
    log.Fatalf("接続エラー: %v", err)
}

// タイムアウト付き接続
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

if err := client.Connect(ctx); err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        log.Println("接続タイムアウト")
    } else {
        log.Fatalf("接続エラー: %v", err)
    }
}
```

---

### Disconnect

サーバーから切断します。

```go
func (c *Client) Disconnect() error
```

**戻り値:**

- `error`: エラーが発生した場合はエラー、成功時は`nil`

**注意:**

- 既に切断済みの場合、エラーは返さずに`nil`を返します
- `defer`を使って確実に切断することを推奨します

**使用例:**

```go
if err := client.Connect(ctx); err != nil {
    log.Fatalf("接続エラー: %v", err)
}
defer client.Disconnect()

// ゲームロジック...
```

---

## ゲーム操作

### Ready

ゲーム開始準備を通知し、現在の状態を取得します。

```go
func (c *Client) Ready(ctx context.Context) (*Response, error)
```

**パラメータ:**

- `ctx` (context.Context): コンテキスト

**戻り値:**

- `*Response`: サーバーからのレスポンス
- `error`: エラーが発生した場合はエラー、成功時は`nil`

**プロトコル:**

1. サーバーから初期メッセージ（`"Ready\n"`）を受信
2. `"gr\r\n"`コマンドを送信
3. レスポンス（10桁の数字）を受信

**エラー:**

- `ErrNotConnected`: 未接続の場合
- その他: 通信エラー、パースエラーなど

**使用例:**

```go
resp, err := client.Ready(ctx)
if err != nil {
    log.Fatalf("エラー: %v", err)
}

if resp.GameOver {
    fmt.Println("ゲーム終了")
    return
}

// 周囲の状態を確認
fmt.Printf("上: %s\n", resp.Values[2])
fmt.Printf("右: %s\n", resp.Values[6])
```

---

### Walk

指定方向に移動します。

```go
func (c *Client) Walk(ctx context.Context, dir Direction) (*Response, error)
```

**パラメータ:**

- `ctx` (context.Context): コンテキスト
- `dir` (Direction): 移動方向（`Up`, `Down`, `Left`, `Right`）

**戻り値:**

- `*Response`: 移動後のレスポンス
- `error`: エラーが発生した場合はエラー、成功時は`nil`

**プロトコル:**

1. `"wu\r\n"` (Up), `"wd\r\n"` (Down), `"wl\r\n"` (Left), `"wr\r\n"` (Right)を送信
2. レスポンス（10桁の数字）を受信
3. 確認応答`"#\r\n"`を送信

**エラー:**

- `ErrNotConnected`: 未接続の場合
- その他: 通信エラー、パースエラーなど

**使用例:**

```go
// 上に移動
resp, err := client.Walk(ctx, chaser.Up)
if err != nil {
    log.Fatalf("エラー: %v", err)
}

if resp.GameOver {
    fmt.Println("ゲーム終了")
    return
}

// 移動後の状態を確認
fmt.Printf("移動後の周囲: %v\n", resp.Values)
```

---

### Look

指定方向の隣接マスを観察します。

```go
func (c *Client) Look(ctx context.Context, dir Direction) (*Response, error)
```

**パラメータ:**

- `ctx` (context.Context): コンテキスト
- `dir` (Direction): 観察方向（`Up`, `Down`, `Left`, `Right`）

**戻り値:**

- `*Response`: 観察結果のレスポンス
- `error`: エラーが発生した場合はエラー、成功時は`nil`

**プロトコル:**

1. `"lu\r\n"` (Up), `"ld\r\n"` (Down), `"ll\r\n"` (Left), `"lr\r\n"` (Right)を送信
2. レスポンス（10桁の数字）を受信
3. 確認応答`"#\r\n"`を送信

**エラー:**

- `ErrNotConnected`: 未接続の場合
- その他: 通信エラー、パースエラーなど

**使用例:**

```go
// 右方向を観察
resp, err := client.Look(ctx, chaser.Right)
if err != nil {
    log.Fatalf("エラー: %v", err)
}

if resp.Values[6] == chaser.Enemy {
    fmt.Println("右に敵がいる!")
}
```

---

### Search

指定方向の直線9マスを探索します。

```go
func (c *Client) Search(ctx context.Context, dir Direction) (*Response, error)
```

**パラメータ:**

- `ctx` (context.Context): コンテキスト
- `dir` (Direction): 探索方向（`Up`, `Down`, `Left`, `Right`）

**戻り値:**

- `*Response`: 探索結果のレスポンス
- `error`: エラーが発生した場合はエラー、成功時は`nil`

**プロトコル:**

1. `"su\r\n"` (Up), `"sd\r\n"` (Down), `"sl\r\n"` (Left), `"sr\r\n"` (Right)を送信
2. レスポンス（10桁の数字）を受信
3. 確認応答`"#\r\n"`を送信

**エラー:**

- `ErrNotConnected`: 未接続の場合
- その他: 通信エラー、パースエラーなど

**使用例:**

```go
// 上方向を探索
resp, err := client.Search(ctx, chaser.Up)
if err != nil {
    log.Fatalf("エラー: %v", err)
}

// 探索結果を確認
for i, val := range resp.Values {
    fmt.Printf("Values[%d]: %s\n", i, val)
}
```

---

### Put

指定方向にブロックを設置します。

```go
func (c *Client) Put(ctx context.Context, dir Direction) (*Response, error)
```

**パラメータ:**

- `ctx` (context.Context): コンテキスト
- `dir` (Direction): 設置方向（`Up`, `Down`, `Left`, `Right`）

**戻り値:**

- `*Response`: 設置後のレスポンス
- `error`: エラーが発生した場合はエラー、成功時は`nil`

**プロトコル:**

1. `"pu\r\n"` (Up), `"pd\r\n"` (Down), `"pl\r\n"` (Left), `"pr\r\n"` (Right)を送信
2. レスポンス（10桁の数字）を受信
3. 確認応答`"#\r\n"`を送信

**エラー:**

- `ErrNotConnected`: 未接続の場合
- その他: 通信エラー、パースエラーなど

**使用例:**

```go
// 敵が上にいる場合、ブロックを設置
if resp.Values[2] == chaser.Enemy {
    resp, err = client.Put(ctx, chaser.Up)
    if err != nil {
        log.Fatalf("エラー: %v", err)
    }
    fmt.Println("ブロックを設置しました")
}
```

---

## エラーハンドリング

### 定義済みエラー

```go
var (
    ErrNotConnected    = errors.New("not connected to server")
    ErrAlreadyConnected = errors.New("already connected to server")
)
```

**使用例:**

```go
import "errors"

if err := client.Connect(ctx); err != nil {
    if errors.Is(err, chaser.ErrAlreadyConnected) {
        fmt.Println("既に接続済みです")
    } else {
        log.Fatalf("接続エラー: %v", err)
    }
}

resp, err := client.Ready(ctx)
if err != nil {
    if errors.Is(err, chaser.ErrNotConnected) {
        fmt.Println("サーバーに接続していません")
    } else {
        log.Fatalf("エラー: %v", err)
    }
}
```

---

## ベストプラクティス

### 1. deferでDisconnectを確実に呼び出す

```go
if err := client.Connect(ctx); err != nil {
    log.Fatalf("接続エラー: %v", err)
}
defer client.Disconnect()
```

### 2. GameOverフラグを確認する

```go
for {
    resp, err := client.Ready(ctx)
    if err != nil {
        log.Fatalf("エラー: %v", err)
    }

    if resp.GameOver {
        break  // ゲーム終了
    }

    // ゲームロジック...
}
```

### 3. タイムアウトを設定する

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

if err := client.Connect(ctx); err != nil {
    log.Fatalf("接続エラー: %v", err)
}
```

### 4. エラーハンドリングを適切に行う

```go
resp, err := client.Walk(ctx, chaser.Up)
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        fmt.Println("タイムアウト")
    } else if errors.Is(err, chaser.ErrNotConnected) {
        fmt.Println("未接続")
    } else {
        log.Fatalf("エラー: %v", err)
    }
    return
}
```

### 5. Values配列のインデックスを定数化する

```go
const (
    IdxUp        = 2
    IdxDown      = 8
    IdxLeft      = 4
    IdxRight     = 6
    IdxUpLeft    = 1
    IdxUpRight   = 3
    IdxDownLeft  = 7
    IdxDownRight = 9
)

// 使用例
if resp.Values[IdxUp] == chaser.Enemy {
    // 上に敵がいる
}
```

### 6. ロギングを活用する

```go
import "log"

log.Printf("接続中: %s:%s\n", config.Host, config.Port)
if err := client.Connect(ctx); err != nil {
    log.Fatalf("接続エラー: %v", err)
}
log.Println("接続成功")
```

---

## サンプルコード集

### 壁回避移動

```go
mode := 1  // 1=Down, 2=Right, 3=Up, 4=Left

for {
    resp, err := client.Ready(ctx)
    if err != nil || resp.GameOver {
        break
    }

    if mode == 1 {
        if resp.Values[8] != chaser.Wall {
            resp, _ = client.Walk(ctx, chaser.Down)
        } else {
            resp, _ = client.Walk(ctx, chaser.Right)
            mode = 2
        }
    } else if mode == 2 {
        if resp.Values[6] != chaser.Wall {
            resp, _ = client.Walk(ctx, chaser.Right)
        } else {
            resp, _ = client.Walk(ctx, chaser.Up)
            mode = 3
        }
    } else if mode == 3 {
        if resp.Values[2] != chaser.Wall {
            resp, _ = client.Walk(ctx, chaser.Up)
        } else {
            resp, _ = client.Walk(ctx, chaser.Left)
            mode = 4
        }
    } else if mode == 4 {
        if resp.Values[4] != chaser.Wall {
            resp, _ = client.Walk(ctx, chaser.Left)
        } else {
            resp, _ = client.Walk(ctx, chaser.Down)
            mode = 1
        }
    }

    if resp.GameOver {
        break
    }
}
```

### アイテム収集

```go
for {
    resp, err := client.Ready(ctx)
    if err != nil || resp.GameOver {
        break
    }

    // 4方向のアイテムチェック
    if resp.Values[2] == chaser.Item {
        resp, _ = client.Walk(ctx, chaser.Up)
    } else if resp.Values[4] == chaser.Item {
        resp, _ = client.Walk(ctx, chaser.Left)
    } else if resp.Values[6] == chaser.Item {
        resp, _ = client.Walk(ctx, chaser.Right)
    } else if resp.Values[8] == chaser.Item {
        resp, _ = client.Walk(ctx, chaser.Down)
    }

    if resp.GameOver {
        break
    }
}
```

### 敵検出と攻撃

```go
for {
    resp, err := client.Ready(ctx)
    if err != nil || resp.GameOver {
        break
    }

    // 4方向の敵チェック
    if resp.Values[2] == chaser.Enemy {
        resp, _ = client.Put(ctx, chaser.Up)
    } else if resp.Values[4] == chaser.Enemy {
        resp, _ = client.Put(ctx, chaser.Left)
    } else if resp.Values[6] == chaser.Enemy {
        resp, _ = client.Put(ctx, chaser.Right)
    } else if resp.Values[8] == chaser.Enemy {
        resp, _ = client.Put(ctx, chaser.Down)
    }

    if resp.GameOver {
        break
    }
}
```

---

## まとめ

CHaserGoライブラリは、Go言語らしい設計で型安全性とエラーハンドリングを提供します。サンプルコードを参考に、独自のゲームAIを開発してください。

詳細な開発ガイドは [DEVELOPMENT.md](DEVELOPMENT.md) を参照してください。
