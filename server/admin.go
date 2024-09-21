package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ophum/simpleident/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func (s *Server) registerAdminRoutes(router gin.IRouter) {
	admin := new_adminHandler(s.db)

	r := router.Group("/admin")

	r.GET("/accounts", handler(admin.accountList))
	r.GET("/accounts/new", handler(admin.accountNew))
	r.POST("/accounts/new", handler(admin.accountCreate))
}

type adminHandler struct {
	db *gorm.DB
}

func new_adminHandler(db *gorm.DB) *adminHandler {
	return &adminHandler{
		db: db,
	}
}

func (h *adminHandler) accountList(ctx *gin.Context) error {
	var accounts []*models.Account
	if err := h.db.Find(&accounts).Error; err != nil {
		return err
	}

	ctx.HTML(http.StatusOK, "admin/account-list", gin.H{
		"Accounts": accounts,
	})
	return nil
}

func (h *adminHandler) accountNew(ctx *gin.Context) error {
	ctx.HTML(http.StatusOK, "admin/account-new", gin.H{})
	return nil
}

type AccountCreateRequest struct {
	Username string `form:"username"`
	Password string `form:"password"`
}

func (h *adminHandler) accountCreate(ctx *gin.Context) error {
	var req AccountCreateRequest
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
	if err := h.db.Create(&models.Account{
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
