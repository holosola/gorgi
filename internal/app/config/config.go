package config

import (
	"github.com/holosola/gorgi/internal/pkg/ex"
	"github.com/spf13/viper"
)

var appConf *viper.Viper

func init() {
	appConf = viper.New()
	appConf.AddConfigPath("./configs/") // 文件所在目录
	appConf.SetConfigName("web")        // 文件名
	appConf.SetConfigType("yaml")       // 文件类型
	err := appConf.ReadInConfig()
	// 读取配置信息失败
	if err != nil {
		ex.Handle(err)
	}
	// 监控配置文件变化
	appConf.WatchConfig()
}

func GetConfig() *viper.Viper {
	return appConf
}
