# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.1] - 2026-02-10

### Fixed

- `TestMockServerMultipleResponses`のタイムアウト問題を修正
  - getReadyコマンドとアクションコマンドの正しいプロトコルフローに対応
  - 初期行"Ready\n"は接続直後の1回のみ受信するように修正
  - 2回目以降はアクションコマンド（walk）を使用するように変更

## [0.1.0] - 2026-02-10

### Added

- **コアライブラリ**:
  - `Client`構造体と基本的な接続管理機能
  - `Connect()`, `Disconnect()`, `Ready()` メソッド
  - `Walk()`, `Look()`, `Search()`, `Put()` の各アクションメソッド
  - 4方向（Up, Down, Left, Right）の`Direction`型
  - マス状態を表す`CellType`型（Empty, Enemy, Wall, Item）
  - サーバーレスポンスを表す`Response`構造体
  - Context対応によるタイムアウト・キャンセル制御

- **プロトコル実装**:
  - CHaserプロトコルの完全実装
  - UTF-8とCP932（Shift_JIS）の自動切り替え
  - ポート40000/50000でのCP932エンコード対応
  - 10桁レスポンスのパース機能
  - ゲームオーバー検出機能

- **テスト**:
  - プロトコル層のユニットテスト（87.5%カバレッジ）
  - クライアント層のユニットテスト（80.0%カバレッジ）
  - モックTCPサーバーによる統合テスト
  - エラーケーステストの充実

- **サンプルプログラム**:
  - `examples/test1`: 基本的な探索ループ
  - `examples/test2`: 壁沿い移動アルゴリズム
  - `examples/test3`: 複雑なAI（敵検出、アイテム収集、壁回避）

- **ドキュメント**:
  - 包括的なREADME.md（インストール、使い方、API概要）
  - 詳細なAPI.md（全型・メソッドの説明、サンプルコード）
  - DEVELOPMENT.md（開発ガイド、TDD、Git Flow）
  - 関連リンク（公式プロジェクト、PortableEditor）

### Technical Details

- **Go Version**: 1.16以上
- **Dependencies**: `golang.org/x/text v0.34.0`
- **Test Coverage**: 80.0%
- **Development Approach**: Test-Driven Development (TDD)
- **Branching Strategy**: Git Flow (main, develop, feature/*)

[0.1.0]: https://github.com/kqnade/CHaserGo/releases/tag/v0.1.0
