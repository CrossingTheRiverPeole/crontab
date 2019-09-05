package main

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

type result struct {
	err    error
	output []byte
}

func main() {

	var (
		cmd        *exec.Cmd
		ctx        context.Context
		cancelFunc context.CancelFunc
		resultChan chan *result
		res *result
	)

	ctx, cancelFunc = context.WithCancel(context.TODO())
	resultChan = make(chan *result, 1000)
	go func() {
		var (
			output []byte
			err    error
		)

		cmd = exec.CommandContext(ctx, "C:/cygwin64/bin/bash.exe", "-c", "sleep 2;echo 2")
		output, err = cmd.CombinedOutput()

		resultChan <- &result{
			err: err,
			output:output,
		}
	}()

	// 等待两秒，接收协程执行结果
	time.Sleep(1 * time.Second)
	// 取消上下文
	cancelFunc()
	//在主协程中等待子协程的退出，并打印执行结果
	res =<- resultChan

	fmt.Println(res.err,string(res.output))

}
