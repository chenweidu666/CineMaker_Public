package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Storage  StorageConfig  `mapstructure:"storage"`
	AI       AIConfig       `mapstructure:"ai"`
	JWT      JWTConfig      `mapstructure:"jwt"`
}

type JWTConfig struct {
	Secret          string `mapstructure:"secret"`
	AccessTokenTTL  int    `mapstructure:"access_token_ttl"`  // minutes, default 30
	RefreshTokenTTL int    `mapstructure:"refresh_token_ttl"` // days, default 7
}

type AppConfig struct {
	Name            string `mapstructure:"name"`
	Version         string `mapstructure:"version"`
	Debug           bool   `mapstructure:"debug"`
	Language        string `mapstructure:"language"` // zh 或 en
	DefaultAdminEmail    string `mapstructure:"default_admin_email"`    // 默认管理员邮箱，首次启动时创建
	DefaultAdminPassword string `mapstructure:"default_admin_password"` // 默认管理员密码
}

type ServerConfig struct {
	Port         int      `mapstructure:"port"`
	Host         string   `mapstructure:"host"`
	CORSOrigins  []string `mapstructure:"cors_origins"`
	ReadTimeout  int      `mapstructure:"read_timeout"`
	WriteTimeout int      `mapstructure:"write_timeout"`
}

type DatabaseConfig struct {
	Type     string `mapstructure:"type"` // sqlite, mysql
	Path     string `mapstructure:"path"` // SQLite数据库文件路径
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	Charset  string `mapstructure:"charset"`
	MaxIdle  int    `mapstructure:"max_idle"`
	MaxOpen  int    `mapstructure:"max_open"`
}

type StorageConfig struct {
	Type      string    `mapstructure:"type"`       // local, cos
	LocalPath string    `mapstructure:"local_path"` // 本地存储路径
	BaseURL   string    `mapstructure:"base_url"`   // 访问URL前缀
	COS       COSConfig `mapstructure:"cos"`
}

type COSConfig struct {
	Bucket    string `mapstructure:"bucket"`
	Region    string `mapstructure:"region"`
	SecretID  string `mapstructure:"secret_id"`
	SecretKey string `mapstructure:"secret_key"`
	CDNURL    string `mapstructure:"cdn_url"`
}

type AIConfig struct {
	DefaultTextProvider  string `mapstructure:"default_text_provider"`
	DefaultImageProvider string `mapstructure:"default_image_provider"`
	DefaultVideoProvider string `mapstructure:"default_video_provider"`
}

const defaultJWTSecret = "cinemaker-default-jwt-secret-change-me"

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	viper.AutomaticEnv()

	// 环境变量绑定：COS 密钥优先从环境变量读取
	viper.BindEnv("storage.cos.secret_id", "COS_SECRET_ID")
	viper.BindEnv("storage.cos.secret_key", "COS_SECRET_KEY")
	viper.BindEnv("storage.cos.bucket", "COS_BUCKET")
	viper.BindEnv("storage.cos.region", "COS_REGION")
	viper.BindEnv("storage.cos.cdn_url", "COS_CDN_URL")
	viper.BindEnv("storage.type", "STORAGE_TYPE")

	viper.SetDefault("jwt.secret", defaultJWTSecret)
	viper.SetDefault("jwt.access_token_ttl", 30)
	viper.SetDefault("jwt.refresh_token_ttl", 7)

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := config.validateJWT(); err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Config) validateJWT() error {
	if c.JWT.Secret == defaultJWTSecret && !c.App.Debug {
		return fmt.Errorf("SECURITY: jwt.secret is still the default value; set a strong secret in config.yaml (minimum 32 bytes)")
	}
	if len(c.JWT.Secret) < 32 {
		return fmt.Errorf("SECURITY: jwt.secret must be at least 32 bytes long (current: %d)", len(c.JWT.Secret))
	}
	return nil
}

func (c *DatabaseConfig) DSN() string {
	if c.Type == "sqlite" {
		return c.Path
	}
	// MySQL DSN
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
		c.Charset,
	)
}
