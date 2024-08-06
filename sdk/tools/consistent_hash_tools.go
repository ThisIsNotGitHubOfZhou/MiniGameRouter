package tools

import (
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
)

// HashFunc 定义哈希函数类型
type HashFunc func(data []byte) uint32

// Map 定义一致性哈希结构
type HashMap struct {
	hash     HashFunc       // 哈希函数
	replicas int            // 虚拟节点倍数
	keys     []int          // 哈希环
	hashMap  map[int]string // 虚拟节点与真实节点的映射
	mu       sync.RWMutex   // 读写锁
}

// New 创建一个一致性哈希Map
func NewHashMap(replicas int, fn HashFunc) *HashMap {
	m := &HashMap{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add 添加真实节点
func (m *HashMap) Add(keys ...string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

// Remove 删除真实节点
func (m *HashMap) Remove(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := 0; i < m.replicas; i++ {
		hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
		index := sort.Search(len(m.keys), func(i int) bool { return m.keys[i] >= hash })
		if index < len(m.keys) && m.keys[index] == hash {
			m.keys = append(m.keys[:index], m.keys[index+1:]...)
			delete(m.hashMap, hash)
		}
	}
}

// Get 获取最近的节点
func (m *HashMap) Get(key string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if len(m.keys) == 0 {
		return ""
	}
	hash := int(m.hash([]byte(key)))
	index := sort.Search(len(m.keys), func(i int) bool { return m.keys[i] >= hash })
	if index == len(m.keys) {
		index = 0
	}
	return m.hashMap[m.keys[index]]
}

// Replicas 返回虚拟节点倍数
func (m *HashMap) Replicas() int {
	return m.replicas
}

// HashFunc 返回哈希函数
func (m *HashMap) HashFunc() HashFunc {
	return m.hash
}
