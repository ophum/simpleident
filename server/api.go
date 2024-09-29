package server

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ophum/simpleident/models"
)

func (s *Server) apiGetUserinfo(ctx *gin.Context) error {
	bearerToken := ctx.GetHeader("Authorization")
	token, ok := strings.CutPrefix(bearerToken, "Bearer ")
	if !ok {
		return errors.New("invalid authorization")
	}

	var oauth2Token models.Oauth2Token
	if err := s.db.Preload("Account").
		Where("token = ?", token).
		First(&oauth2Token).Error; err != nil {
		return err
	}

	if oauth2Token.CreatedAt.Add(time.Hour).Before(time.Now()) {
		return errors.New("token expired")
	}

	ctx.JSON(http.StatusOK, gin.H{
		"id":       oauth2Token.Account.ID,
		"username": oauth2Token.Account.Username,
	})
	return nil
}
