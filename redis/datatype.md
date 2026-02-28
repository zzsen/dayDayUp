
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
| REDIS_STRING | REDIS_ENCODING_INT | 使用整形值实现的字符串对象 |
| REDIS_STRING | REDIS_ENCODING_EMBSTR | 使用embstr编码的简单动态字符串实现的字符串对象 |
| REDIS_STRING | REDIS_ENCODING_RAW | 使用简单动态字符串实现的字符串对象 |
| REDIS_LIST | REDIS_ENCODING_ZIPLIST | 使用压缩列表实现的列表对象 |
| REDIS_LIST | REDIS_ENCODING_LINKEDLIST | 使用双端链表的列表对象 | 
| REDIS_HASH | REDIS_ENCODING_ZIPLIST | 使用压缩列表实现的哈希对象 |
| REDIS_HASH | REDIS_ENCODING_HT | 使用字典实现的哈希对象 |
| REDIS_SET | REDIS_ENCODING_INTSET | 使用整数集合实现的集合对象 |
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
typedef struct dict{
    dictType *type;//类型特定函数，包括一些自定义函数，这些函数使得key和value能够存储
    void *privdata;//私有数据
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

typedef struct dictEntry{
    void *key;//键
    union{
        void *val;
        uint64_t u64;
        int64_t s64;
        double d;
    }v;//值
    struct dictEntry *next; //指向下一个节点的指针
}dictEntry;
```
结构图大致如下:
```
dict                                                            
├─ *type
├─ ht[0] ─────
├─ ht[1] ─────└─ dictht
├─ *privdata       ├─ **table ─── 1 ──────────────── 2 ───── 3 ───── ...                 
├─ rehashidx       ├─ size        │                  │                                                         
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

我们讲一讲 intset，先看源码。

``` c
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

zset 常用的命令有：zadd、zrange、zrem、zcard等。

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

