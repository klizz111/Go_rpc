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
	"syscall"
	"time"
)

type Result string
type Call struct{}

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
	rpc.Register(new(Call))
	rpc.HandleHTTP()

	server := &http.Server{Addr: ":1234"}

	// 使用 goroutine 启动服务
	go func() {
		log.Printf("Serving RPC server on port %d", 1234)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error serving: %v", err)
		}
	}()

	// 设置优雅退出
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop // 等待中断信号

	log.Println("Shutting down server...")

	// 创建一个上下文用于控制关闭超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
