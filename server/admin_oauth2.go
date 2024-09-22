package server

import (
	"crypto/rand"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ophum/simpleident/models"
	csrf "github.com/utrack/gin-csrf"
)

func (s *Server) adminOauth2ClientDetail(ctx *gin.Context) error {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return err
	}

	var client models.Oauth2Client
	if err := s.db.Preload("ClientSecrets").
		Where("id = ?", id).
		First(&client).Error; err != nil {
		return err
	}

	ctx.HTML(http.StatusOK, "admin/oauth2-client-detail", gin.H{
		"Client":    client,
		"CSRFToken": csrf.GetToken(ctx),
	})
	return nil
}

func (s *Server) adminOauth2ClientList(ctx *gin.Context) error {
	var clients []*models.Oauth2Client
	if err := s.db.Preload("ClientSecrets").
		Find(&clients).Error; err != nil {
		return err
	}

	ctx.HTML(http.StatusOK, "admin/oauth2-client-list", gin.H{
		"Clients": clients,
	})
	return nil
}

func (s *Server) adminOauth2ClientNew(ctx *gin.Context) error {
	ctx.HTML(http.StatusOK, "admin/oauth2-client-new", gin.H{
		"CSRFToken": csrf.GetToken(ctx),
	})
	return nil
}

type AdminOauth2ClientCreateRequest struct {
	Name        string `form:"name"`
	Description string `form:"description"`
	CallbackURL string `form:"callback_url"`
}

func (s *Server) adminOauth2ClientCreate(ctx *gin.Context) error {
	var req AdminOauth2ClientCreateRequest
	if err := ctx.ShouldBind(&req); err != nil {
		return err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return err
	}
	if err := s.db.Create(&models.Oauth2Client{
		Model: models.Model{
			ID: id,
		},
		Name:        req.Name,
		Description: req.Description,
		CallbackURL: req.CallbackURL,
	}).Error; err != nil {
		return err
	}

	ctx.Redirect(http.StatusSeeOther, "/admin/oauth2/clients")
	return nil
}

func (s *Server) adminOauth2ClientGenerateSecret(ctx *gin.Context) error {
	clientID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return err
	}

	var client models.Oauth2Client
	if err := s.db.Where("id = ?", clientID).First(&client).Error; err != nil {
		return err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return err
	}

	secret, err := generateSecret(64)
	if err != nil {
		return err
	}

	if err := s.db.Create(&models.Oauth2ClientSecret{
		Model: models.Model{
			ID: id,
		},
		Oauth2ClientID: clientID,
		Secret:         secret,
	}).Error; err != nil {
		return err
	}
	ctx.Redirect(http.StatusSeeOther, "/admin/oauth2/clients/"+clientID.String())
	return nil
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
