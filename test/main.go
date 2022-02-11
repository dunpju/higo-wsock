package main

import (
	"fmt"
	"github.com/dengpju/higo-wsock/wsock"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	r := gin.Default()
	r.Use(wsock.WsConnMiddleWare(r))
	r.Handle("GET", "/conn", func(context *gin.Context) {
		fmt.Println("hhh")
		//wsock.WsRespString("ttt")
		//wsock.WsConn(context).Conn().WriteMessage(websocket.TextMessage, []byte("ttt"))
		context.String(200, "fff")
	})
	err := r.Run(":8080")
	if err != nil {
		log.Fatalln(err)
	}
}
