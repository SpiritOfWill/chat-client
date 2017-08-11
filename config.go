package main

// Config - main config
type Config struct {
	Host string `default:"localhost" toml:"host"`
	Port uint   `default:"8080" toml:"port"`
	// LogLevel string `default:"info" toml:"log_level"`

	// DefaultLogoPath  string `required:"true" toml:"default_logo_path"`
}
