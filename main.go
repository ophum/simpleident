package main

import (
	"embed"
	"html/template"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
)

//go:embed templates/*.tmpl templates/admin/*.tmpl
var f embed.FS

func main() {
	r := gin.Default()
	keyPairs, err := generateSecret(32)
	if err != nil {
		panic(err)
	}
	sessionStore := memstore.NewStore([]byte(keyPairs))
	r.Use(sessions.Sessions("simpleident", sessionStore))

	//FIXME: ディレクトリを分けているが、ファイル名で登録されるので重複できない
	templ := template.Must(template.New("").ParseFS(f, "templates/*.tmpl", "templates/admin/*.tmpl"))
	r.SetHTMLTemplate(templ)

	r.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html.tmpl", gin.H{})
	})

	AccountRegisterRoutes(r)
	OAuth2ClientRegisterRoutes(r)
	if err := r.Run(); err != nil {
		panic(err)
	}
}

type AccountNewRequest struct {
	Username string `form:"username"`
	Password string `form:"password"`
}
