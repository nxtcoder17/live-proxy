package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nxtcoder17/live-proxy/pkg/log"
	"github.com/nxtcoder17/live-proxy/pkg/websocket"
	"github.com/nxtcoder17/live-proxy/templates"
)

func checkServerAlive(addr string) error {
	_, err := net.DialTimeout("tcp", addr, 1*time.Second)
	return err
}

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

	r.Get("/_live-proxy/ws", websocket.HttpHandler(websocket.Args{
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
				if err := checkServerAlive(proxyAddr); err != nil {
					// msg := "proxy destination not reachable"
					msg := `
					<div id="proxy-reachable">false</div>
					<div id="status-icon" class="w-24 h-24">
					  <svg fill="#000000" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><g id="SVGRepo_bgCarrier" stroke-width="0"></g><g id="SVGRepo_tracerCarrier" stroke-linecap="round" stroke-linejoin="round"></g><g id="SVGRepo_iconCarrier"> <path d="M12,16 L12,14.5 C12,13.6715729 12.6715729,13 13.5,13 L19.5,13 C20.3284271,13 21,13.6715729 21,14.5 L21,18.5 C21,19.3284271 20.3284271,20 19.5,20 L18.5,20 C18.7761424,20 19,20.2238576 19,20.5 C19,20.7761424 18.7761424,21 18.5,21 L14.5,21 C14.2238576,21 14,20.7761424 14,20.5 C14,20.2238576 14.2238576,20 14.5,20 L13.5,20 C12.6715729,20 12,19.3284271 12,18.5 L12,17 L10.5,17 C10.2238576,17 10,16.7761424 10,16.5 C10,16.2238576 10.2238576,16 10.5,16 L12,16 Z M7,11 L5.5,11 C5.22385763,11 5,10.7761424 5,10.5 C5,10.2238576 5.22385763,10 5.5,10 L4.5,10 C3.67157288,10 3,9.32842712 3,8.5 L3,4.5 C3,3.67157288 3.67157288,3 4.5,3 L10.5,3 C11.3284271,3 12,3.67157288 12,4.5 L12,8.5 C12,9.32842712 11.3284271,10 10.5,10 L9.5,10 C9.77614237,10 10,10.2238576 10,10.5 C10,10.7761424 9.77614237,11 9.5,11 L8,11 L8,13.5 C8,13.7761424 7.77614237,14 7.5,14 C7.22385763,14 7,13.7761424 7,13.5 L7,11 Z M7.5,15.7928932 L9.14644661,14.1464466 C9.34170876,13.9511845 9.65829124,13.9511845 9.85355339,14.1464466 C10.0488155,14.3417088 10.0488155,14.6582912 9.85355339,14.8535534 L8.20710678,16.5 L9.85355339,18.1464466 C10.0488155,18.3417088 10.0488155,18.6582912 9.85355339,18.8535534 C9.65829124,19.0488155 9.34170876,19.0488155 9.14644661,18.8535534 L7.5,17.2071068 L5.85355339,18.8535534 C5.65829124,19.0488155 5.34170876,19.0488155 5.14644661,18.8535534 C4.95118446,18.6582912 4.95118446,18.3417088 5.14644661,18.1464466 L6.79289322,16.5 L5.14644661,14.8535534 C4.95118446,14.6582912 4.95118446,14.3417088 5.14644661,14.1464466 C5.34170876,13.9511845 5.65829124,13.9511845 5.85355339,14.1464466 L7.5,15.7928932 L7.5,15.7928932 Z M4,4.5 L4,8.5 C4,8.77614237 4.22385763,9 4.5,9 L10.5,9 C10.7761424,9 11,8.77614237 11,8.5 L11,4.5 C11,4.22385763 10.7761424,4 10.5,4 L4.5,4 C4.22385763,4 4,4.22385763 4,4.5 Z M13,14.5 L13,18.5 C13,18.7761424 13.2238576,19 13.5,19 L19.5,19 C19.7761424,19 20,18.7761424 20,18.5 L20,14.5 C20,14.2238576 19.7761424,14 19.5,14 L13.5,14 C13.2238576,14 13,14.2238576 13,14.5 Z"></path> </g></svg>
					</div>
					<span id="status-text">proxy destination not reachable</span>
					`
					if err := ws.WriteMessageText(ctx, []byte(msg)); err != nil {
						return err
					}
					logger.Debug("[ERR]: proxy destination liveness check failed", "err", err)
					<-time.After(1 * time.Second)
					continue
				}

				logger.Info("proxy destination is reachable now")
				// msg := "proxy destination reachable, now"
				msg := `
					<div id="status-icon" class="w-24 h-24">
					  <svg fill="#000000" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><g id="SVGRepo_bgCarrier" stroke-width="0"></g><g id="SVGRepo_tracerCarrier" stroke-linecap="round" stroke-linejoin="round"></g><g id="SVGRepo_iconCarrier"> <path d="M18.5,20 C18.7761424,20 19,20.2238576 19,20.5 C19,20.7761424 18.7761424,21 18.5,21 L14.5,21 C14.2238576,21 14,20.7761424 14,20.5 C14,20.2238576 14.2238576,20 14.5,20 L13.5,20 C12.6715729,20 12,19.3284271 12,18.5 L12,17 L9.5,17 C8.11928813,17 7,15.8807119 7,14.5 L7,11 L5.5,11 C5.22385763,11 5,10.7761424 5,10.5 C5,10.2238576 5.22385763,10 5.5,10 L4.5,10 C3.67157288,10 3,9.32842712 3,8.5 L3,4.5 C3,3.67157288 3.67157288,3 4.5,3 L10.5,3 C11.3284271,3 12,3.67157288 12,4.5 L12,8.5 C12,9.32842712 11.3284271,10 10.5,10 L9.5,10 C9.77614237,10 10,10.2238576 10,10.5 C10,10.7761424 9.77614237,11 9.5,11 L8,11 L8,14.5 C8,15.3284271 8.67157288,16 9.5,16 L12,16 L12,14.5 C12,13.6715729 12.6715729,13 13.5,13 L19.5,13 C20.3284271,13 21,13.6715729 21,14.5 L21,18.5 C21,19.3284271 20.3284271,20 19.5,20 L18.5,20 Z M4,4.5 L4,8.5 C4,8.77614237 4.22385763,9 4.5,9 L10.5,9 C10.7761424,9 11,8.77614237 11,8.5 L11,4.5 C11,4.22385763 10.7761424,4 10.5,4 L4.5,4 C4.22385763,4 4,4.22385763 4,4.5 Z M13,14.5 L13,18.5 C13,18.7761424 13.2238576,19 13.5,19 L19.5,19 C19.7761424,19 20,18.7761424 20,18.5 L20,14.5 C20,14.2238576 19.7761424,14 19.5,14 L13.5,14 C13.2238576,14 13,14.2238576 13,14.5 Z"></path> </g></svg>
					</div>
					<div id="proxy-reachable">true</div>
					<span id="status-text">proxy reachable, now</span>
					`
				if err := ws.WriteMessageText(ctx, []byte(msg)); err != nil {
					return err
				}
				<-time.After(1 * time.Second)
			}
		},
	}))

	r.Get("/_live-proxy/", func(w http.ResponseWriter, r *http.Request) {
		t, err := templates.NewTemplate()
		if err != nil {
			return
		}

		if err := t.Render(w, templates.HomePage, templates.HomePageArgs{
			Title:         "Live Proxy Home",
			WebsocketPath: "/_live-proxy/ws",
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	proxyURL, err := url.Parse(fmt.Sprintf("http://%s", proxyAddr))
	if err != nil {
		panic(err)
	}

	landingURL, err := url.Parse(fmt.Sprintf("http://%s/_live-proxy", addr))
	if err != nil {
		panic(err)
	}

	realproxy := httputil.NewSingleHostReverseProxy(proxyURL)
	landingproxy := httputil.NewSingleHostReverseProxy(landingURL)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		if err := checkServerAlive(proxyAddr); err != nil {
			landingproxy.ServeHTTP(w, r)
			return
		}
		realproxy.ServeHTTP(w, r)
	})

	logger.Info("starting http-server", "addr", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		logger.Error("failed to start http-server", "addr", addr, "err", err)
		os.Exit(1)
	}
}
