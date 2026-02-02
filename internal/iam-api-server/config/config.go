package config

type AppConfig struct {
	Server *ServerConfig `mapstructure:"server"`
	JWT    *JWTConfig    `mapstructure:"jwt"`
	MySQL  *MySQLConfig  `mapstructure:"mysql"`
	Redis  *RedisConfig  `mapstructure:"redis"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Mode         string `mapstructure:"mode"` // debug/release/test
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

type JWTConfig struct {
	SecretKey              string `mapstructure:"secret_key"`
	AcessTokenExpireTime   int    `mapstructure:"access_token_expire_time"`
	RefreshTokenExpireTime int    `mapstructure:"refresh_token_expire_time"`
}

type MySQLConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	DatabaseName string `mapstructure:"database_name"`
}

type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	DatabaseName string `mapstructure:"database_name"`
}
