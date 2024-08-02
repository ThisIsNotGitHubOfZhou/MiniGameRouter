package main

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/rand"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// 随机生成指定长度的字符串
func randomString(length int) (string, error) {
	// 使用当前时间的纳秒数作为种子
	seed := time.Now().UnixNano()
	rand.Seed(seed)

	result := make([]byte, length)
	for i := range result {
		// 生成一个随机索引
		index := rand.Intn(len(charset))
		// 从字符集中选择字符
		result[i] = charset[index]
	}

	return string(result), nil
}

func hashStringToRange(s string, max int) int {

	// 计算字符串的SHA-256哈希值
	hash := sha256.Sum256([]byte(s))

	// 将哈希值的前8个字节转换为一个无符号整数
	hashInt := binary.BigEndian.Uint64(hash[:8])

	// 将无符号整数映射到指定的区间

	return int(hashInt % uint64(max))
}

func main() {
	maxRange := 8
	res := []int{0, 0, 0, 0, 0, 0, 0, 0}
	for i := 0; i < 100000; i++ {
		s, _ := randomString(5)
		hash := hashStringToRange(s, maxRange)
		fmt.Printf("str: %v,hash: %v\n", s, hash)
		if hash != hashStringToRange(s, maxRange) {
			fmt.Println("hash error")
		}
		res[hash]++
	}
	fmt.Println("最后分布：", res)
}
