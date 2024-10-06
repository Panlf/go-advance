package coroutine

import (
	"fmt"
	"testing"
	"time"
)

type person struct {
	name string
	age  int
}

func TestRoutineVariable(t *testing.T) {
	run := func() {
		fmt.Println(">>>>>>>>>")
	}

	go run()

	time.Sleep(1 * time.Second)
	fmt.Println("main over......")
}

func TestRoutineAnonymous(t *testing.T) {
	go func() {
		fmt.Println(">>>>>>>>>")
	}()

	time.Sleep(1 * time.Second)
	fmt.Println("main over......")
}

func TestRoutineFunction(t *testing.T) {
	ch := make(chan string)

	human := &person{}

	go createPeople(human)

	go printlnChan(ch)

	fmt.Println("person is", human)

	a := <-ch

	fmt.Println("ch values is", a)

	time.Sleep(1 * time.Second)
	fmt.Println("main over......")
}

func createPeople(human *person) {
	human.name = "Anan"
	human.age = 18
	fmt.Println("create person is over")
}

func printlnChan(ch chan string) {
	defer func() {
		close(ch)
	}()

	ch <- "aaa"

	fmt.Println(">>>>>>>>>")
}
