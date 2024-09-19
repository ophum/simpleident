package server

import (
	"embed"
	"errors"
	"html/template"
	"net/http"

	"github.com/gin-contrib/sessions"

	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/ophum/simpleident/models"
	"golang.org/x/crypto/bcrypt"

	csrf "github.com/utrack/gin-csrf"
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

func (s *Server) RegisterSession(engine *gin.Engine) {
	store := cookie.NewStore([]byte("secret"))
	engine.Use(sessions.Sessions("simpleident", store))
	engine.Use(csrf.Middleware(csrf.Options{
		Secret: "secret",
		ErrorFunc: func(ctx *gin.Context) {
			ctx.String(http.StatusBadRequest, "CSRF token mismatch")
			ctx.Abort()
		},
	}))
}

func (s *Server) RegisterTemplates(engine *gin.Engine) error {
	templ, err := template.New("").ParseFS(templateFS, "templates/*.tmpl")
	if err != nil {
		return err
	}

	engine.SetHTMLTemplate(templ)
	return nil
}

func (s *Server) RegisterRoutes(r gin.IRouter) {
	r.GET("/sign-in", handler(s.signIn))
	r.POST("/sign-in", handler(s.signInProcess))
	r.GET("/userinfo", handler(s.userinfo))
	r.POST("/sign-out", handler(s.signOut))
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

func (s *Server) signIn(ctx *gin.Context) error {
	session := sessions.Default(ctx)

	if _, ok := session.Get("account_id").(string); ok {
		ctx.Redirect(http.StatusFound, "/userinfo")
		return nil
	}

	ctx.HTML(http.StatusOK, "sign-in.html.tmpl", gin.H{
		"CSRFToken": csrf.GetToken(ctx),
	})
	return nil
}

type SignInRequest struct {
	Username string `form:"username"`
	Password string `form:"password"`
}

func (s *Server) signInProcess(ctx *gin.Context) error {
	var req SignInRequest
	if err := ctx.ShouldBind(&req); err != nil {
		return err
	}

	var account models.Account
	if err := s.db.Where("username = ?", req.Username).First(&account).Error; err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(req.Password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return errors.New("unauthorized")
		}
		return err
	}

	session := sessions.Default(ctx)

	session.Set("account_id", account.ID.String())
	session.Save()

	ctx.Redirect(http.StatusFound, "/userinfo")
	return nil
}

func (s *Server) userinfo(ctx *gin.Context) error {
	session := sessions.Default(ctx)

	accountID, ok := session.Get("account_id").(string)
	if !ok {
		ctx.Redirect(http.StatusFound, "/sign-in")
		return nil
	}

	var account models.Account
	if err := s.db.Where("id = ?", accountID).First(&account).Error; err != nil {
		return err
	}

	ctx.HTML(http.StatusOK, "userinfo.html.tmpl", gin.H{
		"Account":   account,
		"CSRFToken": csrf.GetToken(ctx),
	})
	return nil
}

func (s *Server) signOut(ctx *gin.Context) error {
	session := sessions.Default(ctx)

	session.Clear()
	session.Save()

	ctx.Redirect(http.StatusFound, "/sign-in")
	return nil
}
