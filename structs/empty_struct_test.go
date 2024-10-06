package structs

import (
	"fmt"
	"testing"
	"time"
	"unsafe"
)

type EST struct {
}

func TestStructs(t *testing.T) {
	var b EST
	var c struct{}

	fmt.Printf("b address %p size %d\n", &b, unsafe.Sizeof(b))
	fmt.Printf("b address %p size %d\n", &c, unsafe.Sizeof(c))

	students := make(map[string]struct{}, 10)
	students["张三"] = EST{}
	students["李四"] = struct{}{}
	fmt.Println(len(students))

	teachers := make(chan struct{})
	go func() {
		time.Sleep(3 * time.Second)
		fmt.Println("子协程工作完毕")
		teachers <- struct{}{}
	}()
	<-teachers
}
