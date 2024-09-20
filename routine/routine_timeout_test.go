package routine

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

var list = []string{"aa", "bb", "cc", "dd", "ee", "ff", "gg", "hh", "ii", "jj", "kk", "ll", "mm", "nn"}

func TestTimeoutRoutine(t *testing.T) {
	var wg sync.WaitGroup

	limitChan := make(chan struct{}, 3)

	for _, item := range list {
		wg.Add(1)

		limitChan <- struct{}{}

		go func(s string) {
			defer wg.Done()
			consume(s, limitChan)
		}(item)
	}

	wg.Wait()
	fmt.Println("主协程结束......")
}

func consume(name string, limitChan chan struct{}) {

	defer func() {
		<-limitChan
	}()

	//模拟延迟
	time.Sleep(time.Duration(rand.Intn(5)) * time.Second)

	fmt.Println("name : " + name)
}
