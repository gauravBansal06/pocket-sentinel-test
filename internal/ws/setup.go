package ws

import (
	"net/http"

	"github.com/LambdatestIncPrivate/exemplar/internal/middleware"
	"github.com/LambdatestIncPrivate/exemplar/pkg/lumber"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var logger lumber.Logger

// RegisterRoutes registers routes and logger in the provided router
func RegisterRoutes(router *gin.RouterGroup, initializedLogger lumber.Logger) {
	logger = initializedLogger
	router.GET("/ws", middleware.Authentication(), middleware.WsValidator(), establishSocket)
	logger.Infof("ws routes injected")
}

var upgrader = websocket.Upgrader{} // use default options

func wshandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Errorf("Failed to set websocket upgrade: %+v", err)
		return
	}
	// ready to accept ws messages

	// sample echo implementation
	for {
		t, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		conn.WriteMessage(t, msg)
	}
}

// Websocket upgrade endpoint
func establishSocket(c *gin.Context) {
	wshandler(c.Writer, c.Request)
}
