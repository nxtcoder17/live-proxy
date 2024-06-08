package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nxtcoder17/live-proxy/pkg/log"
	"github.com/nxtcoder17/live-proxy/pkg/websocket"
	"github.com/nxtcoder17/live-proxy/templates"
)

func main() {
	var addr string
	flag.StringVar(&addr, "addr", ":8080", "http service address")

	var proxyAddr string
	flag.StringVar(&proxyAddr, "proxy-addr", "localhost:8081", "websocket proxy address")

	flag.Parse()

	logger := log.NewLogger(log.Options{
		Level:      slog.LevelDebug,
		ShowCaller: true,
	})

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)

	r.Get("/healthy", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Get("/ws", websocket.HttpHandler(websocket.Args{
		Logger: logger,
		Handler: func(ctx websocket.Context, ws websocket.Conn) error {
			defer func() {
				defer ws.Close("going away")
				logger.Info("CLIENT disconnected")
			}()
			for {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				c, err := net.DialTimeout("tcp", proxyAddr, 1*time.Second)
				if err != nil {
					// msg := "proxy destination not reachable"
					msg := `
					<div id="proxy-reachable">false</div>
					<span id="status-text">proxy destination not reachable</span>
					`
					if err := ws.WriteMessageText(ctx, []byte(msg)); err != nil {
						return err
					}
					logger.Error(msg, "err", err)
				}

				if c != nil {
					// msg := "proxy destination reachable, now"
					msg := `
					<div id="proxy-reachable">true</div>
					<span id="status-text">proxy reachable, now</span>
					`
					logger.Debug(msg)
					if err := ws.WriteMessageText(ctx, []byte(msg)); err != nil {
						return err
					}
				}
				<-time.After(1 * time.Second)
			}
		},
	}))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		t, err := templates.NewTemplate()
		if err != nil {
			return
		}

		if err := t.Render(w, templates.HomePage, templates.HomePageArgs{
			Title:        "Live Proxy Home",
			WebsocketURL: fmt.Sprintf("ws://localhost%s/ws", addr),
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	logger.Info("starting http-server", "addr", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		logger.Error("failed to start http-server", "addr", addr, "err", err)
		os.Exit(1)
	}
}
