package main

import (
    "log"

    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
)

func main() {
    server := gin.Default()
    server.Use(cors.Default())
    server.GET("/", func(ctx *gin.Context) {
        ctx.String(200, "get request")
    })
    server.POST("/", func(ctx *gin.Context) {
        ctx.String(200, "post request")
    })
    if err := server.Run(":8080"); err != nil {
        log.Fatal(err)
    }
}
