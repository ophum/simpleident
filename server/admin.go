package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ophum/simpleident/models"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) registerAdminRoutes(router gin.IRouter) {
	r := router.Group("/admin")

	r.GET("/accounts", handler(s.adminAccountList))
	r.GET("/accounts/new", handler(s.adminAccountNew))
	r.POST("/accounts/new", handler(s.adminAccountCreate))

	r.GET("/oauth2/clients", handler(s.adminOauth2ClientList))
	r.GET("/oauth2/clients/new", handler(s.adminOauth2ClientNew))
	r.POST("/oauth2/clients/new", handler(s.adminOauth2ClientCreate))
	r.GET("/oauth2/clients/:id", handler(s.adminOauth2ClientDetail))
	r.POST("/oauth2/clients/:id/generate-secret", handler(s.adminOauth2ClientGenerateSecret))
}

func (s *Server) adminAccountList(ctx *gin.Context) error {
	var accounts []*models.Account
	if err := s.db.Find(&accounts).Error; err != nil {
		return err
	}

	ctx.HTML(http.StatusOK, "admin/account-list", gin.H{
		"Accounts": accounts,
	})
	return nil
}

func (s *Server) adminAccountNew(ctx *gin.Context) error {
	ctx.HTML(http.StatusOK, "admin/account-new", gin.H{})
	return nil
}

type AdminAccountCreateRequest struct {
	Username string `form:"username"`
	Password string `form:"password"`
}

func (s *Server) adminAccountCreate(ctx *gin.Context) error {
	var req AdminAccountCreateRequest
	if err := ctx.ShouldBind(&req); err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return err
	}
	if err := s.db.Create(&models.Account{
		Model: models.Model{
			ID: id,
		},
		Username: req.Username,
		Password: string(hash),
	}).Error; err != nil {
		return err
	}

	ctx.Redirect(http.StatusSeeOther, "/admin/accounts")
	return nil
}
