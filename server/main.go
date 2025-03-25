package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/rpc"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"time"
)

type Result struct {
	Num, Ans int
}

type Call int

func (cal *Call) Square(num int) *Result {
	return &Result{
		Num: num,
		Ans: num * num,
	}
}

func (cal *Call) RpcSquare(num int, result *Result) error {
	// 在处理请求开始时记录
	log.Printf("开始处理 RPC 请求: Square(%d)", num)

	result.Num = num
	result.Ans = num * num

	// 在处理完成后记录
	log.Printf("完成计算: %d^2 = %d", result.Num, result.Ans)
	return nil
}

// 自定义JSON-RPC请求结构
type JSONRPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      interface{}   `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
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

func (cal *Call) RpcRunCommand(command string, result *string) error {
	log.Printf("开始处理 RPC 请求: RunCommand(%s)", command)

	// 安全起见，将命令拆分为命令和参数
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return nil
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("命令执行失败: %v\n", err)
		return err
	}
	// 将输出赋值给 result
	*result = string(out)
	log.Printf("命令执行成功，输出:\n%s", *result)
	return nil
}

// 处理JSON-RPC请求
func handleJSONRPC(w http.ResponseWriter, r *http.Request, cal *Call) {
	// 只接受POST方法
	if r.Method != "POST" {
		log.Printf("错误的HTTP方法: %s", r.Method)
		w.Header().Set("Allow", "POST")
		http.Error(w, "只允许POST方法", http.StatusMethodNotAllowed)
		return
	}

	// 读取请求体
	body, err := ioutil.ReadAll(r.Body)
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
			err := cal.RpcRunCommand(cmd, &result)
			if err != nil {
				resp.Error = map[string]interface{}{
					"code":    -32000,
					"message": err.Error(),
				}
			} else {
				resp.Result = result
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

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	cal := new(Call)
	rpc.Register(cal)

	// 注册路由
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("收到请求: %s %s", r.Method, r.URL.Path)

		// 对于根路径，返回简单的欢迎信息
		if r.URL.Path == "/" {
			w.Write([]byte("RPC服务器正在运行"))
			return
		}

		// 处理 /rpc 端点
		if r.URL.Path == "/rpc" {
			handleJSONRPC(w, r, cal)
			return
		}

		http.NotFound(w, r)
	})

	handler := corsMiddleware(http.DefaultServeMux)

	// 创建自定义的 HTTP 服务器
	server := &http.Server{
		Addr:           ":1234",
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// 启动 goroutine 处理请求
	go func() {
		log.Printf("Serving RPC server on port --%d", 1234)
		if err := server.ListenAndServe(); err != nil {
			log.Fatal("Error serving: ", err)
		}
	}()

	// 等待中断信号以优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("Shutdown Server ...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}
