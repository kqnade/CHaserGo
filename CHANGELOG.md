# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2026-02-10

### Added

- **ゲームサーバー** (`server/`パッケージ):
  - CUIベースのローカル対戦サーバー実装
  - `BoardManager`: ボード管理、ゲームルール実装、勝敗判定
  - `Protocol`: ソケット通信、プロトコル処理、レスポンス生成
  - `DumpSystem`: ゲーム記録、CHaserViewer互換形式出力
  - マップファイル読み込み（N, T, S, D, H, C行形式）
  - 先攻・後攻のポート番号カスタマイズ（デフォルト: 2009, 2010）
  - ダンプファイル出力制御
  - compactCHaserServer（Python版）との完全互換

- **マップジェネレーター** (`mapgen/`パッケージ):
  - ランダムマップ生成機能
  - 小マップ（7×8）を4回転させて大マップ（15×17）を構成
  - ブロック・アイテムのランダム配置
  - エージェント対角配置
  - カスタマイズ可能なブロック数・アイテム数
  - シード指定による再現可能な生成
  - バッチ生成機能
  - MapGenerator（Python版）との完全互換

- **コマンドラインツール**:
  - `cmd/chaser-server`: ゲームサーバーCLI
  - `cmd/chaser-mapgen`: マップジェネレーターCLI
  - 各種オプション（ポート番号、ダンプパス、出力先など）
  - バージョン表示機能

- **CI/CD基盤**:
  - GitHub Actions ワークフロー設定
  - ユニットテスト自動実行
  - 統合テスト（サーバー・クライアント対戦テスト）
  - Lintチェック（golangci-lint）
  - マルチプラットフォームビルド（Windows、macOS、Linux）

- **ビルド・テストインフラ**:
  - Makefile追加（build, test, integration-test, mapgen, fmt, lint, clean, install）
  - 統合テストスクリプト（Linux/macOS用 bash、Windows用 PowerShell）
  - テストマップ生成スクリプト

- **ドキュメント**:
  - READMEの大幅拡充
    - サーバー使用方法の追加
    - マップジェネレーター使用方法の追加
    - CI/CD、統合テストの説明追加
    - クイックスタートガイド追加
    - 互換性情報の追加
  - プロジェクト構造の更新
  - 関連リンクの追加（compactCHaserServer、MapGenerator）

### Changed

- プロジェクトのスコープを「クライアントライブラリ」から「完全なGoエコシステム」に拡大
- すべてのCHaser関連ツールをGoで提供

### Technical Details

- **新規パッケージ**: `server/`, `mapgen/`
- **新規コマンド**: `chaser-server`, `chaser-mapgen`
- **CI/CD**: GitHub Actions with 4 workflows (test, integration-test, lint, build)
- **テストカバレッジ**: 維持（80%以上）
- **互換性**:
  - compactCHaserServer（Python版）との完全互換
  - MapGenerator（Python版）との完全互換
  - CHaserViewer互換ダンプ形式

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
