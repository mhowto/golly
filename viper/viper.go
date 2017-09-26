package viper

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

func SetConfigFile(configName string) error {
	if configName == "" {
		return errors.New("empty config name while initing server")
	}
	//设置配置文件名称
	dir := filepath.Dir(configName)
	base := filepath.Base(configName)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	viper.SetConfigName(name)
	viper.AddConfigPath(dir)
	err := viper.ReadInConfig()
	if err != nil {
		return errors.New("[initConfig] read config failed")
	}

	//设置优先读取环境变量
	viper.AutomaticEnv()
	//设置key前缀,暂时不支持
	// viper.SetEnvPrefix("safe")
	//设置环境变量的替换规则
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)

	return nil
}

func ShouldHaveConfig(item string) string {
	s := viper.GetString(item)
	if s == "" {
		message := fmt.Sprintf("read config %s in the config file failed\n", item)
		panic(message)
	}
	return s
}
