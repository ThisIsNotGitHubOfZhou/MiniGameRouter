package database

import (
	"crypto/sha256"
	"database/sql"
	"discoversvr/config"
	pb "discoversvr/proto"
	"encoding/binary"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"strconv"
	"sync"
	"time"
)

// 数据库连接池
var dbPools map[int]*sql.DB
var dbPoolsMutex sync.RWMutex

// 初始化数据库连接池
func initDBPools() {
	dbPools = make(map[int]*sql.DB)
	for i := 0; i < 32; i++ { // 假设有 32 个分片
		dsn := fmt.Sprintf("root:664597599Zcf!@tcp(9.134.206.110:3306)/route_db")
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			log.Fatalf("Failed to connect to database %d: %v", i, err)
		}
		db.SetMaxOpenConns(100)          // 设置最大打开连接数
		db.SetMaxIdleConns(10)           // 设置最大空闲连接数
		db.SetConnMaxLifetime(time.Hour) // 设置连接的最大生命周期
		dbPools[i] = db
	}
}

func WriteToMysql(info *pb.RouteInfo) error {
	// 组合shardingKey:info.Name生成2位，info.Prefix生成三位，组合在一起生成五位

	// 根据shardingkey选择分片
	dbID := hashStringToRange(info.Name, 4)<<3 + hashStringToRange(info.Prefix, 8)

	config.Logger.Printf("[Info][discover][mysql] WriteToMysql Name: %v, Prefix: %v, DB_ID: %v\n",
		info.Name, info.Prefix, strconv.FormatInt(int64(dbID), 2))

	// 写入mysql
	err := writeToDB(dbID, info)
	if err != nil {
		config.Logger.Printf("[Error][discover][mysql] Failed to write to MySQL: %v\n", err)
		return err
	}
	return nil
}

// writeToDB writes RouteInfo to the specified database shard
func writeToDB(dbID int, info *pb.RouteInfo) error {
	db, ok := dbPools[dbID]
	if !ok {
		return fmt.Errorf("no database found for dbID %d", dbID)
	}

	tableName := fmt.Sprintf("route_info%d", dbID)
	query := fmt.Sprintf("INSERT INTO %s (name, host, port, prefix, metadata) VALUES (?, ?, ?, ?, ?)", tableName)
	config.Logger.Printf("[Info][discover][mysql] 执行写入命令：%v\n", query)
	_, err := db.Exec(query, info.Name, info.Host, info.Port, info.Prefix, info.Metadata)
	if err != nil {
		config.Logger.Println("[Error][discover][mysql] 插入数据错误:", err)
		return fmt.Errorf("failed to execute query: %v", err)
	}

	return nil
}

// hashStringToRange hashes a string using SHA-256 and maps it to a specified range [0, maxRange-1]
func hashStringToRange(s string, maxRange int) int {
	// Step 1: Compute the SHA-256 hash value of the string
	hasher := sha256.New()
	hasher.Write([]byte(s))
	hashValue := hasher.Sum(nil)

	// Step 2: Convert the first 4 bytes of the hash value to a uint32
	hashInt := binary.BigEndian.Uint32(hashValue[:4])

	// Step 3: Map the hash value to the specified range
	return int(hashInt % uint32(maxRange))
}

func ReadFromMysql(info *pb.RouteInfo) error {
	// 组合shardingKey:info.Name+info.Host+info.Port生成2位，info.Prefix生成三位，组合在一起生成五位

	// 根据shardingkey选择分片
	dbID := hashStringToRange(info.Name+info.Host+info.Port, 4)<<3 + hashStringToRange(info.Prefix, 8)

	config.Logger.Printf("[Info][discover][mysql] WriteToMysql Name: %v, Host: %v, Port: %v, Prefix: %v, DB_ID: %v",
		info.Name, info.Host, info.Port, info.Prefix, strconv.FormatInt(int64(dbID), 2))

	// 直接从mysql里面读？

	return nil
}
