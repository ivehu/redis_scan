package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"sort"
	"strings"

	"github.com/go-redis/redis/v8"
)

var (
	rdb        *redis.Client
	host       = flag.String("h", "127.0.0.1", "Redis服务器地址")
	port       = flag.String("p", "6379", "Redis服务器端口")
	password   = flag.String("a", "", "Redis认证密码")
	showKeys   = flag.Bool("show-keys", false, "显示各DB的key数量")
	showBigKey = flag.Bool("show-big-key", false, "显示各类型前10大key")
	showClient = flag.Bool("show-client", false, "显示客户端连接统计")
)

type KeySize struct {
	Key  string
	Size int64
}

func main() {
	flag.Parse()
	connectRedis()
	defer rdb.Close()

	switch {
	case *showKeys:
		showDatabaseKeys()
	case *showBigKey:
		findBigKeys()
	case *showClient:
		analyzeClients()
	default:
		flag.Usage()
	}
}

func connectRedis() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     net.JoinHostPort(*host, *port),
		Password: *password,
		DB:       0,
	})

	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		fmt.Println("Redis连接失败:", err)
		os.Exit(1)
	}
}

func showDatabaseKeys() {
	// 获取Keyspace信息
	info, err := rdb.Info(context.Background(), "Keyspace").Result()
	if err != nil {
		fmt.Println("获取Keyspace信息失败:", err)
		return
	}

	// 解析每个数据库的key数量
	dbKeys := make(map[string]int)
	total := 0

	for _, line := range strings.Split(info, "\n") {
		if strings.HasPrefix(line, "db") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				continue
			}

			// 解析keys数量
			dbName := parts[0]
			fields := strings.Split(parts[1], ",")
			for _, field := range fields {
				if strings.HasPrefix(field, "keys=") {
					keys, _ := fmt.Sscanf(field, "keys=%d", &total)
					if keys == 1 {
						dbKeys[dbName] = total
					}
					break
				}
			}
		}
	}

	// 排序数据库编号
	dbs := make([]string, 0, len(dbKeys))
	for db := range dbKeys {
		dbs = append(dbs, db)
	}
	sort.Strings(dbs)

	// 输出结果
	fmt.Println("\nDatabase Key Statistics:")
	for _, db := range dbs {
		fmt.Printf("%-5s : %d keys\n", db, dbKeys[db])
	}

	// 显示总数（当有多个db时）
	if len(dbs) > 1 {
		total := 0
		for _, count := range dbKeys {
			total += count
		}
		fmt.Printf("\nTotal keys: %d\n", total)
	}
}

var ctx = context.Background()

func findBigKeys() {
	type KeyStat struct {
		key   string
		count int64
	}

	// 修改为存储前5大key的slice
	biggest := map[string][]KeyStat{
		"string": make([]KeyStat, 0),
		"list":   make([]KeyStat, 0),
		"hash":   make([]KeyStat, 0),
		"set":    make([]KeyStat, 0),
		"zset":   make([]KeyStat, 0),
	}

	cursor := uint64(0)
	for {
		keys, newCursor, err := rdb.Scan(ctx, cursor, "*", 1000).Result()
		if err != nil {
			fmt.Println("Scan error:", err)
			break
		}

		for _, key := range keys {
			// Get key type and size based on type
			keyType, _ := rdb.Type(ctx, key).Result()
			var count int64

			switch keyType {
			case "string":
				count, _ = rdb.StrLen(ctx, key).Result()
			case "list":
				count, _ = rdb.LLen(ctx, key).Result()
			case "hash":
				count, _ = rdb.HLen(ctx, key).Result()
			case "set":
				count, _ = rdb.SCard(ctx, key).Result()
			case "zset":
				count, _ = rdb.ZCard(ctx, key).Result()
			default:
				continue
			}

			// 更新类型统计（保留前5）
			if stats, exists := biggest[keyType]; exists {
				// 添加新记录
				stats = append(stats, KeyStat{key: key, count: count})
				// 按元素数量降序排序
				sort.Slice(stats, func(i, j int) bool { return stats[i].count > stats[j].count })
				// 保留前5条
				if len(stats) > 10 {
					stats = stats[:10]
				}
				biggest[keyType] = stats
			}
		}

		if cursor = newCursor; cursor == 0 {
			break
		}
	}

	// 输出结果（修改为遍历前10）
	fmt.Println("\nTop 10 biggest keys per type:")
	unitMap := map[string]string{
		"string": "bytes",
		"list":   "items",
		"hash":   "fields",
		"set":    "members",
		"zset":   "members",
	}

	for t, stats := range biggest {
		if len(stats) == 0 {
			continue
		}
		unit := unitMap[t]
		fmt.Printf("\n[%s]\n", strings.ToUpper(t))
		for i, stat := range stats {
			fmt.Printf("%2d. %-60s (%d %s)\n",
				i+1,
				stat.key, // 直接使用原始key值
				stat.count,
				unit)
		}
	}
}

func analyzeClients() {
	clients, _ := rdb.ClientList(context.Background()).Result()
	ipCount := make(map[string]int)

	for _, client := range strings.Split(clients, "\n") {
		parts := strings.Split(client, " ")
		for _, part := range parts {
			if strings.HasPrefix(part, "addr=") {
				ip := strings.Split(part, "=")[1]
				ip = strings.Split(ip, ":")[0]
				ipCount[ip]++
			}
		}
	}

	fmt.Println("\nClient Connections:")
	for ip, count := range ipCount {
		fmt.Printf("%-15s : %d\n", ip, count)
	}
}
