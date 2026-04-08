package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kqnade/CHaserGo/gui"
	"github.com/kqnade/CHaserGo/server"
)

const version = "0.3.0"

func main() {
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
		fmt.Fprintf(os.Stderr, "CHaser GUI Server - A CHaser game server with real-time GUI\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <mapfile>\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  <mapfile>    Path to the map file (required)\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *showVersion {
		fmt.Printf("CHaser GUI Server version %s\n", version)
		return
	}

	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: map file is required")
		flag.Usage()
		os.Exit(1)
	}

	mapPath := flag.Arg(0)
	if _, err := os.Stat(mapPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: map file not found: %s\n", mapPath)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())

	// スナップショット channel（buffered=1: 常に最新だけ保持）
	ch := make(chan server.BoardSnapshot, 1)

	config := server.ServerConfig{
		MapPath:    mapPath,
		HotPort:    *hotPort,
		CoolPort:   *coolPort,
		DumpPath:   *dumpPath,
		EnableDump: !*noDump,
		SnapshotCh: ch,
	}

	srv, err := server.NewServer(config)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// 状態管理
	state := &gui.GameState{}
	go state.Run(ch)

	// サーバーを goroutine で起動
	go func() {
		log.Println("=== CHaser GUI Server ===")
		log.Printf("Map: %s", mapPath)
		log.Printf("Hot port: %d", *hotPort)
		log.Printf("Cool port: %d", *coolPort)
		if *noDump {
			log.Println("Dump: disabled")
		} else {
			log.Printf("Dump: %s", *dumpPath)
		}
		log.Println("=========================")

		if err := srv.Start(ctx); err != nil {
			log.Printf("Server error: %v", err)
			cancel()
		}
		close(ch)
	}()

	// Ebitengine をメインスレッドで起動
	ebiten.SetWindowSize(gui.ScreenWidth, gui.ScreenHeight)
	ebiten.SetWindowTitle("CHaser Server GUI")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	app := gui.NewApp(state, cancel)
	if err := ebiten.RunGame(app); err != nil {
		log.Printf("GUI error: %v", err)
	}

	// ウィンドウが閉じられたらサーバーも停止
	cancel()
}
