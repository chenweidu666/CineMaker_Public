package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDatabaseConfig_DSN(t *testing.T) {
	tests := []struct {
		name     string
		config   DatabaseConfig
		expected string
	}{
		{
			name: "SQLite database",
			config: DatabaseConfig{
				Type: "sqlite",
				Path: "./data/drama.db",
			},
			expected: "./data/drama.db",
		},
		{
			name: "MySQL database",
			config: DatabaseConfig{
				Type:     "mysql",
				Host:     "localhost",
				Port:     3306,
				User:     "root",
				Password: "password",
				Database: "drama",
				Charset:  "utf8mb4",
			},
			expected: "root:password@tcp(localhost:3306)/drama?charset=utf8mb4&parseTime=True&loc=Local",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.DSN()
			assert.Equal(t, tt.expected, result)
		})
	}
}
