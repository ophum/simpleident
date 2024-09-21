package server

import (
	"errors"
	"net/http"

	"github.com/gin-contrib/sessions"

	"github.com/gin-gonic/gin"
	"github.com/ophum/simpleident/models"
	"golang.org/x/crypto/bcrypt"

	csrf "github.com/utrack/gin-csrf"
	"gorm.io/gorm"
)

type Server struct {
	db                *gorm.DB
	enableAdminServer bool
}

func NewServer(db *gorm.DB, enableAdminServer bool) *Server {
	return &Server{
		db:                db,
		enableAdminServer: enableAdminServer,
	}
}

func (s *Server) RegisterRoutes(r *gin.Engine) {
	if s.enableAdminServer {
		s.registerAdminRoutes(r)
	}

	r.GET("/", handler(s.index))
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

func (s *Server) index(ctx *gin.Context) error {
	ctx.HTML(http.StatusOK, "index", gin.H{})
	return nil
}

func (s *Server) signIn(ctx *gin.Context) error {
	session := sessions.Default(ctx)

	if _, ok := session.Get("account_id").(string); ok {
		ctx.Redirect(http.StatusFound, "/userinfo")
		return nil
	}

	ctx.HTML(http.StatusOK, "sign-in", gin.H{
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

	ctx.HTML(http.StatusOK, "userinfo", gin.H{
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
