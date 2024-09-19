package admin

import (
	"embed"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ophum/simpleident/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

type Server struct {
	db *gorm.DB
}

func NewServer(db *gorm.DB) *Server {
	return &Server{
		db: db,
	}
}

func (s *Server) RegisterTemplates(engine *gin.Engine) error {
	templ, err := template.New("").ParseFS(templateFS, "templates/*.tmpl")
	if err != nil {
		return err
	}

	engine.SetHTMLTemplate(templ)
	return nil
}

func (s *Server) RegisterRoutes(router gin.IRouter) {
	r := router.Group("/admin")

	r.GET("/accounts", handler(s.accountList))
	r.GET("/accounts/new", handler(s.accountNew))
	r.POST("/accounts/new", handler(s.accountCreate))
}

func handler(fn func(ctx *gin.Context) error) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if err := fn(ctx); err != nil {
			_ = ctx.Error(err)
			ctx.Abort()
			return
		}
	}
}

func (s *Server) accountList(ctx *gin.Context) error {
	var accounts []*models.Account
	if err := s.db.Find(&accounts).Error; err != nil {
		return err
	}

	ctx.HTML(http.StatusOK, "account-list.html.tmpl", gin.H{
		"Accounts": accounts,
	})
	return nil
}

func (s *Server) accountNew(ctx *gin.Context) error {
	ctx.HTML(http.StatusOK, "account-new.html.tmpl", gin.H{})
	return nil
}

type AccountCreateRequest struct {
	Username string `form:"username"`
	Password string `form:"password"`
}

func (s *Server) accountCreate(ctx *gin.Context) error {
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
