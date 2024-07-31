package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

func test1(ctx *gin.Context) {
	fmt.Println(111)
}

func main() {
	gin.HandlerFunc(test1)(&gin.Context{})
}
