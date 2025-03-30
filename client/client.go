package main

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

var (
	Port     int
	Host     string
	Endpoint string
)

func main() {
	// 加载配置
	loadConfig()

	// 初始化全局变量
	Port = viper.GetInt("Server.port")
	if Port == 0 {
		Port = 8008
	}
	Host = viper.GetString("Server.host")
	Endpoint = viper.GetString("Rpc.endpoint")

	fmt.Print("Host: ", Host)

	// 打印配置信息
	log.Printf("服务器配置 - 端口: %d, 主机: %s", Port, Host)
	log.Printf("RPC配置 - 端点: %s", Endpoint)

	// 启动HTTP服务器
	err := StartServer(Port)
	if err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
