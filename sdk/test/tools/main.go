package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"math/rand"
	"time"
)

var ctx = context.Background()

func FlushAll0() {
	client := redis.NewClient(&redis.Options{
		Addr:     "21.6.163.18:6380", // Redis 地址
		Password: "664597599Zcf!",    // Redis 密码，没有则留空
		DB:       0,                  // 使用的数据库，默认为0
	})

	// 使用 SCAN 命令逐步扫描键
	var cursor uint64
	for {
		var keys []string
		var err error
		keys, cursor, err = client.Scan(ctx, cursor, "*", 100).Result()
		if err != nil {
			log.Fatalf("Failed to scan keys: %v", err)
		}

		if len(keys) > 0 {
			// 使用 DEL 命令删除扫描到的键
			if err := client.Del(ctx, keys...).Err(); err != nil {
				log.Fatalf("Failed to delete keys: %v", err)
			}
		}

		// 如果 cursor 为 0，表示扫描结束
		if cursor == 0 {
			break
		}
	}

	log.Println("All keys in the current database have been deleted.")
}

func FlushAll1() {
	client := redis.NewClient(&redis.Options{
		Addr:     "21.6.163.18:6380", // Redis 地址
		Password: "664597599Zcf!",    // Redis 密码，没有则留空
		DB:       1,                  // 使用的数据库，默认为0
	})

	// 使用 SCAN 命令逐步扫描键
	var cursor uint64
	for {
		var keys []string
		var err error
		keys, cursor, err = client.Scan(ctx, cursor, "*", 100).Result()
		if err != nil {
			log.Fatalf("Failed to scan keys: %v", err)
		}

		if len(keys) > 0 {
			// 使用 DEL 命令删除扫描到的键
			if err := client.Del(ctx, keys...).Err(); err != nil {
				log.Fatalf("Failed to delete keys: %v", err)
			}
		}

		// 如果 cursor 为 0，表示扫描结束
		if cursor == 0 {
			break
		}
	}

	log.Println("All keys in the current database have been deleted.")
}

func FlushMysql() {
	// 数据库连接字符串
	dsn := fmt.Sprintf("root:664597599Zcf!@tcp(9.134.206.110:3306)/route_db")

	// 创建数据库连接
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// 测试数据库连接
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 清空表的操作
	for i := 0; i <= 31; i++ {
		tableName := fmt.Sprintf("route_info%d", i)
		query := fmt.Sprintf("TRUNCATE TABLE %s", tableName)

		_, err := db.Exec(query)
		if err != nil {
			log.Fatalf("Failed to truncate table %s: %v", tableName, err)
		}

		fmt.Printf("Successfully truncated table %s\n", tableName)
	}
}

func GetTotalRowCount() (int, error) {
	// 数据库连接字符串
	dsn := fmt.Sprintf("root:664597599Zcf!@tcp(9.134.206.110:3306)/route_db")

	// 创建数据库连接
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	total := 0
	for i := 0; i <= 31; i++ {
		tableName := fmt.Sprintf("route_info%d", i)
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
		var count int
		err := db.QueryRow(query).Scan(&count)
		if err != nil {
			return 0, err
		}
		fmt.Printf("Table %s has %d rows\n", tableName, count)
		total += count
	}
	return total, nil
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// 随机生成指定长度的字符串
func randomString(length int, rng *rand.Rand) string { //加锁是防止生成相同字符
	result := make([]byte, length)

	for i := range result {
		// 生成一个随机索引
		index := rng.Intn(len(charset))
		// 从字符集中选择字符
		result[i] = charset[index]
	}
	return string(result)
}

// 随机写total条数据到MySQL
func WriteRandomRouteToMysql(total int) error {
	// 数据库连接字符串
	dsn := fmt.Sprintf("root:664597599Zcf!@tcp(9.134.206.110:3306)/route_db")

	// 创建数据库连接
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for j := 0; j < total; j += 32 {
		for i := 0; i <= 31; i++ {
			tableName := fmt.Sprintf("route_info%d", i)

			query := fmt.Sprintf("INSERT INTO %s (name, host, port, prefix, metadata) VALUES (?, ?, ?, ?, ?)", tableName)
			//config.Logger.Printf("[Info][discover][mysql] 执行写入命令：%v\n", query)
			if db == nil {
				fmt.Println("db is nil")
				return fmt.Errorf("db is nil")
			}
			_, err = db.Exec(query, randomString(10, rng), "0.0.0.0", "666", randomString(10, rng), "{}")
			if err != nil {
				fmt.Println("插入数据错误:", err)
				return fmt.Errorf("failed to execute query: %v", err)
			}
			fmt.Println("插入数据:", j+i)
		}
	}

	return nil
}

func main() {
	// 清空redis
	FlushAll0()
	FlushAll1()
	// 统计总行数
	total, err := GetTotalRowCount()
	if err != nil {
		log.Fatalf("Failed to get total row count: %v", err)
	}
	fmt.Printf("Total row count: %d\n", total)

	// 清空mysql
	//FlushMysql()

	// 写1千万数据到mysql
	//err = WriteRandomRouteToMysql(10000000)
	//if err != nil {
	//	fmt.Printf("Failed to write random route to mysql: %v", err)
	//}

}
