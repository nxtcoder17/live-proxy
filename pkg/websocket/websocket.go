package websocket

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"slices"

	"nhooyr.io/websocket"
)

type Conn struct {
	*websocket.Conn
	// Type   websocket.MessageType
	// Writer io.WriteCloser
	// Reader io.Reader
}

func (c *Conn) WriteMessageText(ctx context.Context, msg []byte) error {
	return c.Write(ctx, websocket.MessageText, msg)
}

func (c *Conn) Close(reason string) error {
	return c.Conn.Close(websocket.StatusGoingAway, reason)
}

type WSHandler func(ctx Context, ws Conn) error

type Args struct {
	Logger         *slog.Logger
	Handler        WSHandler
	SubProtocols   []string
	OriginPatterns []string
}

type Context struct {
	context.Context
	ConnectionID int
}

func HttpHandler(args Args) http.HandlerFunc {
	count := 1
	connections := map[int]struct{}{}

	return func(w http.ResponseWriter, r *http.Request) {
		args.Logger.Info("NEW CONNECTION", "count", count)
		connectionID := count
		connections[connectionID] = struct{}{}
		defer func() {
			args.Logger.Info("CONNECTION CLOSED", "connId", connectionID)
			delete(connections, connectionID)
		}()
		count += 1
		args.Logger.Info("# Active Connections", "num", len(connections))
		c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			Subprotocols:   args.SubProtocols,
			OriginPatterns: args.OriginPatterns,
		})
		if err != nil {
			args.Logger.Debug("while accepting websocket connections", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer c.CloseNow()

		ctx := c.CloseRead(r.Context())

		if args.SubProtocols != nil && slices.Contains(args.SubProtocols, c.Subprotocol()) {
			err := fmt.Errorf(`client must speak the nxtcoder17.me-web subprotocol`)
			args.Logger.Debug("while checking websocket subprotocol", "error", err)
			c.Close(websocket.StatusPolicyViolation, err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		wctx := Context{Context: ctx, ConnectionID: count}
		if err := args.Handler(wctx, Conn{Conn: c}); err != nil {
			args.Logger.Error("while creating socket writer", "err", err)
			return
		}
	}
}
