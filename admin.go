package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AdminRegisterRoutes(r gin.IRouter, basicAuthAccounts gin.Accounts) {
	r.Use(gin.BasicAuth(basicAuthAccounts))

	r.GET("/admin/", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "authroized")
	})
}
