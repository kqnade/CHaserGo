package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/kqnade/CHaserGo/mapgen"
)

const version = "0.2.0"

func main() {
	// コマンドライン引数の定義
	blockNum := flag.Int("b", 9, "Maximum number of blocks in small map")
	flag.IntVar(blockNum, "blockNum", 9, "Maximum number of blocks in small map")

	itemNum := flag.Int("i", 10, "Maximum number of items in small map")
	flag.IntVar(itemNum, "itemNum", 10, "Maximum number of items in small map")

	outputDir := flag.String("o", "./generated_map", "Output directory")
	flag.StringVar(outputDir, "output", "./generated_map", "Output directory")

	seed := flag.Int64("s", 0, "Random seed (0 for current time)")
	flag.Int64Var(seed, "seed", 0, "Random seed (0 for current time)")

	showVersion := flag.Bool("v", false, "Show version")
	flag.BoolVar(showVersion, "version", false, "Show version")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "CHaser MapGenerator - Random map generator for CHaser\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <count>\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  <count>      Number of maps to generate (required)\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s 10                    # Generate 10 maps with default settings\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "  %s -b 15 -i 20 5         # Generate 5 maps with 15 blocks and 20 items\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "  %s -o ./maps -s 12345 3  # Generate 3 maps with specific seed\n", filepath.Base(os.Args[0]))
	}

	flag.Parse()

	// バージョン表示
	if *showVersion {
		fmt.Printf("CHaser MapGenerator version %s\n", version)
		return
	}

	// 生成数のチェック
	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Error: map count is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	var count int
	if _, err := fmt.Sscanf(flag.Arg(0), "%d", &count); err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid count: %s\n", flag.Arg(0))
		os.Exit(1)
	}

	if count <= 0 {
		fmt.Fprintf(os.Stderr, "Error: count must be positive\n")
		os.Exit(1)
	}

	// ジェネレーター作成
	var gen *mapgen.Generator
	if *seed != 0 {
		gen = mapgen.NewGeneratorWithSeed(*seed)
		log.Printf("Using seed: %d", *seed)
	} else {
		gen = mapgen.NewGenerator()
		log.Println("Using random seed")
	}

	// マップ生成
	log.Printf("Generating %d maps...", count)
	log.Printf("Max blocks: %d, Max items: %d", *blockNum, *itemNum)
	log.Printf("Output directory: %s", *outputDir)

	successCount := 0
	for i := 0; i < count; i++ {
		// マップ生成
		m := gen.GenerateMap(*blockNum, *itemNum)

		// ファイル名
		filename := filepath.Join(*outputDir, fmt.Sprintf("RandMap_%d.map", i+1))

		// ファイル保存
		if err := m.SaveToFile(filename); err != nil {
			log.Printf("Warning: failed to save map %d: %v", i+1, err)
			continue
		}

		successCount++
		log.Printf("[%d/%d] Generated: %s", i+1, count, filename)
	}

	if successCount == count {
		log.Printf("Successfully generated %d maps", successCount)
	} else {
		log.Printf("Generated %d/%d maps (some failed)", successCount, count)
		os.Exit(1)
	}
}
