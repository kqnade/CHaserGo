package main

import (
	"context"
	"fmt"
	"log"

	"CHaserGo/chaser"
)

func main() {
	// サーバー情報の入力
	var host, port string
	fmt.Print("IPアドレス: ")
	fmt.Scan(&host)
	fmt.Print("ポート番号: ")
	fmt.Scan(&port)

	// クライアント作成
	client := chaser.NewClient(chaser.ClientConfig{
		Host: host,
		Port: port,
		Name: "テスト1",
	})

	// 接続
	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("接続エラー: %v", err)
	}
	defer client.Disconnect()

	// ゲームループ
	for {
		// Ready
		resp, err := client.Ready(ctx)
		if err != nil {
			log.Fatalf("エラー: %v", err)
		}
		if resp.GameOver {
			break
		}

		// Search Up
		resp, err = client.Search(ctx, chaser.Up)
		if err != nil {
			log.Fatalf("エラー: %v", err)
		}
		if resp.GameOver {
			break
		}

		// Ready
		resp, err = client.Ready(ctx)
		if err != nil {
			log.Fatalf("エラー: %v", err)
		}
		if resp.GameOver {
			break
		}

		// Search Right
		resp, err = client.Search(ctx, chaser.Right)
		if err != nil {
			log.Fatalf("エラー: %v", err)
		}
		if resp.GameOver {
			break
		}

		// Ready
		resp, err = client.Ready(ctx)
		if err != nil {
			log.Fatalf("エラー: %v", err)
		}
		if resp.GameOver {
			break
		}

		// Search Down
		resp, err = client.Search(ctx, chaser.Down)
		if err != nil {
			log.Fatalf("エラー: %v", err)
		}
		if resp.GameOver {
			break
		}

		// Ready
		resp, err = client.Ready(ctx)
		if err != nil {
			log.Fatalf("エラー: %v", err)
		}
		if resp.GameOver {
			break
		}

		// Search Left
		resp, err = client.Search(ctx, chaser.Left)
		if err != nil {
			log.Fatalf("エラー: %v", err)
		}
		if resp.GameOver {
			break
		}
	}

	fmt.Println("ゲーム終了")
}
