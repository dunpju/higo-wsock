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
		wsock.WsRespString("test ws conn")
		fmt.Println("hhh")
		//context.String(200, "fff")
	})
	err := r.Run(":8080")
	if err != nil {
		log.Fatalln(err)
	}
}
