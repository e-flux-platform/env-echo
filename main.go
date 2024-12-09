package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"github.com/elnormous/contenttype"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"
)

var prefixFlag string
var httpAddrFlag string

func init() {
	flag.StringVar(&prefixFlag, "p", "CONFIG_", "prefix for filtering")
	flag.StringVar(&httpAddrFlag, "http", ":8080", "http address")
}

func main() {

	flag.Parse()

	if prefixFlag == "" {
		panic("prefix is required")
	}

	if httpAddrFlag == "" {
		panic("http address is required")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	filteredEnv := make(map[string]string)
	for _, e := range os.Environ() {
		if i := strings.Index(e, "="); i >= 0 {
			if envName, found := strings.CutPrefix(e[:i], prefixFlag); found {
				filteredEnv[envName] = e[i+1:]
			}
		}
	}

	srv := &http.Server{Addr: httpAddrFlag}
	srv.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		availableMediaTypes := []contenttype.MediaType{
			contenttype.NewMediaType("plain/text"),
			contenttype.NewMediaType("application/json"),
		}

		accepted, _, acceptError := contenttype.GetAcceptableMediaType(r, availableMediaTypes)
		if acceptError != nil {
			http.Error(w, acceptError.Error(), http.StatusInternalServerError)
			return
		}

		switch accepted.String() {
		case "plain/text":
			w.Header().Set("Content-Type", "text/plain")
			for k, v := range filteredEnv {
				_, _ = w.Write([]byte(k + "=" + v + "\n"))
			}
			return
		case "application/json":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(filteredEnv)
			return

		}
	})

	g := sync.WaitGroup{}
	g.Add(1)

	go func() {
		<-ctx.Done()
		timeout, c := context.WithTimeout(context.Background(), 5*time.Second)
		defer c()
		defer slog.Info("server stopped")
		defer g.Done()
		_ = srv.Shutdown(timeout)
	}()

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				panic(err)
			}
		}
	}()

	slog.Info("server started", slog.String("prefix", prefixFlag), slog.String("http", httpAddrFlag))

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	select {
	case <-stop:
		slog.Info("signal captured, stopping")
		cancel()
	case <-ctx.Done():
	}

	g.Wait()
}
