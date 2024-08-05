package database

import (
	"crypto/sha256"
	"discoversvr/config"
	pb "discoversvr/proto"
	"encoding/binary"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
	"sync"
)

var dbPoolsMutex sync.RWMutex

// 初始化数据库连接池
// TODO:优化这个部分~高并发不会出现冲突+性能
// TODO:优化快照读写~
// TODO:分片个数配置化~
//func init() {
//	dbPools = make(map[int]*sql.DB)
//	for i := 0; i < 32; i++ { // 假设有 32 个分片
//		dsn := fmt.Sprintf("root:664597599Zcf!@tcp(9.134.206.110:3306)/route_db")
//		db, err := sql.Open("mysql", dsn)
//		if err != nil {
//			log.Fatalf("Failed to connect to database %d: %v", i, err)
//		}
//		db.SetMaxOpenConns(100)          // 设置最大打开连接数
//		db.SetMaxIdleConns(10)           // 设置最大空闲连接数
//		db.SetConnMaxLifetime(time.Hour) // 设置连接的最大生命周期
//		dbPools[i] = db
//	}
//}

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
	//db, ok := dbPools[dbID]
	//if !ok {
	//	return fmt.Errorf("no database found for dbID %d", dbID)
	//}

	tableName := fmt.Sprintf("route_info%d", dbID)
	query := fmt.Sprintf("INSERT INTO %s (name, host, port, prefix, metadata) VALUES (?, ?, ?, ?, ?)", tableName)
	//config.Logger.Printf("[Info][discover][mysql] 执行写入命令：%v\n", query)
	if config.MysqlClient == nil {
		config.Logger.Println("[Error][discover][mysql] db is nil")
	}
	_, err := config.MysqlClient.Exec(query, info.Name, info.Host, info.Port, info.Prefix, info.Metadata)
	if err != nil {
		config.Logger.Println("[Error][discover][mysql] 插入数据错误:", err)
		return fmt.Errorf("failed to execute query: %v", err)
	}

	return nil
}

// hashStringToRange hashes a string using SHA-256 and maps it to a specified range [0, maxRange-1]
func hashStringToRange(s string, max int) int {

	// 计算字符串的SHA-256哈希值
	hash := sha256.Sum256([]byte(s))

	// 将哈希值的前8个字节转换为一个无符号整数
	hashInt := binary.BigEndian.Uint64(hash[:8])

	// 将无符号整数映射到指定的区间

	return int(hashInt % uint64(max))
}

func ReadFromMysqlWithName(name string) ([]*pb.RouteInfo, error) {
	config.Logger.Printf("[Info][discover][mysql][ReadFromMysqlWithName]  Name: %v\n", name)
	// 根据shardingkey选择分片
	dbID := hashStringToRange(name, 4)<<3 + 0
	endDBID := (hashStringToRange(name, 4)+1)<<3 + 0

	var res []*pb.RouteInfo
	// TODO:这里耗时较大重点优化~
	// 只提供name的话，需要遍历部分db
	for i := dbID; i < endDBID; i++ {
		//config.Logger.Printf("[Info][discover][mysql] ReadFromMysqlWithName Name: %v, DB_ID: %v",
		//	name, strconv.FormatInt(int64(i), 2))
		tempRes, err := readFromDBWithName(i, name)
		if err != nil {
			config.Logger.Println("[Error][discover][mysql] readFromDBWithName Error: ", err)
			return nil, err
		}
		res = append(res, tempRes...)
	}

	// 同步到内存
	WriteSyncRoutes(config.SyncRedisClient, res)
	return res, nil

}

func readFromDBWithName(dbID int, name string) ([]*pb.RouteInfo, error) {

	tableName := fmt.Sprintf("route_info%d", dbID)
	query := fmt.Sprintf("SELECT name, host, port, prefix, metadata FROM %s WHERE name = ?", tableName)
	rows, err := config.MysqlClient.Query(query, name)
	defer rows.Close()
	if err != nil {
		config.Logger.Println("[Error][discover][mysql] 查询数据错误:", err)
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}

	var routeInfos []*pb.RouteInfo
	for rows.Next() {
		var routeInfo pb.RouteInfo
		if err := rows.Scan(&routeInfo.Name, &routeInfo.Host, &routeInfo.Port, &routeInfo.Prefix, &routeInfo.Metadata); err != nil {
			config.Logger.Println("[Error][discover][mysql] 扫描数据错误:", err)
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		routeInfos = append(routeInfos, &routeInfo)
	}

	if err := rows.Err(); err != nil {
		config.Logger.Println("[Error][discover][mysql] 读取数据错误:", err)
		return nil, fmt.Errorf("rows error: %v", err)
	}

	return routeInfos, nil
}

func ReadFromMysqlWithPrefix(name, prefix string) ([]*pb.RouteInfo, error) {
	config.Logger.Printf("[Info][discover][mysql][ReadFromMysqlWithPrefix] name: %v, prfix: %v\n", name, prefix)
	// 根据shardingkey选择分片
	dbID := hashStringToRange(name, 4)<<3 + hashStringToRange(prefix, 8)

	//config.Logger.Printf("[Info][discover][mysql] ReadFromMysqlWithPrefix Name: %v, Prefix: %v, DB_ID: %v",
	//	name, prefix, strconv.FormatInt(int64(dbID), 2))

	res, err := readFromDBWithPrefix(dbID, name, prefix)

	// 同步到内存
	WriteSyncRoutes(config.SyncRedisClient, res)
	return res, err

}

func readFromDBWithPrefix(dbID int, name, prefix string) ([]*pb.RouteInfo, error) {

	db := config.MysqlClient

	tableName := fmt.Sprintf("route_info%d", dbID)
	query := fmt.Sprintf("SELECT name, host, port, prefix, metadata FROM %s WHERE name = ? AND prefix = ?", tableName)
	rows, err := db.Query(query, name, prefix)
	defer rows.Close()
	if err != nil {
		config.Logger.Println("[Error][discover][mysql] 查询数据错误:", err)
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}

	var routeInfos []*pb.RouteInfo
	for rows.Next() {
		var routeInfo pb.RouteInfo
		if err := rows.Scan(&routeInfo.Name, &routeInfo.Host, &routeInfo.Port, &routeInfo.Prefix, &routeInfo.Metadata); err != nil {
			config.Logger.Println("[Error][discover][mysql] 扫描数据错误:", err)
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		routeInfos = append(routeInfos, &routeInfo)
	}

	if err := rows.Err(); err != nil {
		config.Logger.Println("[Error][discover][mysql] 读取数据错误:", err)
		return nil, fmt.Errorf("rows error: %v", err)
	}

	return routeInfos, nil
}
