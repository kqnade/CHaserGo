package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/kqnade/CHaserGo/mapgen"
	"github.com/kqnade/CHaserGo/server"
)

const version = "0.2.0"

func main() {
	// コマンドライン引数の定義
	hotPort := flag.Int("f", 2009, "Hot (first) player port")
	flag.IntVar(hotPort, "first-port", 2009, "Hot (first) player port")

	coolPort := flag.Int("s", 2010, "Cool (second) player port")
	flag.IntVar(coolPort, "second-port", 2010, "Cool (second) player port")

	dumpPath := flag.String("d", "./chaser.dump", "Dump file output path")
	flag.StringVar(dumpPath, "dump-path", "./chaser.dump", "Dump file output path")

	noDump := flag.Bool("nd", false, "Disable dump output")
	flag.BoolVar(noDump, "non-dump", false, "Disable dump output")

	showVersion := flag.Bool("v", false, "Show version")
	flag.BoolVar(showVersion, "version", false, "Show version")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "CHaser Server - A compact CHaser game server\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options] [mapfile]\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  [mapfile]    Path to the map file (optional; auto-generated if omitted)\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "  %s map.txt\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "  %s -f 3000 -s 3001 -d game.dump map.txt\n", filepath.Base(os.Args[0]))
	}

	flag.Parse()

	// バージョン表示
	if *showVersion {
		fmt.Printf("CHaser Server version %s\n", version)
		return
	}

	// マップファイルの決定（省略時は自動生成）
	var mapPath string
	if flag.NArg() >= 1 {
		mapPath = flag.Arg(0)
		if _, err := os.Stat(mapPath); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: map file not found: %s\n", mapPath)
			os.Exit(1)
		}
	} else {
		var tmp *os.File
		cleanupTemp := func() {
			if tmp != nil {
				_ = tmp.Close()
				tmp = nil
			}
			if mapPath != "" {
				_ = os.Remove(mapPath)
			}
		}

		tmp, err := os.CreateTemp("", "chaser-*.map")
		if err != nil {
			log.Printf("Failed to create temp map file: %v", err)
			os.Exit(1)
		}
		mapPath = tmp.Name()
		if err := tmp.Close(); err != nil {
			cleanupTemp()
			log.Printf("Failed to close temp map file: %v", err)
			os.Exit(1)
		}
		tmp = nil
		defer os.Remove(mapPath)

		m := mapgen.NewGenerator().GenerateMap(9, 10)
		if err := m.SaveToFile(mapPath); err != nil {
			cleanupTemp()
			log.Printf("Failed to generate map: %v", err)
			os.Exit(1)
		}
		log.Printf("No map file specified. Auto-generated: %s", mapPath)
	}

	// サーバー設定
	config := server.ServerConfig{
		MapPath:    mapPath,
		HotPort:    *hotPort,
		CoolPort:   *coolPort,
		DumpPath:   *dumpPath,
		EnableDump: !*noDump,
	}

	// サーバー作成
	srv, err := server.NewServer(config)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// サーバー起動
	log.Println("=== CHaser Server ===")
	log.Printf("Map: %s", mapPath)
	log.Printf("Hot port: %d", *hotPort)
	log.Printf("Cool port: %d", *coolPort)
	if *noDump {
		log.Println("Dump: disabled")
	} else {
		log.Printf("Dump: %s", *dumpPath)
	}
	log.Println("=====================")

	if err := srv.Start(context.Background()); err != nil {
		log.Fatalf("Server error: %v", err)
	}

	log.Println("Server finished successfully")
}
