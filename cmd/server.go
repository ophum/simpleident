/*
Copyright © 2024 Takahiro INAGAKI <inagaki0106@gmail.com>
*/
package cmd

import (
	"github.com/gin-gonic/gin"
	"github.com/ophum/simpleident/admin"
	"github.com/ophum/simpleident/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return viper.Unmarshal(&config)
	},
	RunE: serverCommand,
}

func init() {
	rootCmd.AddCommand(serverCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func serverCommand(cmd *cobra.Command, args []string) error {
	var db *gorm.DB
	var err error
	switch config.Database.Driver {
	case "sqlite3":
		db, err = gorm.Open(sqlite.Open(config.Database.DSN))
		if err != nil {
			return err
		}
	}

	r := gin.Default()
	adminServer := admin.NewServer(db)
	if err := adminServer.RegisterTemplates(r); err != nil {
		return err
	}
	adminServer.RegisterRoutes(r)

	server := server.NewServer(db)
	if err := server.RegisterTemplates(r); err != nil {
		return err
	}
	server.RegisterSession(r)
	server.RegisterRoutes(r)

	return r.Run(":8080")
}
