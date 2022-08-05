package gag

import (
	"fmt"
	"time"
)

type logger struct{}

func (l logger) Println(str string) {
	s := fmt.Sprintf("{\"time\":\"%v\",\"message\":\"%s\"}", time.Now(), str)
	fmt.Println(s)
}
