package main

import (
	"context"
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

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	cal := new(Call)
	rpc.Register(cal)

	// 不使用默认的rpc.HandleHTTP()，而是手动注册我们的处理器
	// 创建一个rpc处理器
	rpcServer := rpc.NewServer()
	rpcServer.Register(cal)

	// 注册调试路由和RPC处理器
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("收到请求: %s %s", r.Method, r.URL.Path)
		if r.URL.Path == "/rpc" {
			rpcServer.ServeHTTP(w, r)
			return
		}
		// 对于根路径，返回简单的欢迎信息
		if r.URL.Path == "/" {
			w.Write([]byte("RPC服务器正在运行"))
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
