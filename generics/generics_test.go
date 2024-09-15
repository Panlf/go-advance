package generics

import (
	"fmt"
	"testing"
)

func printSomething[T any](s []T) {
	for _, v := range s {
		fmt.Printf("值：%v \n", v)
	}
}

func TestPrintSomething(t *testing.T) {
	printSomething[string]([]string{"aa", "bb", "cc"})
}
