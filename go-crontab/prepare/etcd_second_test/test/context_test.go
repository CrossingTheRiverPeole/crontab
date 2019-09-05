package test

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"
)

func doStuff(ctx context.Context) {
	for {
		time.Sleep(time.Second)
		select {
		case <-ctx.Done():
			log.Println("done")
			return
		default:
			fmt.Println("work")
		}
	}
}

func TestContextWithCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go doStuff(ctx)
	time.Sleep(time.Second * 10)
	cancel()
}


func TestContextWithDeadline(t *testing.T)  {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second * 5))
	go doStuff(ctx)
	time.Sleep(time.Second * 10)
	cancel()
	fmt.Println("down")
}
