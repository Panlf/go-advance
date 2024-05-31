package sync

import (
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/loveleshsharma/gohive"
)

var wg sync.WaitGroup

// 地址管道
var addressChan = make(chan string, 100)

type TcpWorker struct{}

func (tcp *TcpWorker) Run() {

	//函数结束 释放连接
	defer wg.Done()

	for {

		address, ok := <-addressChan

		if !ok {
			break
		}

		conn, err := net.Dial("tcp", address)

		if err != nil {
			continue
		}

		conn.Close()
		fmt.Println("open:", address)
	}

}

func TestDate(t *testing.T) {
	var begin = time.Now()

	var ip = "192.168.0.109"

	var pool_size = 70000
	var pool = gohive.NewFixedPool(pool_size)

	//启动一个线程，用于生成ip:port，并且存放到地址管道中
	go func() {
		for port := 1; port <= 65535; port++ {
			var address = fmt.Sprintf("%s:%d", ip, port)
			addressChan <- address
		}

		close(addressChan)
	}()

	//启动pool_size个工人，处理addressChan中的每个地址
	for work := 0; work < pool_size; work++ {
		wg.Add(1)
		pool.Submit(&TcpWorker{})
	}

	wg.Wait()
	var elapseTime = time.Since(begin)

	//pool.Close()
	fmt.Println("耗时：", elapseTime)
}
