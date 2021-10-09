package main

import (
	"fmt"
	"math/rand"
	"time"
)

var randomPool = "abcdefghijklmnopqrstuvwxyz123456789"

func getRandomID (length int) (str string) {
    rand.Seed(time.Now().UnixNano())
    b := make([]byte, length)
    rand.Read(b)
    return fmt.Sprintf("%x", b)[:length]
}