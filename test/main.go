package main

import (
	"fmt"
	"github.com/dengpju/higo-wsock/wsock"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	r := gin.Default()
	r.Use(wsock.WsConnMiddleWare())
	r.Handle("GET", "/conn", func(context *gin.Context) {
		context.String(200, "test ws conn")
	})
	fmt.Println(r.Routes())
	err := r.Run(":8080")
	if err != nil {
		log.Fatalln(err)
	}
}
