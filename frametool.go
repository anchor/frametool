package main

import (
	"flag"
	"fmt"
	"os"
)

type InputSource int

const (
	InputStdin InputSource = iota
	InputFile
)

type OutputDest int
const (
	OutputStdout OutputDest = iota
)

type OutputPacking int
const (
	FramePacking OutputPacking = iota
	BurstPacking
)

type Config struct {
	Output struct {
		// raw
		Format string
		Packing OutputPacking
		Dest OutputDest
		File string
	}
	Input struct {
		Source InputSource
		Files []string
	}
	Gen struct {
		Count uint
	}
	Split struct {
		ChunkSize uint64
		ChunkPrefix string
	}
}

func main() {
	cmds := make(map[string]func(*Config,FrameReader,Writer)(int), 0)
	cmdHelp := make(map[string]string, 0)
	cmds["cat"] = CatCommand
	cmdHelp["cat"] = "Write frames to stdout."
	cmds["gen"] = GenCommand
	cmdHelp["gen"] = "Generate random frames for testing."
	cmds["split"] = SplitCommand
	cmdHelp["split"] = "Split a burst into multiple smaller bursts."

	outputFormat := flag.String("output-fmt", "raw", "Encoding to use for writing frames. One of (raw,json).")
	burstPacking := flag.Bool("output-burst", false, "Pack output as DataBursts instead of DataFrames.")
	outputDest := flag.String("output", "", "File to output to. If not specified, output is written to stdout.")
	maxChunkSize := flag.Uint64("max-chunk-size", 1024, "Maximum size of bursts to output when splitting.")
	chunkPrefix := flag.String("chunk-prefix", "databurst", "Filename prefix to use when splitting bursts.")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <cmd>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Commands:\n\t")
		for cmd, _ := range cmds {
			fmt.Fprintf(os.Stderr, "%s: %s\n", cmd, cmdHelp[cmd])
		}
		flag.PrintDefaults()
	}
	flag.Parse()
	var cfg Config
	cfg.Output.Format = *outputFormat
	if *burstPacking {
		cfg.Output.Packing = BurstPacking
	} else {
		cfg.Output.Packing = FramePacking
	}
	if *outputDest == "" {
		cfg.Output.Dest = OutputStdout
	} else {
		cfg.Output.File = *outputDest
	}
	cfg.Split.ChunkSize = *maxChunkSize
	cfg.Split.ChunkPrefix = *chunkPrefix
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}
	args := flag.Args()
	cmd := args[0]
	if len(args) > 1 {
		cfg.Input.Source = InputFile
		cfg.Input.Files = args[1:]
	} else {
		cfg.Input.Source = InputStdin
	}

	var reader FrameReader
	switch cfg.Input.Source {
	case InputStdin:
		r := NewStreamBurstReader(os.Stdin)
		reader = *r
	case InputFile:
		r := NewFileReader(cfg.Input.Files)
		reader = *r
	default:
		Errorf("This can't happen.")
		os.Exit(1)
	}
	var writer Writer
	switch cfg.Output.Dest {
	case OutputStdout:
		w := NewStreamWriter(os.Stdout)
		writer = *w
	default:
		Errorf("This can't happen.")
		os.Exit(1)
	}

	if f, ok := cmds[cmd]; ok {
		os.Exit(f(&cfg, reader, writer))
	} else {
		flag.Usage()
		os.Exit(1)
	}
}

