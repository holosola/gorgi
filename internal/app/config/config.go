package config

import (
	"fmt"

	"github.com/holosola/gorgi/internal/pkg/ex"
	"github.com/spf13/viper"
)

var appConf *viper.Viper

func init() {
	fmt.Println("初始化配置~")
	appConf = viper.New()
	appConf.AddConfigPath("./internal/app/config/") // 文件所在目录
	appConf.SetConfigName("web")                    // 文件名
	appConf.SetConfigType("yaml")                   // 文件类型
	err := appConf.ReadInConfig()
	// 读取配置信息失败
	if err != nil {
		ex.Handle(err)
	}
	fmt.Println("初始化配置结束~")

	// 监控配置文件变化
	appConf.WatchConfig()
}

func GetConfig() *viper.Viper {
	return appConf
}
