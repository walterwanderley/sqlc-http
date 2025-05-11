package watcher

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

const defaultPingInterval = 15 * time.Second

type client chan []byte

type WatchStreamer struct {
	watcher       *fsnotify.Watcher
	clients       map[client]struct{}
	connecting    chan client
	disconnecting chan client
	event         chan []byte
	pingInterval  time.Duration
}

func New(dirs ...string) (*WatchStreamer, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	for _, dir := range dirs {
		if err := addRecurssive(w, dir); err != nil {
			return nil, err
		}
	}
	ws := WatchStreamer{
		watcher:       w,
		clients:       make(map[client]struct{}),
		connecting:    make(chan client),
		disconnecting: make(chan client),
		event:         make(chan []byte, 1),
		pingInterval:  defaultPingInterval,
	}
	return &ws, nil
}

func (ws *WatchStreamer) Add(dir string) error {
	return addRecurssive(ws.watcher, dir)
}

func (ws *WatchStreamer) SetPingInterval(interval time.Duration) {
	ws.pingInterval = interval
}

func (ws *WatchStreamer) Start(ctx context.Context) {
	go ws.run(ctx)
	go ws.watch(ctx)
}

func (ws *WatchStreamer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Has("watchList") {
		json.NewEncoder(w).Encode(ws.watcher.WatchList())
		return
	}

	fl, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Flushing not supported", http.StatusNotImplemented)
		return
	}

	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Content-Type", "text/event-stream")

	cl := make(client, 2)
	ws.connecting <- cl

	for {
		select {
		case <-time.After(ws.pingInterval):
			if _, err := w.Write([]byte("event: ping\n\n")); err != nil {
				slog.Error("[watcher] send ping", "error", err.Error())
				return
			}

		case <-r.Context().Done():
			ws.disconnecting <- cl
			return

		case event := <-cl:
			if _, err := w.Write(formatData(event)); err != nil {
				slog.Error("[watcher] send message", "error", err.Error())
				return
			}
			fl.Flush()
		}
	}
}

func formatData(data []byte) []byte {
	var buf bytes.Buffer
	buf.WriteString("data: ")
	buf.Write(data)
	buf.WriteString("\n\n")
	return buf.Bytes()
}

func (ws *WatchStreamer) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case cl := <-ws.connecting:
			slog.Debug("[watcher] connected", "client", cl)
			ws.clients[cl] = struct{}{}

		case cl := <-ws.disconnecting:
			slog.Debug("[watcher] disconnected", "client", cl)
			delete(ws.clients, cl)

		case event := <-ws.event:
			for cl := range ws.clients {
				cl <- event
			}
		}
	}
}

func (ws *WatchStreamer) watch(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			ws.watcher.Close()
			return
		case event, ok := <-ws.watcher.Events:
			if !ok {
				return
			}
			if !event.Has(fsnotify.Write) {
				continue
			}
			slog.Debug("[watcher] file changed", "name", event.Name)
			ws.event <- []byte(event.Name)

		case err, ok := <-ws.watcher.Errors:
			if !ok {
				return
			}
			slog.Error("[watcher] fsnotify", "error", err)
		}
	}
}

func addRecurssive(w *fsnotify.Watcher, dir string) error {
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return w.Add(path)
		}
		return nil
	})
	return err
}
