package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

var ctx = context.Background()

func FlushAll() {
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

func main() {
	// 清空redis
	//FlushAll()

	// 统计总行数
	total, err := GetTotalRowCount()
	if err != nil {
		log.Fatalf("Failed to get total row count: %v", err)
	}
	fmt.Printf("Total row count: %d\n", total)
	// 清空mysql
	FlushMysql()
}
