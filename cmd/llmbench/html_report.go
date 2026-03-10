package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/JanFalkin/llmbench/internal/report"
)

func runHTMLReport(args []string) {
	fs := flag.NewFlagSet("html-report", flag.ExitOnError)

	var input string
	var output string
	var serve bool
	var open bool
	var listenAddr string

	fs.StringVar(&input, "input", "", "Input JSON file")
	fs.StringVar(&output, "output", "report.html", "Output HTML file")
	fs.BoolVar(&serve, "serve", false, "Serve report over local HTTP instead of writing an HTML file")
	fs.BoolVar(&open, "open", false, "Open the report in the default browser (works with --serve)")
	fs.StringVar(&listenAddr, "listen", "127.0.0.1:0", "Address for --serve mode")

	_ = fs.Parse(args)

	if input == "" {
		fmt.Fprintln(os.Stderr, "error: --input is required")
		os.Exit(1)
	}

	if open && !serve {
		fmt.Fprintln(os.Stderr, "error: --open requires --serve")
		os.Exit(1)
	}

	if serve {
		html, err := report.GenerateHTMLReportContent(input)
		if err != nil {
			fmt.Fprintln(os.Stderr, "html-report failed:", err)
			os.Exit(1)
		}

		if err := serveHTMLReport(html, listenAddr, open); err != nil {
			fmt.Fprintln(os.Stderr, "html-report failed:", err)
			os.Exit(1)
		}
		return
	}

	if err := report.GenerateHTMLReport(input, output); err != nil {
		fmt.Fprintln(os.Stderr, "html-report failed:", err)
		os.Exit(1)
	}

	fmt.Println("wrote", output)
}

func serveHTMLReport(html []byte, listenAddr string, open bool) error {
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("listen %q: %w", listenAddr, err)
	}

	var url string
	addr := ln.Addr()
	tcpAddr, ok := addr.(*net.TCPAddr)
	host, port, err := net.SplitHostPort(addr.String())
	if err != nil || !ok {
		// Fallback to the original behavior if we cannot parse the address.
		url = "http://" + addr.String()
	} else {
		// If the listener is on an unspecified address (e.g., 0.0.0.0 or ::),
		// use localhost in the URL so it's reachable and user-friendly.
		if tcpAddr.IP.IsUnspecified() {
			host = "localhost"
		}
		url = "http://" + net.JoinHostPort(host, port)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write(html)
	})

	srv := &http.Server{Handler: mux}

	errCh := make(chan error, 1)
	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	fmt.Println("serving", url)
	fmt.Println("press Ctrl+C to stop")

	if open {
		if err := openBrowser(url); err != nil {
			fmt.Fprintln(os.Stderr, "warning: failed to open browser:", err)
		}
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown server: %w", err)
		}
		return nil
	}
}

const shutdownTimeout = 5 * time.Second

func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}

	return cmd.Start()
}
