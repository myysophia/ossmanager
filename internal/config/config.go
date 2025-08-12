package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig
	JWT      JWTConfig
	Database DatabaseConfig
	Log      LogConfig
	OSS      OSSConfig
}

type AppConfig struct {
	Name             string
	Env              string
	Host             string
	Port             int
	ReadTimeout      int    `mapstructure:"read_timeout"`
	WriteTimeout     int    `mapstructure:"write_timeout"`
	IdleTimeout      int    `mapstructure:"idle_timeout"`
	UploadTempDir    string `mapstructure:"upload_temp_dir"`
	MaxFileSize      int64  `mapstructure:"max_file_size"`
	Workers          int    `mapstructure:"workers"`           // MD5计算的工作协程数量
	ChunkConcurrency int    `mapstructure:"chunk_concurrency"` // 分片上传并发量
}

type JWTConfig struct {
	SecretKey string `mapstructure:"secret_key"`
	ExpiresIn int    `mapstructure:"expires_in"`
	Issuer    string
}

type DatabaseConfig struct {
	Driver          string
	Host            string
	Port            int
	Username        string
	Password        string
	DBName          string `mapstructure:"dbname"`
	SSLMode         string `mapstructure:"sslmode"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

type LogConfig struct {
	Level    string
	Format   string
	Output   string
	FilePath string `mapstructure:"file_path"`
}

type OSSConfig struct {
	AliyunOSS    AliyunOSSConfig    `mapstructure:"aliyun_oss"`
	AWSS3        AWSS3Config        `mapstructure:"aws_s3"`
	CloudflareR2 CloudflareR2Config `mapstructure:"cloudflare_r2"`
}

type AliyunOSSConfig struct {
	AccessKeyID        string `mapstructure:"access_key_id"`
	AccessKeySecret    string `mapstructure:"access_key_secret"`
	Endpoint           string
	Bucket             string
	Region             string
	UploadDir          string `mapstructure:"upload_dir"`
	URLExpireTime      int    `mapstructure:"url_expire_time"`
	TransferAccelerate struct {
		Enabled bool   `mapstructure:"enabled"`
		Type    string `mapstructure:"type"` // "global" 或 "overseas"
	} `mapstructure:"transfer_accelerate"`
	FunctionCompute struct {
		Enabled         bool   `mapstructure:"enabled"`
		Endpoint        string `mapstructure:"endpoint"`
		APIVersion      string `mapstructure:"api_version"`
		AccessKeyID     string `mapstructure:"access_key_id"`
		AccessKeySecret string `mapstructure:"access_key_secret"`
		ServiceName     string `mapstructure:"service_name"`
		FunctionName    string `mapstructure:"function_name"`
	} `mapstructure:"function_compute"`
}

type AWSS3Config struct {
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	Region          string
	Bucket          string
	UploadDir       string `mapstructure:"upload_dir"`
	URLExpireTime   int    `mapstructure:"url_expire_time"`
}

type CloudflareR2Config struct {
	AccountID       string `mapstructure:"account_id"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	Bucket          string
	UploadDir       string `mapstructure:"upload_dir"`
	URLExpireTime   int    `mapstructure:"url_expire_time"`
}

var globalConfig *Config

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	config := &Config{}
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 加载 OSS 配置
	ossConfigPath := "configs/oss.yaml"
	ossViper := viper.New()
	ossViper.SetConfigFile(ossConfigPath)

	if err := ossViper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取 OSS 配置文件失败: %w", err)
	}

	if err := ossViper.Unmarshal(&config.OSS); err != nil {
		return nil, fmt.Errorf("解析 OSS 配置文件失败: %w", err)
	}

	globalConfig = config
	return config, nil
}

// LoadConfigWithEnv 根据环境加载多个配置文件
// 支持传入目录（会自动寻找目录下的配置文件）或者特定配置文件路径
// env 参数可选："dev", "test", "prod"，默认为 "dev"
func LoadConfigWithEnv(configPath string, env string) (*Config, error) {
	if env == "" {
		env = "dev" // 默认使用开发环境
	}

	v := viper.New()

	// 尝试多个可能的配置路径
	configPaths := []string{
		configPath,      // 原始传入路径
		"./configs",     // 相对于运行目录
		"../configs",    // 上一级目录
		"../../configs", // 上两级目录
	}

	configFound := false
	var configFile string

	// 尝试查找配置文件
	for _, path := range configPaths {
		if isDir(path) {
			baseConfigFile := fmt.Sprintf("%s/app.yaml", path)
			if fileExists(baseConfigFile) {
				v.AddConfigPath(path)
				v.SetConfigName("app")
				configFile = baseConfigFile
				configFound = true
				break
			}
		} else if fileExists(path) {
			v.SetConfigFile(path)
			configFile = path
			configFound = true
			break
		}
	}

	if !configFound {
		return nil, fmt.Errorf("无法找到配置文件，已尝试路径: %v", configPaths)
	}

	// 输出找到的配置文件路径，便于调试
	fmt.Printf("使用配置文件: %s\n", configFile)

	v.AutomaticEnv()
	v.SetEnvPrefix("")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 设置环境变量绑定
	v.BindEnv("app.port", "APP_PORT")
	v.BindEnv("database.host", "DB_HOST")
	v.BindEnv("database.port", "DB_PORT")
	v.BindEnv("database.username", "DB_USERNAME")
	v.BindEnv("database.password", "DB_PASSWORD")
	v.BindEnv("database.dbname", "DB_NAME")
	v.BindEnv("database.sslmode", "DB_SSLMODE")
	v.BindEnv("jwt.secret_key", "JWT_SECRET_KEY")

	// 读取基本配置
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取基本配置文件失败: %w", err)
	}

	// 读取环境特定配置
	if env != "" && v.ConfigFileUsed() != "" {
		configDir := filepath.Dir(v.ConfigFileUsed())
		envConfigFile := fmt.Sprintf("%s/app.%s.yaml", configDir, env)

		if fileExists(envConfigFile) {
			envViper := viper.New()
			envViper.SetConfigFile(envConfigFile)

			if err := envViper.ReadInConfig(); err != nil {
				return nil, fmt.Errorf("读取环境配置文件失败: %w", err)
			}

			// 合并环境配置
			if err := v.MergeConfigMap(envViper.AllSettings()); err != nil {
				return nil, fmt.Errorf("合并环境配置失败: %w", err)
			}

			fmt.Printf("已合并环境配置: %s\n", envConfigFile)
		} else {
			fmt.Printf("找不到环境配置文件: %s，将使用默认配置\n", envConfigFile)
		}
	}

	// 解析配置到结构体
	config := &Config{}
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 加载 OSS 配置
	configDir := filepath.Dir(v.ConfigFileUsed())
	ossConfigFile := fmt.Sprintf("%s/oss.yaml", configDir)

	if fileExists(ossConfigFile) {
		ossViper := viper.New()
		ossViper.SetConfigFile(ossConfigFile)
		
		// 为OSS配置添加自动环境变量支持
		ossViper.AutomaticEnv()
		ossViper.SetEnvPrefix("")
		ossViper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		
		// 绑定OSS相关的环境变量
		ossViper.BindEnv("aliyun_oss.access_key_id", "ALIYUN_ACCESS_KEY_ID")
		ossViper.BindEnv("aliyun_oss.access_key_secret", "ALIYUN_ACCESS_KEY_SECRET")
		ossViper.BindEnv("aliyun_oss.endpoint", "ALIYUN_OSS_ENDPOINT")
		ossViper.BindEnv("aliyun_oss.bucket", "ALIYUN_OSS_BUCKET")
		ossViper.BindEnv("aliyun_oss.region", "ALIYUN_OSS_REGION")
		ossViper.BindEnv("aliyun_oss.upload_dir", "ALIYUN_OSS_UPLOAD_DIR")
		
		// AWS S3
		ossViper.BindEnv("aws_s3.access_key_id", "AWS_ACCESS_KEY_ID")
		ossViper.BindEnv("aws_s3.secret_access_key", "AWS_SECRET_ACCESS_KEY")
		ossViper.BindEnv("aws_s3.region", "AWS_REGION")
		ossViper.BindEnv("aws_s3.bucket", "AWS_S3_BUCKET")
		ossViper.BindEnv("aws_s3.upload_dir", "AWS_S3_UPLOAD_DIR")
		
		// Cloudflare R2
		ossViper.BindEnv("cloudflare_r2.account_id", "CLOUDFLARE_ACCOUNT_ID")
		ossViper.BindEnv("cloudflare_r2.access_key_id", "CLOUDFLARE_R2_ACCESS_KEY_ID")
		ossViper.BindEnv("cloudflare_r2.secret_access_key", "CLOUDFLARE_R2_SECRET_ACCESS_KEY")
		ossViper.BindEnv("cloudflare_r2.bucket", "CLOUDFLARE_R2_BUCKET")
		ossViper.BindEnv("cloudflare_r2.upload_dir", "CLOUDFLARE_R2_UPLOAD_DIR")

		if err := ossViper.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("读取 OSS 配置文件失败: %w", err)
		}

		fmt.Printf("已加载OSS配置: %s\n", ossConfigFile)

		// 尝试加载环境特定的 OSS 配置
		if env != "" {
			ossEnvConfigFile := fmt.Sprintf("%s/oss.%s.yaml", configDir, env)
			if fileExists(ossEnvConfigFile) {
				ossEnvViper := viper.New()
				ossEnvViper.SetConfigFile(ossEnvConfigFile)

				if err := ossEnvViper.ReadInConfig(); err != nil {
					return nil, fmt.Errorf("读取环境 OSS 配置文件失败: %w", err)
				}

				// 合并 OSS 环境配置
				if err := ossViper.MergeConfigMap(ossEnvViper.AllSettings()); err != nil {
					return nil, fmt.Errorf("合并环境 OSS 配置失败: %w", err)
				}

				fmt.Printf("已合并OSS环境配置: %s\n", ossEnvConfigFile)
			}
		}

		if err := ossViper.Unmarshal(&config.OSS); err != nil {
			return nil, fmt.Errorf("解析 OSS 配置文件失败: %w", err)
		}
	} else {
		fmt.Printf("找不到OSS配置文件: %s，将使用默认OSS配置\n", ossConfigFile)
	}

	globalConfig = config
	return config, nil
}

// 检查是否是目录
func isDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// 检查文件是否存在
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// GetConfig 获取全局配置
func GetConfig() *Config {
	return globalConfig
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.Username, c.Password, c.DBName, c.SSLMode)
}

// GetConnMaxLifetime 获取数据库连接最大生命周期
func (c *DatabaseConfig) GetConnMaxLifetime() time.Duration {
	return time.Duration(c.ConnMaxLifetime) * time.Second
}

// GetJWTExpiration 获取 JWT 过期时间
func (c *JWTConfig) GetJWTExpiration() time.Duration {
	return time.Duration(c.ExpiresIn) * time.Second
}

// GetOSSURLExpiration 获取对象存储 URL 过期时间
func (c *AliyunOSSConfig) GetOSSURLExpiration() time.Duration {
	return time.Duration(c.URLExpireTime) * time.Second
}

func (c *AWSS3Config) GetOSSURLExpiration() time.Duration {
	return time.Duration(c.URLExpireTime) * time.Second
}

func (c *CloudflareR2Config) GetOSSURLExpiration() time.Duration {
	return time.Duration(c.URLExpireTime) * time.Second
}
