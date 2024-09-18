package main

import (
	"embed"
	"flag"
	"html/template"
	"net/http"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

//go:embed templates/*.tmpl templates/admin/*.tmpl
var f embed.FS

var config Config

type Config struct {
	AdminBasicAuthAccounts gin.Accounts `yaml:"adminBasicAuthAccounts"`
}

func init() {
	configPath := flag.String("config", "config.yaml", "config.yaml")
	flag.Parse()

	f, err := os.Open(*configPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if err := yaml.NewDecoder(f).Decode(&config); err != nil {
		panic(err)
	}
}

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

	AdminRegisterRoutes(r, config.AdminBasicAuthAccounts)
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
