package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/rpc"

	"github.com/spf13/viper"
)

// 配置结构体
/* type Config struct {
	Server struct {
		Port int    `mapstructure:"port"`
		Host string `mapstructure:"host"`
	} `mapstructure:"server"`
	RPC struct {
		Endpoint string `mapstructure:"endpoint"`
	} `mapstructure:"rpc"`
} */

type Call struct{}

// 自定义JSON-RPC请求结构
type JSONRPCRequest struct {
	JSONRPC  string        `json:"jsonrpc"`
	ID       interface{}   `json:"id"`
	Method   string        `json:"method"`
	Params   []interface{} `json:"params"`
	Authcode string        `json:"authcode"`
}

// 自定义JSON-RPC响应结构
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// 添加CORS中间件
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 设置CORS头部
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// 对于OPTIONS请求直接返回
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// 处理请求
		next.ServeHTTP(w, r)
	})
}

// 处理JSON-RPC请求
func handleJSONRPC(w http.ResponseWriter, r *http.Request) {
	// 只接受POST方法
	if r.Method != "POST" {
		log.Printf("错误的HTTP方法: %s", r.Method)
		w.Header().Set("Allow", "POST")
		http.Error(w, "只允许POST方法", http.StatusMethodNotAllowed)
		return
	}

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("读取请求体失败: %v", err)
		http.Error(w, "读取请求失败", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// 解析JSON-RPC请求
	var req JSONRPCRequest
	if err := json.Unmarshal(body, &req); err != nil {
		log.Printf("解析JSON-RPC请求失败: %v", err)
		http.Error(w, "无效的JSON-RPC请求", http.StatusBadRequest)
		return
	}

	// 检查授权码
	if string(req.Authcode) != viper.GetString("Authcode") {
		log.Printf("无效的授权码: %s", req.Authcode)
		http.Error(w, "无效的授权码", http.StatusUnauthorized)
		return
	}

	log.Printf("收到RPC请求: %s, 参数: %v", req.Method, req.Params)

	// 准备响应
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
	}

	// 处理RPC方法调用
	if req.Method == "Call.RpcRunCommand" && len(req.Params) >= 1 {
		cmd, ok := req.Params[0].(string)
		if !ok {
			resp.Error = map[string]interface{}{
				"code":    -32602,
				"message": "无效的参数",
			}
		} else {
			var result string

			// 从配置中获取RPC端点
			rpcEndpoint := Endpoint
			if rpcEndpoint == "" {
				rpcEndpoint = "localhost:1234" // 默认值
			}

			client, err := rpc.DialHTTP("tcp", rpcEndpoint)
			if err != nil {
				log.Printf("RPC连接失败: %v", err)
				resp.Error = map[string]interface{}{
					"code":    -32603,
					"message": "RPC服务连接失败",
					"data":    err.Error(),
				}
			} else {
				defer client.Close()

				// 调用远程RPC方法
				err := client.Call("Call.RpcRunCommand", cmd, &result)
				if err != nil {
					log.Printf("RPC调用失败: %v", err)
					resp.Error = map[string]interface{}{
						"code":    -32000,
						"message": "RPC方法调用失败",
						"data":    err.Error(),
					}
				} else {
					log.Printf("RPC调用成功，结果: %s", result)
					resp.Result = result
				}
			}
		}
	} else {
		resp.Error = map[string]interface{}{
			"code":    -32601,
			"message": "方法不存在",
		}
	}

	// 发送响应
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(resp); err != nil {
		log.Printf("编码响应失败: %v", err)
		http.Error(w, "服务器内部错误", http.StatusInternalServerError)
	}
}

func loadConfig() {
	// 设置配置文件名称和类型
	viper.SetConfigName("config") // 配置文件名称(不带扩展名)
	viper.SetConfigType("toml")   // 配置文件类型

	// 添加配置文件可能的路径
	viper.AddConfigPath(".")         // 当前目录
	viper.AddConfigPath("./config")  // config子目录
	viper.AddConfigPath("../config") // 上一级的config目录

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Printf("警告: 配置文件未找到: %v", err)
		} else {
			log.Fatalf("配置文件读取失败: %v", err)
		}
	} else {
		log.Printf("使用配置文件: %s", viper.ConfigFileUsed())
	}
}

func handleCode(w http.ResponseWriter, r *http.Request) {
	// 只接受POST方法
	if r.Method != "POST" {
		log.Printf("错误的HTTP方法: %s", r.Method)
		w.Header().Set("Allow", "POST")
		http.Error(w, "只允许POST方法", http.StatusMethodNotAllowed)
		return
	}

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("读取请求体失败: %v", err)
		http.Error(w, "读取请求失败", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// 读取shortcode字段
	var req map[string]interface{}
	if err := json.Unmarshal(body, &req); err != nil {
		log.Printf("解析JSON-RPC请求失败: %v", err)
		http.Error(w, "无效的JSON-RPC请求", http.StatusBadRequest)
		return
	}

	// 验证shortcode合法性
	code := req["shortcode"].(string)
	// log.Print(code)
	command := viper.GetString("Code." + code)
	if command == "" {
		log.Printf("无效的shortcode: %s", code)
		http.Error(w, "无效的shortcode", http.StatusBadRequest)
		return
	}

	// 发送RPC请求
	client, err := rpc.DialHTTP("tcp", Endpoint)
	if err != nil {
		log.Printf("RPC连接失败: %v", err)
		http.Error(w, "RPC服务连接失败", http.StatusInternalServerError)
		return
	}

	var result string
	err = client.Call("Call.RpcRunCommand", command, &result)
	if err != nil {
		log.Printf("RPC调用失败: %v", err)
		http.Error(w, "RPC方法调用失败", http.StatusInternalServerError)
		return
	}

	// 返回响应
	responsebody := map[string]interface{}{
		"result": result,
	}

	log.Print(responsebody)

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(responsebody); err != nil {
		log.Printf("编码响应失败: %v", err)
		http.Error(w, "服务器内部错误", http.StatusInternalServerError)
	}
}

// 启动HTTP服务器
func StartServer(port int) error {

	// 创建HTTP处理器
	mux := http.NewServeMux()
	mux.HandleFunc("/rpc", func(w http.ResponseWriter, r *http.Request) {
		handleJSONRPC(w, r)
	})

	mux.HandleFunc("/code", func(w http.ResponseWriter, r *http.Request) {
		handleCode(w, r)
		// log.Print(w)
	})

	// 应用CORS中间件
	handler := corsMiddleware(mux)

	// 启动服务器
	addr := fmt.Sprintf(":%d", port)
	log.Printf("HTTP RPC服务器启动在 %s", addr)
	return http.ListenAndServe(addr, handler)
}
