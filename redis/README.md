## redis
Redis是一个开源的使用**C语言**编写、支持网络、可**基于内存**亦可持久化的日志型、**Key-Value**的NoSQL数据库。

但是底层存储不是使用C语言的字符串类型，而是自己开发了一种数据类型SDS进行存储，SDS即**Simple Dynamic String** ，是一种动态字符串。

### 特点
#### 读写速度快
redis官网测试读写能到10万左右每秒。
速度快的原因:
1. 数据存储在内存中, 访问内存的速度是远远大于访问磁盘
2. Redis采用单线程的架构, 避免了上下文的切换和多线程带来的竞争, 也就不存在加锁释放锁的操作, 减少了CPU的消耗
3. 采用了非阻塞IO多路复用机制。

### 数据结构丰富
Redis不仅仅支持简单的key-value类型的数据, 同时还提供list, set, zset, hash等数据结构。

### 支持持久化
Redis提供了RDB和AOF两种持久化策略, 能最大限度地保证Redis服务器宕机重启后数据不会丢失。

### 支持高可用
可以使用主从复制, 并且提供哨兵机制, 保证服务器的高可用。

### 客户端语言多
因为Redis受到社区和各大公司的广泛认可, 所以客户端语言涵盖了所有的主流编程语言, 比如Java, C, C++, PHP, NodeJS等等。

## Redis的五种数据类型底层实现原理
Redis的五种数据类型底层实现原理章节摘抄自: [Redis的五种数据类型底层实现原理是什么](https://zhuanlan.zhihu.com/p/344918922)


Redis是一个**Key-Value**型的内存数据库, 它所有的key都是字符串, 而value常见的数据类型有五种：**string, list, set, zset, hash**。

Redis的这些数据结构, 在底层都是使用**redisObject**来进行表示。**redisObject**中有三个重要的属性, 分别是**type**、 **encoding** 和 **ptr**。
  >**type** : 保存的value的类型 
  >
  >**encoding** : 保存的value的编码
  >
  >**ptr** : 一个指针，指向实际保存的value的数据结构

| type | encoding | 对象 |
| --- | --- | --- |
| REDIS_STRING | REDIS_ENCODING_INT | 使用整形值实现的字符串对象 ｜
| REDIS_STRING | REDIS_ENCODING_EMBSTR | 使用embstr编码的简单动态字符串实现的字符串对象 |
| REDIS_STRING | REDIS_ENCODING_RAW | 使用简单动态字符串实现的字符串对象 |
| REDIS_LIST | REDIS_ENCODING_ZIPLIST | 使用压缩列表实现的列表对象 |
| REDIS_LIST | REDIS_ENCODING_LINKEDLIST | 使用双端链表的列表对象 | 
| REDIS_HASH | REDIS_ENCODING_ZIPLIST | 使用压缩列表实现的哈希对象 |
| REDIS_HASH | REDIS_ENCODING_HT | 使用字典实现的哈希对象 |
| REDIS_SET | REDIS_ENCODING_INTSET | 使用证书集合实现的集合对象 |
| REDIS_SET | REDIS_ENCODING_HT | 使用字典实现的集合对象 |
| REDIS_ZSET | REDIS_ENCODING_ZIPLIST | 使用压缩列表实现的有序集合对象 |
| REDIS_ZSET | REDIS_ENCODING_SKIPLIST | 使用跳跃表和字典实现的有序集合对象 |

### string(字符串)
字符串对象的 encoding 有三种，分别是：int、raw、embstr。
``` c
struct sdshdr{
 int len;/*字符串长度*/
 int free;/*未使用的字节长度*/
 char buf[];/*保存字符串的字节数组*/
}
```
SDS与C语言的字符串有什么区别呢？
* **遍历复杂度低**, C语言获取字符串长度是从头到尾遍历，时间复杂度是O(n)，而**SDS有len属性记录字符串长度，时间复杂度为O(1)**。
* **避免缓冲区溢出**, SDS在需要修改时，会先检查空间是否满足大小，如果不满足，则先扩展至所需大小再进行修改操作。
* **空间预分配**, 当SDS需要进行扩展时，Redis会为SDS分配好内存，并且根据特定的算法分配多余的free空间，避免了连续执行字符串添加带来的内存分配的消耗。
* **惰性释放**, 如果需要缩短字符串，不会立即回收多余的内存空间，而是用free记录剩余的空间，以备下次扩展时使用，避免了再次分配内存的消耗。
* **二进制安全**, c语言在存储字符串时采用N+1的字符串数组，末尾使用'\0'标识字符串的结束，如果我们存储的字符串中间出现'\0'，那就会导致识别出错。而SDS因为记录了字符串的长度len，则没有这个问题。
字符串类型的应用是非常广泛的，比如可以把对象转成JSON字符串存储到Redis中作为缓存，也可以使用decr、incr命令用于计数器的实现，又或者是用setnx命令为基础实现分布式锁等等。
需要注意的是：**Redis 规定了字符串的长度不得超过 512 MB**。

### hash(字典)
哈希对象的 encoding 有两种，分别是：ziplist、hashtable。
当哈希对象保存的键值对数量小于 512，并且所有键值对的长度都小于 64 字节时，使用ziplist(压缩列表)存储；否则使用 hashtable 存储。

Redis中的hashtable跟Java中的HashMap类似，都是通过"数组+链表"的实现方式解决部分的哈希冲突。
源码定义:
``` c
typedf struct dict{
    dictType *type;//类型特定函数，包括一些自定义函数，这些函数使得key和value能够存储
    void *private;//私有数据
    dictht ht[2];//两张hash表 
    int rehashidx;//rehash索引，字典没有进行rehash时，此值为-1
    unsigned long iterators; //正在迭代的迭代器数量
}dict;

typedef struct dictht{
     //哈希表数组
     dictEntry **table;
     //哈希表大小
     unsigned long size;
     //哈希表大小掩码，用于计算索引值
     //总是等于 size-1
     unsigned long sizemask;
     //该哈希表已有节点的数量
     unsigned long used; 
}dictht;

typedf struct dictEntry{
    void *key;//键
    union{
        void val;
        unit64_t u64;
        int64_t s64;
        double d;
    }v;//值
    struct dictEntry *next；//指向下一个节点的指针
}dictEntry;
```
结构图大致如下:
```
dict                                                            
├─ *type
├─ ht[0] ─────
├─ ht[1] ─────└─ dictht
├─ *privdata       ├─ **table ─── 1 ──────────────── 2 ───── 3 ───── ...                 
├─ rehashid        ├─ size        │                  │                                                         
└─ iterators       ├─ sizemask    dictEntry          dictEntry                                                 
                   └─ used         ├─ *val            ├─ *val                                                 
                                   ├─ *key            ├─ *key                                                   
                                   └─ *next           └─ *next                                             
                                        │                  │                                                                
                                        dictEntry          dictEntry                                                        
                                          ├─ *val            ├─ *val                                                        
                                          ├─ *key            ├─ *key                                                          
                                          └─ *next           └─ *next                                                        
                                               │                  │                                                       
                                               null               null                                                   
                                                                        
```
#### 渐进式rehash
优缺点：

优点是把rehash操作分散到每一个字典操作和定时函数上，避免了一次性集中式rehash带来的服务器压力。

缺点是在rehash期间需要使用两个hash表，占用内存稍大。

hash类型的常用命令有：hget、hset、hgetall 等。


### list(链表)
列表的 encoding 有两种，分别是：ziplist、linkedlist。
当列表的长度小于 512，并且所有元素的长度都小于 64 字节时，使用ziplist存储；否则使用 linkedlist 存储。

Redis中的linkedlist类似于Java中的LinkedList，是一个链表，底层的实现原理也和LinkedList类似。这意味着list的插入和删除操作效率会比较快，时间复杂度是O(1)。
源码定义:
```c
typedef struct listNode {
    struct listNode *prev;
    struct listNode *next;
    void *value;
} listNode;

typedef struct listIter {
    listNode *next;
    int direction;
} listIter;

typedef struct list {
    listNode *head;
    listNode *tail;
    void *(*dup)(void *ptr);
    void (*free)(void *ptr);
    int (*match)(void *ptr, void *key);
    unsigned long len;
} list;
```
list类型常用的命令有：lpush、rpush、lpop、rpop、lrange等。


### set(集合)
set类型的特点很简单，无序，不重复，跟Java的HashSet类似。
它的编码有两种，分别是intset和hashtable。
如果value可以转成整数值，并且长度不超过512的话就使用intset存储，否则采用hashtable。

hashtable在前面讲hash类型时已经讲过，这里的set集合采用的hashtable几乎一样，只是哈希表的value都是NULL。这个不难理解，比如用Java中的HashMap实现一个HashSet，我们只用HashMap的key就是了。

``` c
我们讲一讲intset，先看源码。

typedef struct intset{
    uint32_t encoding;//编码方式

    uint32_t length;//集合包含的元素数量

    int8_t contents[];//保存元素的数组
}intset;
```
length记录集合有多少个元素，这样获取元素个数的时间复杂度就是O(1)。
set数据类型常用的命令有：sadd、spop、smembers、sunion等等。
Redis为set类型提供了求交集，并集，差集的操作，可以非常方便地实现譬如**共同关注、共同爱好、共同好友**等功能。

### zset(有序集合)
zset是Redis中比较有特色的数据类型，它和set一样是不可重复的，区别在于多了score值，用来代表排序的权重。也就是当你需要一个有序的，不可重复的集合列表时，就可以考虑使用这种数据类型。
zset的编码有两种，分别是：ziplist、skiplist。当zset的长度小于 128，并且所有元素的长度都小于 64 字节时，使用ziplist存储；否则使用 skiplist 存储。
zet常用的命令有：zadd、zrange、zrem、zcard等。

zset的特点非常适合应用于开发**排行榜**的功能。

skiplist，也就是跳跃表

```
L4   -INF ────────────────────────────────────────────────── 87
       │                              
L3   -INF ────────────────────── 24 ──────────────────────── 87
       │                         │             
L2   -INF ───────── 6 ────────── 24 ────────── 48 ────────── 87
       │            │            │             │
L1   -INF ─── 1 ─── 6 ─── 11 ─── 24 ─── 37 ─── 48 ─── 60 ─── 87
```
跳跃表的数据结构如上图所示，好处在于查询的时候，可以减少时间复杂度，如果是一个链表，要插入并且保持有序的话，那就要从头结点开始遍历，遍历到合适的位置然后插入，如果这样性能肯定是不理想的。

所以问题的关键在于**能不能像使用二分查找一样定位到插入的点**，答案就是使用**跳跃表**。
比如我们要插入38，那么查找的过程就是这样。
1. 从L4层，查询87，需要查询1次。
2. 到L3层，查询到在->24->87之间，需要查询2次。
3. 到L2层，查询->48，需要查询1次。
4. 到L1层，查询->37->48，查询2次。确定在37->48之间是插入点。

有没有发现经过L4，L3，L2层的查询后已经跳过了很多节点，当到了L1层遍历时已经把范围缩小了很多。这就是跳跃表的优势。这种方式有点类似于二分查找，所以他的**时间复杂度为O(logN)**。


## 数据库和缓存双写一致性
### 对于读请求的处理
1. 先读cache, 再读db

2. 如果, cache hit, 则直接返回数据

3. 如果, cache miss, 则访问db, 并将数据set回缓存

### 更新策略
写操作, 既要操作数据库中的数据, 又要操作缓存里的数据。
因此, 有以下两个方案:
1. 先操作数据库, 再操作缓存

2. 先操作缓存, 再操作数据库

操作缓存可以分为更新缓存和删除缓存, 故原来的两个方案可以细化为4个方案

1. 先操作数据库, 再删除缓存

2. 先删除缓存, 再操作数据库

3. 先操作数据库, 再更新缓存

4. 先更新缓存, 再操作数据库

### 1. 先操作数据库, 再删除缓存
Cache-Aside pattern 原则:
**失效**：应用程序先从cache取数据, 没有得到, 则从数据库中取数据, 成功后, 放到缓存中。
**命中**：应用程序从cache中取数据, 取到后返回。
**更新**：先把数据存到数据库中, 成功后, 再让缓存失效。

假设有两个请求, 请求A做查询操作, 请求B做更新操作, 那么会有如下情形产生
1) 缓存刚好失效
2) 请求A查询数据库, 得一个旧值
3) 请求B将新值写入数据库
4) 请求B删除缓存
5) 请求A将查到的旧值写入缓存
ok, 如果发生上述情况, 确实是会发生脏数据。
然而, 发生这种情况的概率并不高
发生上述情况有一个先天性条件, 就是步骤3的写数据库操作比步骤2的读数据库操作耗时更短, 才有可能使得步骤4先于步骤5。可是, 数据库的读操作的速度远快于写操作的（不然做读写分离干嘛, 做读写分离的意义就是因为读操作比较快, 耗资源少）, 因此步骤3耗时比步骤2更短, 这一情形很难出现。

此外, 如果缓存删除失败, 可以引入消息队列, 应用程序自己消费消息(消息里是要删除的key), 重试删除缓存, 直至成功 

### 2. 先删除缓存, 再操作数据库
假设A、B两个线程
1) 请求A进行写操作, 删除缓存
2) 请求B查询发现缓存不存在
3) 请求B去数据库查询得到旧值
4) 请求B将旧值写入缓存
5) 请求A将新值写入数据库


为了避免这个情况, 可以使用**延时双删策略**

即:
1) 先淘汰缓存
2) 再写数据库
3) 休眠, 再次淘汰缓存

休眠时间在读数据的耗时的基础上加几百ms, 如果有主从同步延时, 则睡眠时间修改为在主从同步的延时时间基础上, 加几百ms

### 3. 先操作数据库, 再更新缓存
从线程安全考虑:
假设A、B两个线程
1) A先更新数据库
2) B再更新数据库
3) B先更新缓存
4) A后更新缓存
这样就导致数据库是最新的数据, 但是缓存中是旧的脏数据。

从实际场景上考虑
1) 如果写数据库场景比较多, 而读数据场景比较少, 采用这种方案就会导致, 数据还没读到, 缓存被频繁的更新, 浪费性能。
2) 如果写入数据库的值, 不是直接写入缓存的, 而是要经过计算再写入缓存(如, 类型转换, 序列化等)。那么, 每次写入数据库后, 都再次计算写入缓存的值, 无疑是浪费性能的。此时, 删除缓存更为适合。

### 4. 先更新缓存, 再操作数据库

假设A、B两个线程
1) A先更新缓存
2) B再更新缓存
3) B先更新数据库
4) A后更新数据库
这样就导致缓存是最新的数据, 但是数据库中是旧的脏数据。

另外, 如果更新数据库失败, 则缓存里的数据就是脏数据了


### 参考文档
双写一致性部分摘抄自 [分布式之数据库和缓存双写一致性方案解析](https://zhuanlan.zhihu.com/p/48334686)

## 缓存穿透、缓存击穿、缓存雪崩

### 缓存穿透
查询**一定不存在的数据**, 因为查不到数据所以也不会写入缓存, 所以每次都会查询数据存储, 导致数据存储压力过大。
#### 解决方案
* 由于请求的参数是不合法的(每次都请求不存在的参数), 可以使用布隆过滤器(BloomFilter)或者压缩filter提前拦截, 不合法就不让这个请求到数据库层
* 从数据库找不到时, 也缓存空对象, 即: 将key-value对写为key-null。
  >这种情况一般会将空对象设置一个较短的过期时间。

### 缓存击穿
高并发下, **当某个缓存失效时, 可能出现多个进程同时查询数据存储**, 导致数据存储压力过大。

#### 解决方案
* 设置热点数据永远不过期。

* 使用互斥锁(mutex key)
比较常用的做法, 是使用mutex。就是在缓存失效的时候（判断拿出来的值为空）, 不是立即去查数据库, 而是先使用缓存工具的某些带成功操作返回值的操作（比如Redis的SETNX或者Memcache的ADD）去set一个mutex key, 当操作返回成功时, 再进行查数据库的操作并回设缓存; 否则, 就重试整个get缓存的方法。


### 缓存雪崩
高并发下, **大量缓存同时失效**, 导致大量请求同时查询数据存储, 导致数据存储压力过大。

#### 解决方案
* 设置热点数据永远不过期

* 使用多级缓存机制, 比如同时使用redsi和memcache缓存, 请求->redis->memcache->db

* 用加锁或者队列的方式保证来保证不会有大量的线程对数据库一次性进行读写, 从而避免失效时大量的并发请求落到底层存储系统上
  >加锁排队只是为了减轻数据库的压力, 并没有提高系统吞吐量。假设在高并发下, 缓存重建期间key是锁着的, 这是过来大部分请求都是阻塞的。会导致用户等待超时

* 将缓存失效时间分散开, 比如可以在原有的失效时间基础上增加一个随机值, 比如1-5分钟随机, 这样缓存过期时间的重复率就会降低, 就很难引发集体失效的事件。

## 为什么Redis集群有16384个槽
根据公式HASH_SLOT=CRC16(key) mod 16384，计算出映射到哪个分片上，然后Redis去相应的节点进行操作
**疑问**: CRC16算法产生的hash值有16bit，该算法可以产生2^16=65536个值。即，值是分布在0~65535之间。那在做mod运算的时候，为什么不mod65536，而选择mod16384?

首先, 官方回答: [why redis-cluster use 16384 slots? #2576](https://github.com/redis/redis/issues/2576)

这里原文引用一下:
```
The reason is:

Normal heartbeat packets carry the full configuration of a node, that can be replaced in an idempotent way with the old in order to update an old config. This means they contain the slots configuration for a node, in raw form, that uses 2k of space with16k slots, but would use a prohibitive 8k of space using 65k slots.
At the same time it is unlikely that Redis Cluster would scale to more than 1000 mater nodes because of other design tradeoffs.
So 16k was in the right range to ensure enough slots per master with a max of 1000 maters, but a small enough number to propagate the slot configuration as a raw bitmap easily. Note that in small clusters the bitmap would be hard to compress because when N is small the bitmap would have slots/N bits set that is a large percentage of bits set.
```

大概翻译一下:
1. 如果槽位为65536，发送心跳信息的消息头达8k，**发送的心跳包过于庞大**。
在消息头中，最占空间的是myslots[CLUSTER_SLOTS/8]。 当槽位为65536时，这块的大小是: 65536÷8÷1024=8kb
因为每秒钟，redis节点需要发送一定数量的ping消息作为心跳包，**如果槽位为65536，这个ping消息的消息头太大了，浪费带宽**。

2. redis的集群主节点数量基本不可能超过1000个。
**集群节点越多，心跳包的消息体内携带的数据越多**。redis作者在官方回答也说了, 由于存在设计权衡取舍, redis cluster节点数量不太可能超过1000个。 
因此，对于节点数在1000以内的集群，16384个槽位够用了。没有必要拓展到65536个。

3. 槽位越小，节点少的情况下，压缩比高
Redis主节点的配置信息中，它所负责的哈希槽是通过一张bitmap的形式来保存的，在传输过程中，会对bitmap进行压缩，但是如果bitmap的填充率 slots/N 很高的话(N表示节点数)，bitmap的压缩率就很低。 如果节点数很少，而哈希槽数量很多的话，bitmap的压缩率就很低。


## redis持久化
前面说过, Redis是一个基于内存的非关系型的数据库，数据保存在内存中，但是内存中的数据也容易发生丢失。

Redis提供了持久化的机制，分别是 **RDB(Redis DataBase)** 和 **AOF(Append Only File)** 。


### RDB
RDB持久化就是将当前进程的数据以**生成快照**的形式持久化到磁盘中。


RDB持久化的时候会**单独fork一个子进程**来进行持久化，因此RDB持久化有如下特点：
1. 开机恢复数据快。

2. 写入持久化文件快。

RDB的持久化也是Redis默认的持久化机制，它会把内存中的数据以快照的形式写入默认文件名为`dump.rdb`中保存。

#### 持久化触发时机
在RDB机制中触发内存中的数据进行持久化，有以下三种方式：
1. save命令：
save命令**不会fork子进程**，通过**阻塞当前Redis服务器**，直到RDB完成为止，所以该命令在生产中一般不会使用。

2. bgsave命令：
bgsave命令会**在后台fork一个与Redis主线程一模一样的子线程，由子线程负责内存中的数据持久化**。
这样fork与主线程一样的子线程消耗了内存，但是不会阻塞主线程处理客户端请求，是**以空间换时间**的方式快照内存中的数据到到文件中。
bgsave命令阻塞只会发生在fork子线程的时候，这段时间发生的非常短，可以忽略不计

3. 自动化
除了上面在命令行使用save和bgsave命令触发持久化，也可以在redis.conf配置文件中，完成配置，如下图所示
```

################################ SNAPSHOTTING  ################################
# Save the DB on disk:
#   save <seconds> <changes>

#   after 900 sec (15 min) if at least 1 key changed
save 900 1
#   after 300 sec (5 min) if at least 10 keys changed
save 300 10
#   after 60 sec if at least 10000 keys changed
save 60 10000
```

##### save和bgsave的对比区别
1. save是同步持久化数据，而bgsave是异步持久化数据。
2. save不会fork子进程，通过主进程持久化数据，会阻塞处理客户端的请求，而bdsave会fork子进程持久化数据，同时还可以处理客户端请求，高效。
3. save不会消耗内存，而bgsave会消耗内存

#### RDB的优缺点
##### 缺点
1. 如果要尽量避免丢失数据, 则RDB不适合
	>在生产快照的间隔中, 如果发生异常停机, 期间的数据就会丢失
2. 每次RDB时, 都需要fork一个子进程, 由子进程跟进持久化, 数据集较大时, fork可能会比较耗时, 造成服务器在一段时间内停止处理客户端

##### 优点
1. RBD是一个文件, 保存了某个时间点的数据集
	> 例如, 可以每个小时备份一个rdb文件, 遇到问题可以随时回退到对应时间点

2. 非常适用于灾难恢复
	>只有一个文件, 而且内容紧凑, 可以(加密后)传输到其他服务器, 然后恢复服务

3. 最大化redis性能, 父进程在保存RDB文件时, 唯一要做的就是fork一个子进程, 然后剩下的工作都由子进程处理, 父进程无需处理任何的磁盘I/O操作

4. RDB在恢复大数据集时, 速度要优于AOF

### AOF
AOF是**以日志的形式记录Redis中的每一次的增删改操作**，不会记录查询操作，以文本的形式记录，打开记录的日志文件就可以查看操作记录。**默认不开启AOF**

#### AOF触发机制
AOF带来的持久化更加安全可靠，默认提供三种触发机制，如下所示：

1. `no`：表示等操作系统等数据缓存同步到磁盘中（快、持久化没保证）。

2. `always`：同步持久化，每次发生数据变更时，就会立即记录到磁盘中（慢，安全）。

3. `everysec`：表示每秒同步一次（默认值，很快，但是会丢失一秒内的数据）。

```
# appendfsync always
appendfsync everysec
# appendfsync no
```

AOF中每秒同步是**异步**完成的，**效率很高**，由于该机制对日志文件的写入操作是采用append的形式, 因此在写入时即使宕机，也不会丢失已经存入日志文件的数据，数据的完整性是非常高的。


#### AOF重写机制
写入所有操作到日志文件时，会出现很多重复操作，甚至无效操作(如: 先i++, 再i--)，导致日志文件越来越大。记录的文件臃肿，就浪费了资源空间，所以Redis提供了rewrite机制。

redis提供了`bgrewriteaof`命令。将内存中的数据以命令的方式保存到临时文件中，同时会fork出一条新进程来将文件重写。

重写AOF的日志文件不是读取旧的日志文件瘦身，而是将内存中的数据用命令的方式重写一个AOF文件，重新保存替换原来旧的日志文件，因此内存中的数据才是最新的。

重写操作也会**fork一个子进程来处理重写操作**，重写以内存中的数据作为重写的源，避免了操作的冗余性，保证了数据的最新。

在Redis以append的形式将修改的数据写入老的磁盘中    ，同时Redis也会创建一个新的文件用于记录此期间有哪些命令被执行。


#### AOF运作方式
1. redis执行fork一个子进程
2. AOF文件内容写进临时文件
3. 父进程一边把写入指令累积到内存缓存, 一边将这些改动追加到AOF文件末尾
4. 子进程开始重写工作, 完成后, 给父进程发送信号, 父进程收到信号后, 把内存缓存的数据追加到新AOF的末尾
#### AOF的优缺点
##### 优点
1. **可靠性较高**, AOF更好保证数据不会被丢失，最多只丢失一秒内的数据
2. 通过fork一个子进程处理持久化操作，保证了主进程不会进程io操作，能高效的处理客户端的请求。
3. 另外重写操作保证了数据的有效性，即使日志文件过大也会进行重写。

4. **可读性高**, 即使某一时刻有人执行flushall清空了所有数据，只需要拿到aof的日志文件，然后把最后一条的flushall给删除掉，就可以恢复数据。

##### 缺点
对于相同数量的数据集而言，AOF文件通常要大于RDB文件。RDB 在恢复大数据集时的速度比 AOF 的恢复速度要快。AOF在运行效率上往往会慢于RDB。

	
### 相关连接
[浅谈redis持久化](https://blog.csdn.net/zzsan/article/details/119391466)
