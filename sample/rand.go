package main

import (
	"fmt"
	"math/rand"
	"time"
)

// ランダムな文字列を生成する関数
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func main() {
	// 1秒ごとにランダムな文字列を出力する
	ticker := time.NewTicker(1000 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		fmt.Println(randomString(50)) // 10文字のランダムな文字列を生成して出力
	}
}
