# CHaserGo 開発ガイド

このドキュメントでは、CHaserGoの開発に参加する方法について説明します。

## 目次

- [開発環境のセットアップ](#開発環境のセットアップ)
- [プロジェクト構造](#プロジェクト構造)
- [開発プロセス](#開発プロセス)
- [テスト駆動開発（TDD）](#テスト駆動開発tdd)
- [Git Flow](#git-flow)
- [テスト](#テスト)
- [コーディング規約](#コーディング規約)
- [コントリビュート](#コントリビュート)

---

## 開発環境のセットアップ

### 必要なツール

- **Go**: バージョン1.16以上
- **Git**: バージョン管理用
- **VSCode** または **GoLand**: 推奨IDE

### リポジトリのクローン

```bash
git clone https://github.com/kqnade/CHaserGo.git
cd CHaserGo
```

### 依存関係のインストール

```bash
go mod download
```

### テストの実行

```bash
# 全テスト実行
go test ./...

# カバレッジ付き実行
go test -cover ./chaser

# HTMLカバレッジレポート生成
go test -coverprofile=coverage.out ./chaser
go tool cover -html=coverage.out -o coverage.html
```

---

## プロジェクト構造

```
CHaserGo/
├── .gitignore             # Git無視ファイル設定
├── go.mod                 # Goモジュール設定
├── go.sum                 # 依存関係ロックファイル
├── LICENSE                # ライセンス情報
├── README.md              # プロジェクト概要
├── chaser/                # メインパッケージ
│   ├── client.go          # クライアント実装（220行）
│   ├── client_test.go     # クライアントテスト（617行）
│   ├── protocol.go        # プロトコル処理（141行）
│   ├── protocol_test.go   # プロトコルテスト（232行）
│   └── testserver/        # テスト用モックサーバー
│       └── mock.go        # モックTCPサーバー実装（187行）
├── examples/              # サンプルプログラム
│   ├── test1/             # 基本探索ループ
│   │   └── main.go        # Test1.rb移植（109行）
│   ├── test2/             # 壁沿い移動
│   │   └── main.go        # Test2.rb移植（87行）
│   └── test3/             # 複雑なAI
│       └── main.go        # Test3.rb移植（267行）
└── docs/                  # ドキュメント
    ├── API.md             # API詳細ドキュメント
    └── DEVELOPMENT.md     # 開発者向けドキュメント（このファイル）
```

---

## 開発プロセス

CHaserGoはTest-Driven Development (TDD)とGit Flowを採用しています。

### 開発フロー

1. **Issue作成**: 新機能やバグ修正のIssueを作成
2. **ブランチ作成**: `feature/*`または`hotfix/*`ブランチを作成
3. **TDDサイクル**: テスト → 実装 → リファクタリング
4. **コミット**: 機能や修正毎にコミット
5. **プルリクエスト**: `develop`ブランチへのPRを作成
6. **レビュー**: コードレビューを受ける
7. **マージ**: レビュー後、マージ

---

## テスト駆動開発（TDD）

### TDDの基本サイクル

1. **Red**: テストを書く（失敗する）
2. **Green**: 最小限のコードで テストを通す
3. **Refactor**: コードをリファクタリング

### テスト作成のガイドライン

#### ユニットテストの書き方

```go
// chaser/example_test.go
package chaser

import (
    "testing"
)

func TestExample(t *testing.T) {
    // テーブル駆動テスト
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"case1", "input1", "expected1"},
        {"case2", "input2", "expected2"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := exampleFunc(tt.input)
            if result != tt.expected {
                t.Errorf("exampleFunc(%q) = %q, want %q", tt.input, result, tt.expected)
            }
        })
    }
}
```

#### 統合テストの書き方

```go
// chaser/integration_test.go
package chaser

import (
    "context"
    "testing"

    "CHaserGo/chaser/testserver"
)

func TestIntegration(t *testing.T) {
    // モックサーバー起動
    server := testserver.NewMockServer("0")
    if err := server.Start(); err != nil {
        t.Fatalf("failed to start mock server: %v", err)
    }
    defer server.Stop()

    // レスポンス設定
    server.SetResponses([]string{"1234567890"})

    // クライアント接続
    client := NewClient(ClientConfig{
        Host: "127.0.0.1",
        Port: server.Port(),
        Name: "TestClient",
    })

    ctx := context.Background()
    if err := client.Connect(ctx); err != nil {
        t.Fatalf("failed to connect: %v", err)
    }
    defer client.Disconnect()

    // テストロジック...
}
```

### カバレッジ目標

- **全体**: 80%以上
- **重要な関数**: 100%

---

## Git Flow

### ブランチ戦略

```
main (本番リリース用)
  └── develop (開発の主軸)
        ├── feature/xxx (新機能)
        ├── feature/yyy (新機能)
        └── hotfix/zzz (緊急修正)
```

#### ブランチの種類

- **main**: 本番リリース用ブランチ（タグ付き）
- **develop**: 開発の主軸ブランチ
- **feature/\***: 新機能開発用ブランチ
- **release/\***: リリース準備用ブランチ
- **hotfix/\***: 緊急修正用ブランチ

### 新機能の開発

```bash
# developブランチから最新を取得
git checkout develop
git pull origin develop

# feature/xxxブランチを作成
git checkout -b feature/xxx

# TDDサイクルで開発
# 1. テスト作成
# 2. 実装
# 3. テスト実行
go test ./...

# コミット（機能毎）
git add .
git commit -m "feat: add new feature xxx"

# developにマージ
git checkout develop
git merge feature/xxx --no-ff -m "Merge feature/xxx into develop"

# リモートにpush
git push origin develop
```

### コミットメッセージ規約

以下のプレフィックスを使用します:

- `feat:` - 新機能追加
- `fix:` - バグ修正
- `test:` - テスト追加・修正
- `docs:` - ドキュメント変更
- `refactor:` - リファクタリング
- `style:` - フォーマット変更
- `deps:` - 依存関係変更
- `release:` - リリース準備

**例:**

```
feat: add Walk method to Client
fix: correct protocol parsing for GameOver
test: add integration tests for Ready method
docs: update README with installation instructions
```

---

## テスト

### テストの実行

```bash
# 全テスト実行
go test ./...

# 詳細出力
go test -v ./chaser

# カバレッジ付き実行
go test -cover ./chaser

# 特定テスト実行
go test -run TestNewClient ./chaser

# HTMLカバレッジレポート
go test -coverprofile=coverage.out ./chaser
go tool cover -html=coverage.out -o coverage.html
```

### テストの種類

#### 1. プロトコルテスト (protocol_test.go)

- `TestParseResponse`: レスポンスパース
- `TestEncodeNameForPort`: 文字エンコーディング
- `TestDirectionToCommand`: コマンド生成

#### 2. クライアントテスト (client_test.go)

- `TestNewClient`: クライアント作成
- `TestConnectAndDisconnect`: 接続・切断
- `TestReady`: Ready機能
- `TestWalk`: Walk機能
- `TestLook`: Look機能
- `TestSearch`: Search機能
- `TestPut`: Put機能
- エラーケーステスト各種

#### 3. モックサーバーテスト (testserver/mock.go)

- モックサーバー起動・停止
- レスポンス送信
- プロトコル再現

### テスト作成のポイント

1. **テーブル駆動テスト**: 複数のケースを効率的にテスト
2. **モックサーバー**: 実サーバーなしでテスト
3. **エラーケース**: 正常系だけでなくエラーケースもテスト
4. **並行処理**: goroutineを使う場合は競合をテスト

---

## コーディング規約

### Go標準スタイル

```bash
# コードフォーマット
gofmt -w .

# 静的解析
go vet ./...

# リンター実行（golangci-lint推奨）
golangci-lint run
```

### ネーミング規約

- **パッケージ名**: 小文字、単一単語（例: `chaser`, `testserver`）
- **型名**: PascalCase（例: `Client`, `Response`, `Direction`）
- **関数・メソッド名**: PascalCase（公開）、camelCase（非公開）
- **変数名**: camelCase、短く明確に
- **定数名**: PascalCase（例: `Up`, `Down`）

### コメント規約

```go
// NewClient は新しいクライアントを作成する（接続は行わない）
func NewClient(config ClientConfig) *Client {
    // ...
}

// Direction は移動・観察方向を表す型
type Direction int

const (
    Up   Direction = iota // 上
    Down                  // 下
    Left                  // 左
    Right                 // 右
)
```

### エラーハンドリング

```go
// センチネルエラーの定義
var (
    ErrNotConnected    = errors.New("not connected to server")
    ErrAlreadyConnected = errors.New("already connected to server")
)

// エラーラッピング
if err != nil {
    return fmt.Errorf("failed to connect: %w", err)
}
```

---

## コントリビュート

### Issue報告

バグや機能要望がある場合、GitHubのIssueで報告してください。

**テンプレート:**

```markdown
## 概要
[バグ・機能要望の概要]

## 再現手順（バグの場合）
1. ...
2. ...
3. ...

## 期待される動作
[期待される動作]

## 実際の動作
[実際の動作]

## 環境
- OS: Windows/macOS/Linux
- Go Version: 1.xx
- CHaserGo Version: v0.x.x
```

### プルリクエスト

1. **Forkする**: GitHubでリポジトリをFork
2. **ブランチ作成**: `feature/*`ブランチを作成
3. **開発**: TDDで開発
4. **テスト**: 全テストが通ることを確認
5. **コミット**: 規約に従ってコミット
6. **PR作成**: `develop`ブランチへのPRを作成

**PRテンプレート:**

```markdown
## 概要
[変更内容の概要]

## 変更内容
- [ ] 新機能追加
- [ ] バグ修正
- [ ] ドキュメント更新
- [ ] テスト追加

## 関連Issue
Fixes #123

## テスト
- [ ] ユニットテスト追加
- [ ] 全テスト通過
- [ ] カバレッジ80%以上

## チェックリスト
- [ ] `go vet ./...` でエラーなし
- [ ] `gofmt -l .` で整形済み
- [ ] コミットメッセージが規約に従っている
```

---

## リリースプロセス

### バージョニング

セマンティックバージョニング（Semantic Versioning）を採用:

- **MAJOR**: 互換性のない変更
- **MINOR**: 後方互換性のある機能追加
- **PATCH**: 後方互換性のあるバグ修正

**例:** `v1.2.3`

### リリース手順

```bash
# developから最新を取得
git checkout develop
git pull origin develop

# release/vX.Y.Zブランチ作成
git checkout -b release/v0.2.0

# バージョン確定、CHANGELOG.md更新
# 全テスト実行
go test ./...

# リリースコミット
git commit -am "release: prepare v0.2.0"

# mainにマージ
git checkout main
git merge release/v0.2.0 --no-ff -m "Release v0.2.0"

# タグ付け
git tag v0.2.0

# push
git push origin main
git push origin v0.2.0

# developにマージバック
git checkout develop
git merge main
git push origin develop
```

---

## FAQ

### Q1. テストが失敗する場合は?

```bash
# 詳細出力でテスト実行
go test -v ./chaser

# 特定テストのみ実行
go test -run TestReadyErrors ./chaser

# カバレッジ確認
go test -cover ./chaser
```

### Q2. モックサーバーのポート競合エラーは?

```go
// ポート自動割り当てを使用
server := testserver.NewMockServer("0")
```

### Q3. CP932エンコーディングのテストは?

```go
// ポート40000または50000でテスト
config := ClientConfig{
    Host: "127.0.0.1",
    Port: "40000",  // CP932エンコード
    Name: "テスト",
}
```

---

## 参考資料

- [Go公式ドキュメント](https://golang.org/doc/)
- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Semantic Versioning](https://semver.org/)
- [Git Flow](https://nvie.com/posts/a-successful-git-branching-model/)

---

## まとめ

CHaserGoの開発にご協力いただき、ありがとうございます。TDDとGit Flowを活用して、高品質なコードを維持しましょう。

質問や提案がある場合は、GitHubのIssueでお気軽にご連絡ください。
