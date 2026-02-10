package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"CHaserGo/chaser"
)

func main() {
	// 乱数初期化
	rand.Seed(time.Now().UnixNano())

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
		Name: "テストⅢ",
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
		// Ready
		resp, err := client.Ready(ctx)
		if err != nil {
			log.Fatalf("エラー: %v", err)
		}
		if resp.GameOver { // 先頭が0になったら終了
			break
		}

		// 上に相手がいたら
		if resp.Values[2] == chaser.Enemy {
			resp, err = client.Put(ctx, chaser.Up)
			if err != nil {
				log.Fatalf("エラー: %v", err)
			}
			if resp.GameOver {
				break
			}
		} else if resp.Values[4] == chaser.Enemy { // 左に相手がいたら
			resp, err = client.Put(ctx, chaser.Left)
			if err != nil {
				log.Fatalf("エラー: %v", err)
			}
			if resp.GameOver {
				break
			}
		} else if resp.Values[6] == chaser.Enemy { // 右に相手がいたら
			resp, err = client.Put(ctx, chaser.Right)
			if err != nil {
				log.Fatalf("エラー: %v", err)
			}
			if resp.GameOver {
				break
			}
		} else if resp.Values[8] == chaser.Enemy { // 下に相手がいたら
			resp, err = client.Put(ctx, chaser.Down)
			if err != nil {
				log.Fatalf("エラー: %v", err)
			}
			if resp.GameOver {
				break
			}
		} else if resp.Values[1] == chaser.Enemy { // 左上に相手がいたら
			if mode == 1 { // 下に進んでいるとき
				resp, err = client.Put(ctx, chaser.Up) // 相手は右に進んでいる
			} else if mode == 2 { // 右に進んでいるとき
				resp, err = client.Put(ctx, chaser.Left) // 相手は下に進んでいる
			} else if mode == 3 { // 上に進んでいるとき
				resp, err = client.Put(ctx, chaser.Left) // 相手は右に進んでいる可能性もあるが下に進んでいると仮定する
			} else if mode == 4 { // 左に進んでいるとき
				resp, err = client.Put(ctx, chaser.Up) // 相手は下に進んでいる可能性もあるが右に進んでいると仮定する
			}
			if err != nil {
				log.Fatalf("エラー: %v", err)
			}
			if resp.GameOver {
				break
			}
		} else if resp.Values[3] == chaser.Enemy { // 右上に相手がいたら
			if mode == 1 { // 下に進んでいるとき
				resp, err = client.Put(ctx, chaser.Up) // 相手は左に進んでいる
			} else if mode == 2 { // 右に進んでいるとき
				resp, err = client.Put(ctx, chaser.Up) // 相手は下に進んでいる可能性もあるが左に進んでいると仮定する
			} else if mode == 3 { // 上に進んでいるとき
				resp, err = client.Put(ctx, chaser.Right) // 相手は左に進んでいる可能性もあるが下に進んでいると仮定する
			} else if mode == 4 { // 左に進んでいるとき
				resp, err = client.Put(ctx, chaser.Right) // 相手は下に進んでいる
			}
			if err != nil {
				log.Fatalf("エラー: %v", err)
			}
			if resp.GameOver {
				break
			}
		} else if resp.Values[9] == chaser.Enemy { // 右下に相手がいたら
			if mode == 1 { // 下に進んでいるとき
				resp, err = client.Put(ctx, chaser.Right) // 相手は左に進んでいる可能性もあるが上に進んでいると仮定する
			} else if mode == 2 { // 右に進んでいるとき
				resp, err = client.Put(ctx, chaser.Down) // 相手は上に進んでいる可能性もあるが左に進んでいると仮定する
			} else if mode == 3 { // 上に進んでいるとき
				resp, err = client.Put(ctx, chaser.Down) // 相手は左に進んでいる
			} else if mode == 4 { // 左に進んでいるとき
				resp, err = client.Put(ctx, chaser.Right) // 相手は上に進んでいる
			}
			if err != nil {
				log.Fatalf("エラー: %v", err)
			}
			if resp.GameOver {
				break
			}
		} else if resp.Values[7] == chaser.Enemy { // 左下に相手がいたら
			if mode == 1 { // 下に進んでいるとき
				resp, err = client.Put(ctx, chaser.Left) // 相手は右に進んでいる可能性もあるが上に進んでいると仮定する
			} else if mode == 2 { // 右に進んでいるとき
				resp, err = client.Put(ctx, chaser.Left) // 相手は上に進んでいる
			} else if mode == 3 { // 上に進んでいるとき
				resp, err = client.Put(ctx, chaser.Down) // 相手は右に進んでいる
			} else if mode == 4 { // 左に進んでいるとき
				resp, err = client.Put(ctx, chaser.Down) // 相手は上に進んでいる可能性もあるが右に進んでいると仮定する
			}
			if err != nil {
				log.Fatalf("エラー: %v", err)
			}
			if resp.GameOver {
				break
			}
		} else if resp.Values[2] == chaser.Item { // 上がアイテムならば
			resp, err = client.Walk(ctx, chaser.Up)
			if err != nil {
				log.Fatalf("エラー: %v", err)
			}
			if resp.GameOver {
				break
			}
		} else if resp.Values[4] == chaser.Item { // 左がアイテムならば
			resp, err = client.Walk(ctx, chaser.Left)
			if err != nil {
				log.Fatalf("エラー: %v", err)
			}
			if resp.GameOver {
				break
			}
		} else if resp.Values[6] == chaser.Item { // 右がアイテムならば
			resp, err = client.Walk(ctx, chaser.Right)
			if err != nil {
				log.Fatalf("エラー: %v", err)
			}
			if resp.GameOver {
				break
			}
		} else if resp.Values[8] == chaser.Item { // 下がアイテムならば
			resp, err = client.Walk(ctx, chaser.Down)
			if err != nil {
				log.Fatalf("エラー: %v", err)
			}
			if resp.GameOver {
				break
			}
		} else {
			// 壁回避移動ロジック
			walked := false
			for !walked {
				// modeの値で分岐する
				if mode == 1 {
					if resp.Values[8] != chaser.Wall { // 下がブロックでないなら
						resp, err = client.Walk(ctx, chaser.Down) // 下に進む
						if err != nil {
							log.Fatalf("エラー: %v", err)
						}
						if resp.GameOver {
							break
						}
						walked = true
					} else { // 下がブロックなら
						mode = 2
						if resp.Values[4] != chaser.Wall { // 左がブロックでないなら
							if rand.Intn(2) == 0 { // 半々の確率でmode変更
								mode = 4
							}
						}
					}
				} else if mode == 2 {
					if resp.Values[6] != chaser.Wall { // 右がブロックでないなら
						resp, err = client.Walk(ctx, chaser.Right) // 右に進む
						if err != nil {
							log.Fatalf("エラー: %v", err)
						}
						if resp.GameOver {
							break
						}
						walked = true
					} else { // 右がブロックなら
						mode = 3
						if resp.Values[8] != chaser.Wall { // 下がブロックでないなら
							if rand.Intn(2) == 0 { // 半々の確率でmode変更
								mode = 1
							}
						}
					}
				} else if mode == 3 {
					if resp.Values[2] != chaser.Wall { // 上がブロックでないなら
						resp, err = client.Walk(ctx, chaser.Up) // 上に進む
						if err != nil {
							log.Fatalf("エラー: %v", err)
						}
						if resp.GameOver {
							break
						}
						walked = true
					} else { // 上がブロックなら
						mode = 4
						if resp.Values[6] != chaser.Wall { // 右がブロックでないなら
							if rand.Intn(2) == 0 { // 半々の確率でmode変更
								mode = 2
							}
						}
					}
				} else if mode == 4 {
					if resp.Values[4] != chaser.Wall { // 左がブロックでないなら
						resp, err = client.Walk(ctx, chaser.Left) // 左に進む
						if err != nil {
							log.Fatalf("エラー: %v", err)
						}
						if resp.GameOver {
							break
						}
						walked = true
					} else { // 左がブロックなら
						mode = 1
						if resp.Values[2] != chaser.Wall { // 上がブロックでないなら
							if rand.Intn(2) == 0 { // 半々の確率でmode変更
								mode = 3
							}
						}
					}
				}
			}

			if resp.GameOver {
				break
			}
		}
	}

	fmt.Println("ゲーム終了")
}
