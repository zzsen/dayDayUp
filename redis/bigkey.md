# Redis 大 Key 问题与优化

## 1. 什么是大 Key

大 Key 不是指 key 本身的字符串长度大，而是指 key 对应的 **value 占用的内存过大** 或 **包含的元素数量过多**。

### 1.1 大 Key 的判定标准

没有绝对的标准，通常以下情况可认定为大 Key：

| 数据类型 | 大 Key 阈值（参考值） |
|----------|----------------------|
| String | value 大小 > **10KB**（有些场景以 1MB 为界） |
| Hash | field 数量 > **5000** 或 value 总大小 > **10MB** |
| List | 元素数量 > **5000** |
| Set | 成员数量 > **5000** |
| ZSet | 成员数量 > **5000** |

> 具体阈值应结合业务场景和 Redis 实例配置来定。阿里云等云服务商通常以 String > 10KB、其他类型元素 > 5000 个作为告警阈值。

### 1.2 大 Key 的危害

| 危害 | 说明 |
|------|------|
| **阻塞主线程** | Redis 命令执行是单线程的，操作大 Key（读/写/删除）会长时间占用线程，阻塞其他请求 |
| **内存不均** | 在 Redis Cluster 中，大 Key 导致某个节点内存远大于其他节点，造成数据倾斜 |
| **网络带宽压力** | 读取大 Key 时传输大量数据，可能打满网卡带宽，影响同实例的其他请求 |
| **主从同步延迟** | 大 Key 的写入会生成大的 RDB/AOF 数据，增加主从同步延迟 |
| **删除导致阻塞** | 直接 `DEL` 大 Key，Redis 需要逐个释放元素的内存，可能阻塞数秒甚至更久 |
| **持久化影响** | fork 子进程做 RDB 快照或 AOF 重写时，大 Key 增加 Copy-On-Write 的内存开销 |
| **慢查询** | 对大 Key 的操作容易触发慢查询（slowlog），影响整体 QPS |

## 2. 如何发现大 Key

### 2.1 redis-cli --bigkeys

Redis 自带的大 Key 扫描工具，会遍历所有 key，找出每种数据类型中最大的 key。

> `--bigkeys` 是 `redis-cli` 的**启动参数**（非 Redis 命令），只能在终端中使用，RDM / RESP.app 的控制台**无法使用**。RESP.app 新版自带 Memory Analysis 功能，可作为替代。

```bash
redis-cli --bigkeys

# 建议在从节点执行，避免影响主节点
redis-cli -h <slave-host> -p 6379 --bigkeys

# 配合 -i 参数控制扫描速度（每次 SCAN 间隔 0.1 秒）
redis-cli --bigkeys -i 0.1
```

输出示例：

```
[00.00%] Biggest string found so far '"user:profile:10086"' with 52428 bytes
[00.00%] Biggest hash   found so far '"order:detail:20230101"' with 12680 fields

-------- summary -------
Biggest string found '"user:profile:10086"' has 52428 bytes
Biggest   hash found '"order:detail:20230101"' has 12680 fields
```

**局限性**：只能找到每种类型最大的那一个 key，无法列出所有大 Key。

### 2.2 MEMORY USAGE 命令（Redis 4.0+）

精确查看某个 key 占用的内存字节数。`MEMORY USAGE` 是标准 Redis 命令，可以在 redis-cli 交互模式、RDM / RESP.app 控制台、代码中调用等任何能执行 Redis 命令的地方使用。

```bash
MEMORY USAGE <key> [SAMPLES <count>]
```

```bash
> MEMORY USAGE user:profile:10086
(integer) 52480

> MEMORY USAGE order:detail:20230101 SAMPLES 0
(integer) 1048576
```

`SAMPLES` 参数指定对集合类型采样的元素个数，`SAMPLES 0` 表示精确计算所有元素（耗时更长）。

### 2.3 SCAN + 脚本批量扫描

通过 `SCAN` 遍历所有 key，逐个检测大小，可以找出所有大 Key。

```bash
#!/bin/bash
# 扫描所有 key，输出大于 10KB 的 key
redis-cli --no-auth-warning SCAN 0 COUNT 100 | while read key; do
    size=$(redis-cli MEMORY USAGE "$key" 2>/dev/null)
    if [ "$size" -gt 10240 ] 2>/dev/null; then
        echo "$key -> $size bytes"
    fi
done
```

实际生产环境建议使用 Python/Go 脚本，配合 `SCAN` + `TYPE` + `MEMORY USAGE` 做精细化扫描，并控制扫描频率。

### 2.4 RDB 离线分析工具

通过分析 RDB 文件找出大 Key，**不影响线上服务**。

| 工具 | 说明 |
|------|------|
| [redis-rdb-tools](https://github.com/sripathikrishnan/redis-rdb-tools) | Python 编写，支持导出为 JSON/CSV、按大小过滤 |
| [rdb](https://github.com/HDT3213/rdb) | Go 编写，性能更好，支持大 Key 分析 |

```bash
# redis-rdb-tools 示例：找出大于 10KB 的 key
rdb -c memory dump.rdb --bytes 10240 -f bigkeys.csv
```

### 2.5 监控告警

在生产环境中，应建立大 Key 的常态化监控：

- Redis 慢查询日志（`SLOWLOG`）：大 Key 操作通常会出现在慢查询中
- 云服务商控制台：阿里云、AWS ElastiCache 等通常自带大 Key 分析功能
- Prometheus + Redis Exporter：自定义指标监控

## 3. 大 Key 的产生原因

| 原因 | 典型场景 |
|------|----------|
| **数据未设过期时间** | 不断向 List/Set/Hash 追加数据，但从不清理 |
| **业务设计不合理** | 将大量数据聚合到一个 key 中（如一个 Hash 存所有用户信息） |
| **数据增长未预估** | 初期数据量小无感知，随业务增长 key 逐渐膨胀 |
| **缓存全量数据** | 将数据库查询结果整个序列化后存入一个 String key |
| **消息积压** | 使用 List 做消息队列，消费者故障导致数据堆积 |

## 4. 如何避免大 Key

### 4.1 拆分大 Key

**核心思路**：将一个大 Key 拆分为多个小 Key，分散存储。

#### Hash 拆分

将一个大 Hash 按照一定规则拆分为多个小 Hash。

```
# 拆分前：一个 Hash 存 100 万用户
HSET user:all  uid:1 "{...}"
HSET user:all  uid:2 "{...}"
...
HSET user:all  uid:1000000 "{...}"

# 拆分后：按 uid 取模分桶（1000 个桶）
HSET user:bucket:{uid % 1000}  uid:1 "{...}"
HSET user:bucket:0  uid:1000 "{...}"
HSET user:bucket:1  uid:1 "{...}"
...
```

读取时先计算桶号，再从对应的 Hash 中获取：

```go
bucket := uid % 1000
field := fmt.Sprintf("uid:%d", uid)
result := redis.HGet(ctx, fmt.Sprintf("user:bucket:%d", bucket), field)
```

#### String 拆分

大 String（如序列化的 JSON）拆分为多个字段存入 Hash，或按业务维度拆分为多个独立 key。

```
# 拆分前
SET user:10086 "{name: 'xxx', avatar: '<大量base64>', settings: '{...}'}"

# 拆分后：将大字段独立存储
SET    user:10086:base    "{name: 'xxx', age: 25}"
SET    user:10086:avatar  "<base64数据>"
HSET   user:10086:settings  theme "dark"
HSET   user:10086:settings  lang  "zh-CN"
```

#### List / Set / ZSet 拆分

按时间、ID 范围等维度分片。

```
# 拆分前：一个 List 存百万条消息
LPUSH msg:queue "{msg1}" "{msg2}" ...

# 拆分后：按日期分片
LPUSH msg:queue:20250303 "{msg1}"
LPUSH msg:queue:20250304 "{msg2}"
```

### 4.2 数据压缩

对于 String 类型的大 value，在写入前进行压缩。

```go
import "github.com/klauspost/compress/snappy"

// 写入时压缩
data, _ := json.Marshal(obj)
compressed := snappy.Encode(nil, data)
redis.Set(ctx, key, compressed, ttl)

// 读取时解压
compressed, _ := redis.Get(ctx, key).Bytes()
data, _ := snappy.Decode(nil, compressed)
json.Unmarshal(data, &obj)
```

常用压缩算法对比：

| 算法 | 压缩率 | 速度 | 适用场景 |
|------|--------|------|----------|
| Snappy | 中等 | 极快 | 追求速度，可接受中等压缩率 |
| LZ4 | 中等 | 极快 | 与 Snappy 类似，压缩率略高 |
| Gzip | 高 | 慢 | 追求高压缩率，对速度不敏感 |
| Zstd | 高 | 快 | 兼顾压缩率和速度，推荐 |

### 4.3 设置合理的过期时间

确保所有缓存 key 都有过期时间，避免数据无限增长。

```bash
SET key value EX 3600         # 设置 1 小时过期
EXPIRE key 86400              # 设置 24 小时过期
```

对于集合类型，如果不能整体过期，需要在业务层定期清理过期元素：

```go
// 定期清理 ZSet 中过期的数据（如保留最近 7 天的数据）
cutoff := time.Now().AddDate(0, 0, -7).Unix()
redis.ZRemRangeByScore(ctx, key, "-inf", strconv.FormatInt(cutoff, 10))
```

### 4.4 合理选择数据结构

| 场景 | 推荐方案 | 说明 |
|------|----------|------|
| 存储对象的多个字段 | Hash（字段数适中时） | 避免将整个 JSON 序列化为 String |
| 大量 bool 标记 | Bitmap | 如用户签到，百万用户仅占 ~125KB |
| 去重计数 | HyperLogLog | 百万级别去重计数仅占 12KB（有约 0.81% 误差） |
| 存储大文本/文件 | 外部存储（OSS/S3） | Redis 中只存 URL 或元数据 |

### 4.5 业务层限制

在写入 Redis 前做大小检查：

```go
const MaxValueSize = 10 * 1024 // 10KB

func safeSet(ctx context.Context, key string, value []byte, ttl time.Duration) error {
    if len(value) > MaxValueSize {
        return fmt.Errorf("value size %d exceeds limit %d", len(value), MaxValueSize)
    }
    return rdb.Set(ctx, key, value, ttl).Err()
}
```

对于集合类型，限制元素数量：

```go
const MaxListLen = 5000

func safeLPush(ctx context.Context, key string, values ...interface{}) error {
    length, _ := rdb.LLen(ctx, key).Result()
    if length+int64(len(values)) > MaxListLen {
        return fmt.Errorf("list length would exceed limit %d", MaxListLen)
    }
    return rdb.LPush(ctx, key, values...).Err()
}
```

## 5. 已有大 Key 的优化方案

### 5.1 安全删除大 Key

直接 `DEL` 大 Key 会阻塞 Redis 主线程。推荐以下方式：

#### 方式 1：UNLINK（Redis 4.0+，推荐）

```bash
UNLINK <key>
```

`UNLINK` 将 key 从 keyspace 中移除（O(1)），然后在**后台线程**中异步释放内存，不阻塞主线程。

#### 方式 2：渐进式删除（Redis 4.0 之前）

对于低版本 Redis，需要分批删除元素，再删除 key：

```bash
# Hash：使用 HSCAN + HDEL 分批删除
HSCAN bigkey 0 COUNT 100
HDEL bigkey field1 field2 field3 ...
# 重复直到所有 field 删除完毕
DEL bigkey

# List：使用 LTRIM 逐步截断
LTRIM bigkey 0 -101    # 每次从尾部删除 100 个
# 重复直到 LLEN 为 0
DEL bigkey

# Set：使用 SSCAN + SREM 分批删除
SSCAN bigkey 0 COUNT 100
SREM bigkey member1 member2 member3 ...
DEL bigkey

# ZSet：使用 ZREMRANGEBYRANK 分批删除
ZREMRANGEBYRANK bigkey 0 99   # 每次删除 100 个
# 重复直到 ZCARD 为 0
DEL bigkey
```

Go 代码示例（渐进式删除 Hash）：

```go
func deleteHashProgressively(ctx context.Context, rdb *redis.Client, key string) error {
    for {
        var cursor uint64
        var fields []string
        var err error
        
        fields, cursor, err = rdb.HScan(ctx, key, cursor, "*", 100).Result()
        if err != nil {
            return err
        }
        
        if len(fields) > 0 {
            // fields 包含 [field1, value1, field2, value2, ...]
            fieldNames := make([]string, 0, len(fields)/2)
            for i := 0; i < len(fields); i += 2 {
                fieldNames = append(fieldNames, fields[i])
            }
            rdb.HDel(ctx, key, fieldNames...)
        }
        
        if cursor == 0 {
            break
        }
        
        time.Sleep(10 * time.Millisecond)
    }
    return rdb.Del(ctx, key).Err()
}
```

#### 方式 3：开启 lazyfree（Redis 4.0+）

通过配置让 Redis 的某些删除操作自动异步化：

```bash
# redis.conf
lazyfree-lazy-expire yes       # key 过期时异步释放
lazyfree-lazy-server-del yes   # RENAME 等内部删除时异步释放
lazyfree-lazy-user-del yes     # DEL 命令改为异步（等同于 UNLINK），Redis 6.0+
replica-lazy-flush yes         # 从节点全量同步前清空数据时异步执行
```

开启 `lazyfree-lazy-user-del` 后，`DEL` 行为等同于 `UNLINK`。

### 5.2 大 Key 的拆分迁移

对于已存在的大 Key，可以编写脚本进行在线拆分：

```go
func migrateHashToShards(ctx context.Context, rdb *redis.Client, srcKey string, shardCount int64) error {
    var cursor uint64
    for {
        fields, nextCursor, err := rdb.HScan(ctx, srcKey, cursor, "*", 100).Result()
        if err != nil {
            return err
        }
        
        pipe := rdb.Pipeline()
        for i := 0; i < len(fields); i += 2 {
            field, value := fields[i], fields[i+1]
            // 按 field 哈希值分桶
            shard := crc32.ChecksumIEEE([]byte(field)) % uint32(shardCount)
            dstKey := fmt.Sprintf("%s:shard:%d", srcKey, shard)
            pipe.HSet(ctx, dstKey, field, value)
        }
        pipe.Exec(ctx)
        
        cursor = nextCursor
        if cursor == 0 {
            break
        }
        time.Sleep(10 * time.Millisecond)
    }
    
    // 验证完成后删除源 key
    return rdb.Unlink(ctx, srcKey).Err()
}
```

### 5.3 读取优化

如果暂时无法拆分，可以优化读取方式：

```bash
# 差：一次性获取整个 Hash
HGETALL bigkey          # O(N)，会阻塞

# 好：只获取需要的字段
HGET bigkey field1      # O(1)
HMGET bigkey f1 f2 f3   # O(N)，N 为请求的字段数

# 差：一次性获取整个 List
LRANGE bigkey 0 -1      # O(N)

# 好：分页获取
LRANGE bigkey 0 99      # 每次获取 100 个
LRANGE bigkey 100 199
```

## 6. 大 Key 问题 Checklist

```
预防阶段：
 1. [ ] 所有缓存 key 设置合理的 TTL
 2. [ ] 集合类型控制元素数量上限
 3. [ ] 大 value 写入前压缩（Snappy/Zstd）
 4. [ ] 业务层做 value 大小检查
 5. [ ] 合理选择数据结构（Bitmap、HyperLogLog 等）
 6. [ ] 大文本/文件存外部存储，Redis 仅存引用

发现阶段：
 7. [ ] 定期执行 redis-cli --bigkeys 扫描
 8. [ ] 慢查询日志监控
 9. [ ] 接入大 Key 告警（云服务/自建监控）

治理阶段：
10. [ ] 删除大 Key 使用 UNLINK 或渐进式删除
11. [ ] 开启 lazyfree 配置
12. [ ] 制定大 Key 拆分迁移方案
13. [ ] 优化读取方式（避免 HGETALL、LRANGE 0 -1 等全量操作）
```

## 参考资料

- [阿里云 Redis 大 Key 最佳实践](https://help.aliyun.com/document_detail/353223.html)
- [Redis 官方文档 - UNLINK](https://redis.io/commands/unlink/)
- [Redis 官方文档 - MEMORY USAGE](https://redis.io/commands/memory-usage/)
- [Redis 官方文档 - Lazy Freeing](https://redis.io/docs/management/config/#lazy-freeing)
