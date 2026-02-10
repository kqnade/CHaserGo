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
		fmt.Print("ポート番号 [2009]: ")
		fmt.Scanln(&port)
		if port == "" {
			port = "2009"
		}
	}

	log.Printf("接続先: %s:%s", host, port)

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
			log.Printf("エラー: %v", err)
			break
		}
		if resp.GameOver {
			break
		}

		// Search Up
		resp, err = client.Search(ctx, chaser.Up)
		if err != nil {
			log.Printf("エラー: %v", err)
			break
		}
		if resp.GameOver {
			break
		}

		// Ready
		resp, err = client.Ready(ctx)
		if err != nil {
			log.Printf("エラー: %v", err)
			break
		}
		if resp.GameOver {
			break
		}

		// Search Right
		resp, err = client.Search(ctx, chaser.Right)
		if err != nil {
			log.Printf("エラー: %v", err)
			break
		}
		if resp.GameOver {
			break
		}

		// Ready
		resp, err = client.Ready(ctx)
		if err != nil {
			log.Printf("エラー: %v", err)
			break
		}
		if resp.GameOver {
			break
		}

		// Search Down
		resp, err = client.Search(ctx, chaser.Down)
		if err != nil {
			log.Printf("エラー: %v", err)
			break
		}
		if resp.GameOver {
			break
		}

		// Ready
		resp, err = client.Ready(ctx)
		if err != nil {
			log.Printf("エラー: %v", err)
			break
		}
		if resp.GameOver {
			break
		}

		// Search Left
		resp, err = client.Search(ctx, chaser.Left)
		if err != nil {
			log.Printf("エラー: %v", err)
			break
		}
		if resp.GameOver {
			break
		}
	}

	fmt.Println("ゲーム終了")
}
