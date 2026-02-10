package chaser

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

// ClientConfig はクライアント設定
type ClientConfig struct {
	Host string // サーバーホスト（例: "127.0.0.1"）
	Port string // サーバーポート（例: "2001"）
	Name string // プレイヤー名
}

// Client はCHaserサーバーへの接続を管理する
type Client struct {
	config ClientConfig
	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
}

// エラー定義
var (
	ErrNotConnected    = errors.New("not connected to server")
	ErrAlreadyConnected = errors.New("already connected to server")
	ErrGameOver        = errors.New("game over")
)

// NewClient はクライアントを作成する（接続は行わない）
func NewClient(config ClientConfig) *Client {
	return &Client{
		config: config,
	}
}

// Connect はサーバーに接続する
func (c *Client) Connect(ctx context.Context) error {
	if c.conn != nil {
		return ErrAlreadyConnected
	}

	// タイムアウト付きでダイヤル
	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(c.config.Host, c.config.Port))
	if err != nil {
		return fmt.Errorf("failed to connect to %s:%s: %w", c.config.Host, c.config.Port, err)
	}

	c.conn = conn
	c.reader = bufio.NewReader(conn)
	c.writer = bufio.NewWriter(conn)

	// 名前をエンコードして送信
	nameBytes, err := encodeNameForPort(c.config.Name, c.config.Port)
	if err != nil {
		c.conn.Close()
		c.conn = nil
		return fmt.Errorf("failed to encode name: %w", err)
	}

	_, err = c.writer.Write(nameBytes)
	if err != nil {
		c.conn.Close()
		c.conn = nil
		return fmt.Errorf("failed to send name: %w", err)
	}

	_, err = c.writer.WriteString("\n")
	if err != nil {
		c.conn.Close()
		c.conn = nil
		return fmt.Errorf("failed to send name: %w", err)
	}

	err = c.writer.Flush()
	if err != nil {
		c.conn.Close()
		c.conn = nil
		return fmt.Errorf("failed to flush name: %w", err)
	}

	return nil
}

// Disconnect はサーバーから切断する
func (c *Client) Disconnect() error {
	if c.conn == nil {
		return nil
	}

	err := c.conn.Close()
	c.conn = nil
	c.reader = nil
	c.writer = nil

	if err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}

	return nil
}

// Ready はゲーム開始準備を通知する
// Ruby版のgetReadyに相当
func (c *Client) Ready(ctx context.Context) (*Response, error) {
	if c.conn == nil {
		return nil, ErrNotConnected
	}

	// getReadyは特殊: 初期行を読み取る + "gr\r" 送信
	// 1. 初期行を読み取る（"Ready"など）
	_, err := c.reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read initial line: %w", err)
	}

	// 2. "gr\r"を送信（Ruby版では\r\nではなく\rのみ）
	_, err = c.writer.WriteString("gr\r\n")
	if err != nil {
		return nil, fmt.Errorf("failed to send getReady command: %w", err)
	}

	err = c.writer.Flush()
	if err != nil {
		return nil, fmt.Errorf("failed to flush getReady command: %w", err)
	}

	// 3. レスポンスを読み取る
	line, err := c.reader.ReadString('\n')
	if err != nil {
		// EOFはゲーム終了を意味する
		if err.Error() == "EOF" || strings.Contains(err.Error(), "connection reset") {
			return &Response{GameOver: true}, nil
		}
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// 4. レスポンスをパース
	resp, err := parseResponse(line)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return resp, nil
}

// sendCommand は汎用的なコマンド送信処理
// getReady以外のコマンド（walk, look, search, put）で使用
func (c *Client) sendCommand(ctx context.Context, cmd string) (*Response, error) {
	if c.conn == nil {
		return nil, ErrNotConnected
	}

	// 1. コマンド送信（"XX\r\n"形式）
	_, err := c.writer.WriteString(cmd + "\r\n")
	if err != nil {
		return nil, fmt.Errorf("failed to send command: %w", err)
	}

	err = c.writer.Flush()
	if err != nil {
		return nil, fmt.Errorf("failed to flush command: %w", err)
	}

	// 2. レスポンス読み取り
	line, err := c.reader.ReadString('\n')
	if err != nil {
		// EOFはゲーム終了を意味する
		if err.Error() == "EOF" || strings.Contains(err.Error(), "connection reset") {
			return &Response{GameOver: true}, nil
		}
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// 3. レスポンスをパース
	resp, err := parseResponse(line)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// ゲームオーバーの場合は確認応答を送らない
	if resp.GameOver {
		return resp, nil
	}

	// 4. 確認応答送信（"#\r\n"）
	_, err = c.writer.WriteString("#\r\n")
	if err != nil {
		return nil, fmt.Errorf("failed to send acknowledgment: %w", err)
	}

	err = c.writer.Flush()
	if err != nil {
		return nil, fmt.Errorf("failed to flush acknowledgment: %w", err)
	}

	return resp, nil
}

// Walk は指定方向に移動する
func (c *Client) Walk(ctx context.Context, dir Direction) (*Response, error) {
	cmd := directionToCommand(dir, 'w')
	return c.sendCommand(ctx, cmd)
}

// Look は指定方向の隣接マスを観察する
func (c *Client) Look(ctx context.Context, dir Direction) (*Response, error) {
	cmd := directionToCommand(dir, 'l')
	return c.sendCommand(ctx, cmd)
}

// Search は指定方向の直線9マスを探索する
func (c *Client) Search(ctx context.Context, dir Direction) (*Response, error) {
	cmd := directionToCommand(dir, 's')
	return c.sendCommand(ctx, cmd)
}

// Put は指定方向にブロックを設置する
func (c *Client) Put(ctx context.Context, dir Direction) (*Response, error) {
	cmd := directionToCommand(dir, 'p')
	return c.sendCommand(ctx, cmd)
}

// SetDeadline は接続のデッドラインを設定する
func (c *Client) SetDeadline(t time.Time) error {
	if c.conn == nil {
		return ErrNotConnected
	}
	return c.conn.SetDeadline(t)
}
