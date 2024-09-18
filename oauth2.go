package main

import (
	"crypto/rand"
	"encoding/gob"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
)

type OAuth2Client struct {
	ID            uuid.UUID
	Name          string
	Description   string
	CallbackURL   string
	ClientSecrets []string
	CreatedAt     time.Time
}

type OAuth2IssuedCode struct {
	ID        uuid.UUID
	ClientID  uuid.UUID
	AccountID uuid.UUID
	Code      string
	ExpiredAt time.Time
}

type OAuth2AccessToken struct {
	ID          uuid.UUID
	ClientID    uuid.UUID
	AccountID   uuid.UUID
	AccessToken string
	ExpiredAt   time.Time
}

func oauth2IssuedCodeDeepCopy(a *OAuth2IssuedCode) *OAuth2IssuedCode {
	return &OAuth2IssuedCode{
		ID:        a.ID,
		ClientID:  a.ClientID,
		AccountID: a.AccountID,
		Code:      a.Code,
		ExpiredAt: a.ExpiredAt,
	}
}

func oauth2AccessTokenDeepCopy(a *OAuth2AccessToken) *OAuth2AccessToken {
	return &OAuth2AccessToken{
		ID:          a.ID,
		ClientID:    a.ClientID,
		AccountID:   a.AccountID,
		AccessToken: a.AccessToken,
		ExpiredAt:   a.ExpiredAt,
	}
}

var oauth2ClientStore = NewStore("tmp/oauth2-clients.json", oauth2ClientDeepCopy)
var issuedCodeStore = NewStore("tmp/oauth2-codes.json", oauth2IssuedCodeDeepCopy)
var accessTokenStore = NewStore("tmp/oauth2-access-tokens.json", oauth2AccessTokenDeepCopy)

func oauth2ClientDeepCopy(a *OAuth2Client) *OAuth2Client {
	return &OAuth2Client{
		ID:            a.ID,
		Name:          a.Name,
		Description:   a.Description,
		CallbackURL:   a.CallbackURL,
		ClientSecrets: a.ClientSecrets,
		CreatedAt:     a.CreatedAt,
	}
}

func init() {
	gob.Register(&OAuth2Client{})
}
func OAuth2ClientRegisterRoutes(r gin.IRouter) {
	r.GET("/admin/oauth2/clients", func(ctx *gin.Context) {
		clients, err := oauth2ClientStore.List()
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.HTML(http.StatusOK, "oauth2-client-list.html.tmpl", gin.H{
			"Clients": clients,
		})
	})

	r.GET("/admin/oauth2/clients/:id", func(ctx *gin.Context) {
		idString := ctx.Param("id")
		id, err := uuid.Parse(idString)
		if err != nil {
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}

		client, err := oauth2ClientStore.Get(id)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		ctx.HTML(http.StatusOK, "oauth2-client-detail.html.tmpl", client)
	})

	r.GET("/admin/oauth2/clients/new", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "oauth2-client-new.html.tmpl", gin.H{})
	})

	r.POST("/admin/oauth2/clients/new", func(ctx *gin.Context) {
		var req OAuth2ClientNewRequest
		if err := ctx.ShouldBindWith(&req, binding.Form); err != nil {
			ctx.AbortWithError(http.StatusBadRequest, err)
			return
		}

		id, err := uuid.NewV7()
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		if err := oauth2ClientStore.Add(id, &OAuth2Client{
			ID:          id,
			Name:        req.Name,
			Description: req.Description,
			CallbackURL: req.CallbackURL,
			CreatedAt:   time.Now(),
		}); err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		ctx.Redirect(http.StatusSeeOther, "/admin/oauth2/clients")
	})

	r.POST("/admin/oauth2/clients/:id/create-secret", func(ctx *gin.Context) {
		id, err := parseIDFromParam(ctx)
		if err != nil {
			ctx.AbortWithError(http.StatusBadRequest, err)
			return
		}

		client, err := oauth2ClientStore.Get(id)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		secret, err := generateSecret(64)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		client.ClientSecrets = append(client.ClientSecrets, secret)

		if err := oauth2ClientStore.Update(client.ID, client); err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		ctx.Redirect(http.StatusSeeOther, "/admin/oauth2/clients/"+client.ID.String())
	})

	r.GET("/oauth2/authorize", func(ctx *gin.Context) {
		var req AuthorizeRequest
		if err := ctx.ShouldBindWith(&req, binding.Form); err != nil {
			ctx.AbortWithError(http.StatusBadRequest, err)
			return
		}

		session := sessions.Default(ctx)
		if req.ResponseType != "code" {
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}

		if req.RedirectURI != "" {
			u, err := url.Parse(req.RedirectURI)
			if err != nil {
				ctx.AbortWithError(http.StatusUnprocessableEntity, err)
				return
			}
			if u.Fragment != "" {
				ctx.AbortWithStatus(http.StatusUnprocessableEntity)
				return
			}
			session.Set("redirect_uri", req.RedirectURI)
		}
		id, err := uuid.Parse(req.ClientID)
		if err != nil {
			ctx.AbortWithError(http.StatusUnprocessableEntity, err)
			return
		}

		client, err := oauth2ClientStore.Get(id)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		session.Set("client", client)
		session.Set("state", req.State)
		session.Save()

		ctx.HTML(http.StatusOK, "oauth2-authorize.html.tmpl", gin.H{
			"Client": client,
		})
	})

	r.POST("/oauth2/authorize", func(ctx *gin.Context) {
		session := sessions.Default(ctx)

		client, ok := session.Get("client").(*OAuth2Client)
		if !ok {
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}
		state := session.Get("state").(string)

		cb, err := url.Parse(client.CallbackURL)
		if err != nil {
			ctx.AbortWithError(http.StatusUnprocessableEntity, err)
			return
		}
		redirectURI := session.Get("redirect_uri").(string)
		if redirectURI != "" {
			sr, err := url.Parse(redirectURI)
			if err != nil {
				ctx.AbortWithError(http.StatusUnprocessableEntity, err)
				return
			}

			cb.RawQuery = sr.RawQuery
			if cb != sr {
				ctx.AbortWithStatus(http.StatusUnprocessableEntity)
				return
			}
		}

		serverError := ErrorResponseQuery{"server_error", state}
		code, err := generateSecret(32)
		if err != nil {
			cb.RawQuery = serverError.Encode()
			ctx.Redirect(http.StatusFound, cb.String())
			return
		}

		issedCodeID, err := uuid.NewV7()
		if err != nil {
			cb.RawQuery = serverError.Encode()
			ctx.Redirect(http.StatusFound, cb.String())
			return
		}
		if err := issuedCodeStore.Add(issedCodeID, &OAuth2IssuedCode{
			ID:        issedCodeID,
			ClientID:  client.ID,
			Code:      code,
			AccountID: uuid.Must(uuid.Parse("0191f70a-d3a2-7970-8287-d2bc4264c3c5")),
			ExpiredAt: time.Now().Add(time.Minute),
		}); err != nil {
			cb.RawQuery = serverError.Encode()
			ctx.Redirect(http.StatusFound, cb.String())
			return
		}

		q := url.Values{}
		q.Set("code", code)
		q.Set("state", state)
		cb.RawQuery = q.Encode()
		ctx.Redirect(http.StatusFound, cb.String())
	})

	r.POST("/oauth2/token", func(ctx *gin.Context) {
		var req OAuth2TokenRequest
		if err := ctx.ShouldBind(&req); err != nil {
			ctx.AbortWithError(http.StatusBadRequest, err)
			return
		}

		if req.GrantType != "authorization_code" {
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}

		issuedCode, err := issuedCodeStore.Find(func(a *OAuth2IssuedCode) bool {
			return a.Code == req.Code
		})
		if err != nil {
			log.Println("code not found")
			ctx.AbortWithError(http.StatusForbidden, err)
			return
		}

		if issuedCode.ExpiredAt.Before(time.Now()) {
			log.Println("code expired")
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}

		if req.ClientID != issuedCode.ClientID.String() {
			log.Println("invalid client_id")
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}

		token, err := generateSecret(20)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		tokenID, err := uuid.NewV7()
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		if err := accessTokenStore.Add(tokenID, &OAuth2AccessToken{
			ID:          tokenID,
			ClientID:    issuedCode.ClientID,
			AccountID:   issuedCode.AccountID,
			AccessToken: token,
			ExpiredAt:   time.Now().Add(time.Hour),
		}); err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"access_token":  token,
			"token_type":    "bearer",
			"expires_in":    3600,
			"refresh_token": "",
		})

	})
}

type OAuth2ClientNewRequest struct {
	Name        string `form:"name"`
	Description string `form:"description"`
	CallbackURL string `form:"callback_url"`
}

type OAuth2TokenRequest struct {
	GrantType   string `form:"grant_type"`
	Code        string `form:"code"`
	RedirectURI string `form:"redirect_uri"`
	ClientID    string `form:"client_id"`
}

func parseIDFromParam(ctx *gin.Context) (uuid.UUID, error) {
	idString := ctx.Param("id")
	id, err := uuid.Parse(idString)
	if err != nil {
		return uuid.UUID{}, err
	}
	return id, nil
}

func generateSecret(length int) (string, error) {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	secret := ""
	for _, v := range b {
		secret += string(letters[int(v)%len(letters)])
	}
	return secret, nil

}

type AuthorizeRequest struct {
	ResponseType string `form:"response_type"`
	ClientID     string `form:"client_id"`
	State        string `form:"state"`
	RedirectURI  string `form:"redirect_uri"`
}

type ErrorResponseQuery struct {
	Error string
	State string
}

func (e *ErrorResponseQuery) Encode() string {
	v := url.Values{}
	v.Set("error", e.Error)
	if e.State != "" {
		v.Set("state", e.State)
	}
	return v.Encode()
}
