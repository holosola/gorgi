package config

type AppConf struct {
	Database *Database `mapstructure:"database"`
	Cache    *Cache    `mapstructure:"cache"`
	Log      Log       `mapstructure:"log"`
}

type Database struct {
	Driver   string `mapstructure:"driver"`
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	DB       string `mapstructure:"db"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type Cache struct {
	Driver   string `mapstructure:"driver"`
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	DB       string `mapstructure:"db"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type Log struct {
	Prefix   string `mapstructure:"prefix,omitempty"`
	FileName string `mapstructure:"filename,omitempty"`
	Level    string `mapstructure:"level"`
	MaxSize  int    `mapstructure:"maxsize,omitempty"`
	MaxAge   int    `mapstructure:"maxage,omitempty"`
	Compress bool   `mapstructure:"compress,omitempty"`
}
