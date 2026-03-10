package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/JanFalkin/llmbench/internal/report"
)

func runHTMLReport(args []string) {
	fs := flag.NewFlagSet("html-report", flag.ExitOnError)

	var input string
	var output string

	fs.StringVar(&input, "input", "", "Input JSON file")
	fs.StringVar(&output, "output", "report.html", "Output HTML file")

	_ = fs.Parse(args)

	if input == "" {
		fmt.Fprintln(os.Stderr, "error: --input is required")
		os.Exit(1)
	}

	if err := report.GenerateHTMLReport(input, output); err != nil {
		fmt.Fprintln(os.Stderr, "html-report failed:", err)
		os.Exit(1)
	}

	fmt.Println("wrote", output)
}
