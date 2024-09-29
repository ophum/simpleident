package server

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ophum/simpleident/models"
	csrf "github.com/utrack/gin-csrf"
)

type Oauth2ResponseType string

const (
	Oauth2ResponseTypeCode Oauth2ResponseType = "code"
)

type Oauth2AuthorizeRequest struct {
	ResponseType Oauth2ResponseType `form:"response_type"`
	ClientID     string             `form:"client_id"`
	RedirectURI  string             `form:"redirect_uri"`
	State        string             `form:"state"`
}

func (s *Server) oauth2Authorize(ctx *gin.Context) error {
	session := sessions.Default(ctx)

	var req Oauth2AuthorizeRequest
	if err := ctx.ShouldBind(&req); err != nil {
		return err
	}

	if req.ResponseType != Oauth2ResponseTypeCode {
		return errors.New("invalid response type")
	}

	clientID, err := uuid.Parse(req.ClientID)
	if err != nil {
		return err
	}

	var client models.Oauth2Client
	if err := s.db.Where("id = ?", clientID).First(&client).Error; err != nil {
		return err
	}

	redirectURI := client.CallbackURL
	if req.RedirectURI != "" {
		u, err := url.Parse(req.RedirectURI)
		if err != nil {
			return err
		}

		if u.Fragment != "" {
			return errors.New("invalid redirect_uri")
		}

		if !strings.HasPrefix(u.String(), client.CallbackURL) {
			return errors.New("invalid redirect_uri")
		}

		redirectURI = req.RedirectURI
	}
	session.Set("redirect_uri", redirectURI)
	session.Set("client_id", client.ID.String())
	session.Set("state", req.State)
	if err := session.Save(); err != nil {
		return err
	}

	accountID, ok := session.Get("account_id").(string)
	if !ok {
		v := url.Values{}
		v.Set("return", ctx.Request.URL.String())
		ctx.Redirect(http.StatusSeeOther, "/sign-in?"+v.Encode())
		return nil
	}

	var account models.Account
	if err := s.db.Where("id = ?", accountID).First(&account).Error; err != nil {
		return err
	}

	ctx.HTML(http.StatusOK, "oauth2-authorize", gin.H{
		"CSRFToken":   csrf.GetToken(ctx),
		"Client":      client,
		"Account":     account,
		"RedirectURI": redirectURI,
	})
	return nil
}

func (s *Server) oauth2PostAuthorize(ctx *gin.Context) error {
	session := sessions.Default(ctx)

	clientID, ok := session.Get("client_id").(string)
	if !ok {
		return errors.New("invalid client_id")
	}

	var client models.Oauth2Client
	if err := s.db.Where("id = ?", clientID).First(&client).Error; err != nil {
		return err
	}

	state := session.Get("state").(string)

	redirectURI, err := url.Parse(session.Get("redirect_uri").(string))
	if err != nil {
		return err
	}

	accountID, err := uuid.Parse(session.Get("account_id").(string))
	if err != nil {
		return err
	}

	code, err := generateSecret(32)
	if err != nil {
		return err
	}

	codeID, err := uuid.NewV7()
	if err != nil {
		return err
	}

	if err := s.db.Create(&models.Oauth2Code{
		Model: models.Model{
			ID: codeID,
		},
		Oauth2ClientID: client.ID,
		Code:           code,
		AccountID:      accountID,
	}).Error; err != nil {
		return err
	}

	q := redirectURI.Query()
	q.Set("code", code)
	q.Set("state", state)
	redirectURI.RawQuery = q.Encode()

	ctx.Redirect(http.StatusFound, redirectURI.String())
	return nil
}

type Oauth2GrantType string

const (
	Oauth2GrantTypeAuthorizationCode Oauth2GrantType = "authorization_code"
)

type Oauth2TokenRequest struct {
	GrantType   Oauth2GrantType `form:"grant_type"`
	Code        string          `form:"code"`
	RedirectURI string          `form:"redirect_uri"`
	ClientID    string          `form:"client_id"`
}

// FIXME: error handling
func (s *Server) oauth2PostToken(ctx *gin.Context) error {
	var req Oauth2TokenRequest
	if err := ctx.ShouldBind(&req); err != nil {
		return err
	}

	clientID, err := uuid.Parse(req.ClientID)
	if err != nil {
		return err
	}

	if req.GrantType != Oauth2GrantTypeAuthorizationCode {
		return errors.New("invalid grant_type")
	}

	var code models.Oauth2Code
	if err := s.db.Where("code = ?", req.Code).First(&code).Error; err != nil {
		return err
	}

	if code.Oauth2ClientID != clientID {
		return errors.New("invalid clientID")
	}

	if code.CreatedAt.Add(time.Minute * 5).Before(time.Now()) {
		return errors.New("code expired")
	}

	token, err := generateSecret(20)
	if err != nil {
		return err
	}

	tokenID, err := uuid.NewV7()
	if err != nil {
		return err
	}

	if err := s.db.Create(&models.Oauth2Token{
		Model: models.Model{
			ID: tokenID,
		},
		Oauth2ClientID: clientID,
		Token:          token,
		AccountID:      code.AccountID,
	}).Error; err != nil {
		return err
	}
	ctx.JSON(http.StatusOK, gin.H{
		"access_token":  token,
		"token_type":    "bearer",
		"expires_in":    3600,
		"refresh_token": "",
	})
	return nil
}
