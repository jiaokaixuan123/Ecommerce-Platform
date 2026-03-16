package config

import (
	"github.com/spf13/viper"  // Go 生态常用的配置管理库
)

// Config 总配置结构体：聚合所有子配置
type Config struct {
	App      AppConfig      `mapstructure:"app"`      // 应用基础配置
	MySQL    MySQLConfig    `mapstructure:"mysql"`    // MySQL 数据库配置
	Redis    RedisConfig    `mapstructure:"redis"`    // Redis 缓存配置
	JWT      JWTConfig      `mapstructure:"jwt"`      // JWT 鉴权配置
	Kafka    KafkaConfig    `mapstructure:"kafka"`    // Kafka 消息队列配置
}

// AppConfig 应用基础配置
type AppConfig struct {
	Name string `mapstructure:"name"` // 应用名称
	Mode string `mapstructure:"mode"` // 运行模式 debug | release
	Port int    `mapstructure:"port"` // 服务端口
}

// MySQLConfig MySQL 配置
type MySQLConfig struct {
	DSN         string `mapstructure:"dsn"`          // 连接字符串（如 user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local）
	MaxOpenConn int    `mapstructure:"max_open_conn"`// 最大打开连接数
	MaxIdleConn int    `mapstructure:"max_idle_conn"`// 最大空闲连接数
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`     // Redis 地址（host:port）
	Password string `mapstructure:"password"` // 密码
	DB       int    `mapstructure:"db"`       // 数据库编号
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret     string `mapstructure:"secret"`     // JWT 签名密钥
	ExpireHour int    `mapstructure:"expire_hour"`// token 过期小时数
}

// KafkaConfig Kafka 配置
type KafkaConfig struct {
	Brokers []string `mapstructure:"brokers"` // Kafka 集群地址列表（如 ["127.0.0.1:9092", "127.0.0.2:9092"]）
}

func Load(path string) (*Config, error) {
	// 1. 指定要加载的配置文件路径（如 "./config.yaml"）
	viper.SetConfigFile(path)
	// 2. 开启自动从环境变量覆盖配置（优先级：环境变量 > 配置文件）
	viper.AutomaticEnv()

	// 3. 读取配置文件内容
	if err := viper.ReadInConfig(); err != nil {
		return nil, err // 读取失败（文件不存在/权限问题等），返回错误
	}

	// 4. 将配置解析到 Config 结构体中
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err // 解析失败（字段类型不匹配等），返回错误
	}
	return &cfg, nil // 加载成功，返回配置对象
}
