package cmd

type Config struct {
	Database *ConfigDatabase
}

type ConfigDatabase struct {
	Driver string
	DSN    string
}
