package main

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

const eventBuffer = 256

func main() {
	listen := flag.Int("listen", 3443, "HTTPS listen port")
	host := flag.String("host", "", "additional SAN for cert (e.g. api.local)")
	lan := flag.Bool("lan", false, "bind on 0.0.0.0 for LAN access (default 127.0.0.1)")
	tui := flag.Bool("tui", false, "full-screen dashboard mode")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Usage = usage
	flag.Parse()

	if *showVersion {
		fmt.Println(Version)
		return
	}

	args := flag.Args()
	if len(args) != 1 {
		usage()
		os.Exit(2)
	}
	targetPort, err := strconv.Atoi(args[0])
	if err != nil || targetPort <= 0 || targetPort > 65535 {
		fmt.Fprintf(os.Stderr, "invalid target port: %q\n", args[0])
		os.Exit(2)
	}

	extra := []string{}
	if *host != "" {
		extra = append(extra, *host)
	}
	certPath, keyPath, err := ensureCert(extra)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load cert: %v\n", err)
		os.Exit(1)
	}

	bindHost := "127.0.0.1"
	if *lan {
		bindHost = "0.0.0.0"
	}
	listenAddr := fmt.Sprintf("%s:%d", bindHost, *listen)
	events := make(chan Event, eventBuffer)
	handler := newProxyHandler(targetPort, events)

	srv := &http.Server{
		Addr:      listenAddr,
		Handler:   handler,
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12},
	}

	uiDone := make(chan struct{})
	go func() {
		defer close(uiDone)
		if *tui {
			runTUI(fmt.Sprintf("localhost:%d", *listen), fmt.Sprintf("http://localhost:%d", targetPort), events)
		} else {
			runLogUI(os.Stderr, fmt.Sprintf("localhost:%d", *listen), fmt.Sprintf("http://localhost:%d", targetPort), events)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- srv.ListenAndServeTLS("", "")
	}()

	select {
	case err := <-serveErr:
		if errors.Is(err, http.ErrServerClosed) {
			break
		}
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		close(events)
		<-uiDone
		os.Exit(1)
	case <-sigCh:
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	close(events)
	<-uiDone
}

func usage() {
	fmt.Fprintln(os.Stderr, `Usage: httpsdev <target-port> [flags]

  httpsdev 5173
      Proxy https://localhost:3443 → http://localhost:5173

Flags:`)
	flag.PrintDefaults()
}
