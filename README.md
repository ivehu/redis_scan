# Redis 扫描工具

专业的Redis服务器诊断工具，支持多种扫描模式。

## 功能特性
✅ 数据库Key统计  
✅ Top5大Key扫描  
✅ 客户端连接分析  
✅ 跨平台支持（Windows/Linux/macOS）

## 安装使用

### 快速安装
```bash
# 克隆仓库
git clone https://github.com/ivehu/redis_scan.git
cd redis_scan

# 构建项目
go get github.com/go-redis/redis/v8
go mod tidy
make build

# 安装项目
make install    

# 查看帮助
./redis_scan -h

# 显示各DB key数量
./redis_scan -show-keys

# 扫描前10大Key（按类型）
./redis_scan -h 192.168.1.100 -p 6379 -a password -show-big-key

# 客户端连接统计
./redis_scan -show-client
