package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"CHaserGo/chaser"
)

func main() {
	// サーバー情報の取得（環境変数 → 標準入力 → デフォルト）
	host := os.Getenv("CHASER_HOST")
	port := os.Getenv("CHASER_PORT")

	if host == "" {
		fmt.Print("IPアドレス [127.0.0.1]: ")
		fmt.Scanln(&host)
		if host == "" {
			host = "127.0.0.1"
		}
	}

	if port == "" {
		fmt.Print("ポート番号 [2010]: ")
		fmt.Scanln(&port)
		if port == "" {
			port = "2010"
		}
	}

	log.Printf("接続先: %s:%s", host, port)

	// クライアント作成
	client := chaser.NewClient(chaser.ClientConfig{
		Host: host,
		Port: port,
		Name: "テスト②",
	})

	// 接続
	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("接続エラー: %v", err)
	}
	defer client.Disconnect()

	mode := 1

	// ゲームループ
	for {
		// Ready - 準備信号を送り制御情報と周囲情報を取得
		resp, err := client.Ready(ctx)
		if err != nil {
			log.Fatalf("エラー: %v", err)
		}
		if resp.GameOver { // 制御情報が0なら終了
			break
		}

		// modeの値で分岐する
		if mode == 1 {
			if resp.Values[8] != chaser.Wall { // 下が壁でないなら
				resp, err = client.Walk(ctx, chaser.Down) // 下に進む
			} else { // 下が壁なら
				resp, err = client.Walk(ctx, chaser.Right) // 右に進む
				mode = 2
			}
		} else if mode == 2 {
			if resp.Values[6] != chaser.Wall { // 右が壁でないなら
				resp, err = client.Walk(ctx, chaser.Right) // 右に進む
			} else { // 右が壁なら
				resp, err = client.Walk(ctx, chaser.Up) // 上に進む
				mode = 3
			}
		} else if mode == 3 {
			if resp.Values[2] != chaser.Wall { // 上が壁でないなら
				resp, err = client.Walk(ctx, chaser.Up) // 上に進む
			} else { // 上が壁なら
				resp, err = client.Walk(ctx, chaser.Left) // 左に進む
				mode = 4
			}
		} else if mode == 4 {
			if resp.Values[4] != chaser.Wall { // 左が壁でないなら
				resp, err = client.Walk(ctx, chaser.Left) // 左に進む
			} else { // 左が壁なら
				resp, err = client.Walk(ctx, chaser.Down) // 下に進む
				mode = 1
			}
		}

		if err != nil {
			log.Fatalf("エラー: %v", err)
		}

		if resp.GameOver { // 制御情報が0なら終了
			break
		}
	}

	fmt.Println("ゲーム終了")
}
