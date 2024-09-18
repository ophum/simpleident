package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Account struct {
	ID        uuid.UUID
	Username  string
	Password  string // hashed
	CreatedAt time.Time
}

var accountStore = NewStore("tmp/accounts.json", accountDeepCopy)

func accountDeepCopy(a *Account) *Account {
	return &Account{
		ID:        a.ID,
		Username:  a.Username,
		Password:  a.Password,
		CreatedAt: a.CreatedAt,
	}
}

func AccountRegisterRoutes(r gin.IRouter) {
	r.GET("/admin/accounts", func(ctx *gin.Context) {
		accounts, err := accountStore.List()
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.HTML(http.StatusOK, "account-list.html.tmpl", gin.H{
			"Accounts": accounts,
		})
	})

	r.GET("/admin/accounts/new", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "account-new.html.tmpl", gin.H{})
	})

	r.POST("/admin/accounts/new", func(ctx *gin.Context) {
		var req AccountNewRequest
		if err := ctx.ShouldBindWith(&req, binding.Form); err != nil {
			ctx.AbortWithError(http.StatusBadRequest, err)
			return
		}

		id, err := uuid.NewV7()
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		if err := accountStore.Add(id, &Account{
			ID:        id,
			Username:  req.Username,
			Password:  string(hashed),
			CreatedAt: time.Now(),
		}); err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		ctx.Redirect(http.StatusSeeOther, "/admin/accounts")
	})

	r.GET("/api/me", func(ctx *gin.Context) {
		authz := ctx.GetHeader("Authorization")
		if !strings.HasPrefix(authz, "Bearer ") {
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}

		token := strings.TrimPrefix(authz, "Bearer ")

		accessToken, err := accessTokenStore.Find(func(a *OAuth2AccessToken) bool {
			return a.AccessToken == token
		})
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		if accessToken.ExpiredAt.Before(time.Now()) {
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}

		account, err := accountStore.Get(accessToken.AccountID)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"id":       account.ID,
			"username": account.Username,
		})

	})
}
