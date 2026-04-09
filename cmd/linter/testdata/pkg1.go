package main

import (
	"log"
	"os"
)

func main() {
	// Здесь не будет ругаться
	log.Fatal("this is allowed")
	os.Exit(1)
}

func Foo() {
	// Здесь будет ругаться
	// формулируем ожидания: анализатор должен находить ошибку, описанную в комментарии want
	panic("should be detected")     // want "panic usage detected"
	log.Fatal("should be detected") // want "log.Fatal usage detected"
	os.Exit(1)                      // want "os.Exit usage detected"
}

func Bar() {
	log.Println("safe")
}
