package idGenerator

import (
	"math/rand"
	"time"
)

func GetRandomID() string {
	rand.Seed(time.Now().UTC().UnixNano())
	buf := make([]byte, 5)
	for i, _ := range buf {
		buf[i] = byte(97 + rand.Intn(123-97))
	}

	return string(buf[:])
}
