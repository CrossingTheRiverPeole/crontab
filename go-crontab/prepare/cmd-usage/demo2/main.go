package main

import (
	"fmt"
	"os/exec"
)

func main()  {
	var  (
		cmd *exec.Cmd
		output []byte
		err error
	)
	cmd = exec.Command("C:/cygwin64/bin/bash.exe", "-c", "echo 2")

	cmd.Output()
	if output, err = cmd.CombinedOutput(); err != nil{
		cmd.Output()
		fmt.Println(err)
		return
	}
	//打印子进程的输出
	fmt.Println(string(output))
}
