
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

3. 槽位越小，bitmap 体积越小，即使压缩效果差也能接受

      Redis主节点的配置信息中，它所负责的哈希槽是通过一张bitmap的形式来保存的，在传输过程中，会对bitmap进行压缩。但bitmap的压缩效果取决于填充率（slots/N，N 为节点数）：当节点数较少时，每个节点负责的槽位占比高，bitmap中大量bit被置为1，压缩效果很差。在这种情况下，16384个槽的bitmap原始大小仅2KB，即使压缩率低也完全可以接受；而65536个槽的bitmap原始大小为8KB，压缩效果同样差的情况下，带宽开销就难以接受了。
