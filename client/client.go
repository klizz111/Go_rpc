package main

import (
	"fmt"
	"log"
	"net/rpc"

	"github.com/spf13/viper"
)

type Result struct {
	Num, Ans int
}

func main() {
	// 设置日志输出格式
	log.SetFlags(log.Ldate | log.Ltime)

	client, err := rpc.DialHTTP("tcp", "localhost:1234")
	if err != nil {
		log.Fatalf("连接RPC服务器失败: %v", err)
	}

	var result string

	// 读取配置文件
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	err = viper.ReadInConfig()
	if err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}
	command := viper.GetString("Command.ls")

	// 发起异步调用
	asyncCall := client.Go("Call.RpcRunCommand", command, &result, nil)

	// 等待调用完成
	<-asyncCall.Done

	if asyncCall.Error != nil {
		log.Printf("RPC调用出错: %v", asyncCall.Error)
		return
	}

	// 使用 fmt.Printf 直接打印到标准输出
	fmt.Printf("\n命令执行结果:\n%s\n", result)
}
