package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: llmbench <command>")
		fmt.Println("commands: benchmark, sweep, html-report")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "benchmark":
		runBenchmark(os.Args[2:])
	case "sweep":
		runSweep(os.Args[2:])
	case "html-report":
		runHTMLReport(os.Args[2:])
	default:
		fmt.Println("unknown command:", os.Args[1])
	}
}
