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
// NOTE:分片个数配置化~
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

// 计算幂函数
func intPow(base, exponent int) int {
	result := 1
	for exponent > 0 {
		if exponent%2 == 1 {
			result *= base
		}
		base *= base
		exponent /= 2
	}
	return result
}

func InitMysql() {
	query := `
        SELECT COUNT(*)
        FROM information_schema.tables
        WHERE table_schema = ? AND table_name = ?
    `
	var count int
	dbMaxID := intPow(2, config.NameSplitSize)*intPow(2, config.PrefixSplitSize) - 1
	dbMaxIDStr := strconv.Itoa(dbMaxID)
	err := config.MysqlClient.QueryRow(query, "route_db", "route_info"+dbMaxIDStr).Scan(&count)
	if err != nil {
		config.Logger.Fatalf("[Error][discover][mysql] Failed to query RouteInfo table: %v\n", err)
		return
	}
	if count > 0 {
		config.Logger.Println("[Info][discover][mysql] RouteInfo table already exists")
	} else {
		config.Logger.Printf("[Info][discover][mysql] RouteInfo table route_info%v not exists\n", dbMaxIDStr)
		// 使用create like  route_info来创建到dbMaxID号表

		for dbID := 0; dbID <= dbMaxID; dbID++ {
			daIDStr := strconv.Itoa(dbID)
			createTableQuery := fmt.Sprintf("CREATE TABLE IF NOT EXISTS route_info%s LIKE route_info", daIDStr)
			_, err := config.MysqlClient.Exec(createTableQuery)
			if err != nil {
				config.Logger.Fatalf("[Error][discover][mysql] Failed to create RouteInfo table: %v\n", err)
				return
			}
			config.Logger.Println("[Info][discover][mysql] RouteInfo table created successfully")
		}

	}
}
func WriteToMysql(info *pb.RouteInfo) error {
	// 组合shardingKey:info.Name生成2位，info.Prefix生成三位，组合在一起生成五位

	// 根据shardingkey选择分片
	dbID := hashStringToRange(info.Name, intPow(2, config.NameSplitSize))<<config.PrefixSplitSize + hashStringToRange(info.Prefix, intPow(2, config.PrefixSplitSize))

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

func UpdateToMysql(name, prefix string, info *pb.RouteInfo) error {

	// 根据shardingkey选择分片
	dbID := hashStringToRange(name, intPow(2, config.NameSplitSize))<<config.PrefixSplitSize + hashStringToRange(prefix, intPow(2, config.PrefixSplitSize))
	config.Logger.Printf("[Info][discover][mysql] UpdateToMysql Name: %v, Prefix: %v, DB_ID: %v\n",
		name, prefix, strconv.FormatInt(int64(dbID), 2))
	// 写入mysql
	err := updateToMysql(dbID, name, prefix, info)
	if err != nil {
		config.Logger.Printf("[Error][discover][mysql] Failed to update to MySQL: %v\n", err)
		return err
	}
	return nil
}

func updateToMysql(dbID int, name, prefix string, info *pb.RouteInfo) error {
	tableName := fmt.Sprintf("route_info%d", dbID)
	query := fmt.Sprintf("UPDATE %s SET name = ?, host = ?, port = ?, prefix = ?, metadata = ? WHERE name = ? AND prefix = ?", tableName)
	//config.Logger.Printf("[Info][discover][mysql] 执行写入命令：%v\n", query)
	if config.MysqlClient == nil {
		config.Logger.Println("[Error][discover][mysql] db is nil")
		return fmt.Errorf("db is nil")
	}
	_, err := config.MysqlClient.Exec(query, info.Name, info.Host, info.Port, info.Prefix, info.Metadata, name, prefix)
	if err != nil {
		config.Logger.Println("[Error][discover][mysql] 更新数据错误:", err)
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
	// TODO：需要先在redis里读吗,注意别造成循环引用！因为现在读redis可能会调用这个函数~
	if !config.IsK8s {
		config.ReadRouteTotalTimes.Inc()
	}
	// 根据sharding key选择分片
	dbID := hashStringToRange(name, intPow(2, config.NameSplitSize))<<config.PrefixSplitSize + 0
	endDBID := (hashStringToRange(name, intPow(2, config.NameSplitSize))+1)<<config.PrefixSplitSize + 0

	var res []*pb.RouteInfo

	// 只提供name的话，需要遍历部分db
	for i := dbID; i < endDBID; i++ {
		//startTime := time.Now()              // 记录开始时间
		//elapsedTime := time.Since(startTime) // 计算耗时
		//config.Logger.Printf("~~~~~~~~~~~~~~~~~~~~~~ReadFromMysqlWithName耗时: %v\n", elapsedTime)
		tempRes, err := readFromDBWithName(i, name)
		if err != nil {
			config.Logger.Println("[Error][discover][mysql] readFromDBWithName Error: ", err)
			return nil, err
		}
		res = append(res, tempRes...)
	}

	// 同步到内存
	// TODO:这里耗时较大重点优化~
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

// NOTE:这是改进后，5w量级能有10%优化~
//func readFromDBWithName(dbID int, name string) ([]*pb.RouteInfo, error) {
//	tableName := fmt.Sprintf("route_info%d", dbID)
//	query := fmt.Sprintf("SELECT name, host, port, prefix, metadata FROM %s WHERE name = ?", tableName)
//
//	// 使用预编译语句
//	stmt, err := config.MysqlClient.Prepare(query)
//	if err != nil {
//		config.Logger.Println("[Error][discover][mysql] 预编译查询语句错误:", err)
//		return nil, fmt.Errorf("failed to prepare query: %v", err)
//	}
//	defer stmt.Close()
//
//	rows, err := stmt.Query(name)
//	if err != nil {
//		config.Logger.Println("[Error][discover][mysql] 查询数据错误:", err)
//		return nil, fmt.Errorf("failed to execute query: %v", err)
//	}
//	defer rows.Close()
//
//	var routeInfos []*pb.RouteInfo
//	for rows.Next() {
//		var routeInfo pb.RouteInfo
//		if err := rows.Scan(&routeInfo.Name, &routeInfo.Host, &routeInfo.Port, &routeInfo.Prefix, &routeInfo.Metadata); err != nil {
//			config.Logger.Println("[Error][discover][mysql] 扫描数据错误:", err)
//			return nil, fmt.Errorf("failed to scan row: %v", err)
//		}
//		routeInfos = append(routeInfos, &routeInfo)
//	}
//
//	if err := rows.Err(); err != nil {
//		config.Logger.Println("[Error][discover][mysql] 读取数据错误:", err)
//		return nil, fmt.Errorf("rows error: %v", err)
//	}
//
//	return routeInfos, nil
//}

func ReadFromMysqlWithPrefix(name, prefix string) ([]*pb.RouteInfo, error) {
	config.Logger.Printf("[Info][discover][mysql][ReadFromMysqlWithPrefix] name: %v, prfix: %v\n", name, prefix)
	// TODO：需要先在redis里读吗,注意别造成循环引用！因为现在读redis可能会调用这个函数~
	if !config.IsK8s {
		config.ReadRouteTotalTimes.Inc()
	}
	// 根据sharding key选择分片
	dbID := hashStringToRange(name, intPow(2, config.NameSplitSize))<<config.PrefixSplitSize + hashStringToRange(prefix, intPow(2, config.PrefixSplitSize))

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
