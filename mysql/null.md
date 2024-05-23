# 浅谈mysql中的null
照旧, 在开始前, 先附上本次试验的ddl, 然后插入数据, 随机抽取几条幸运数据的name设为null
``` sql
CREATE TABLE `user` (
  `id` int NOT NULL COMMENT 'id',
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL COMMENT '姓名',
  `age` int NOT NULL COMMENT '年龄',
  `sex` tinyint(1) NOT NULL COMMENT '性别',
  `phone` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL COMMENT '电话',
  PRIMARY KEY (`id`),
  KEY `idx_name_age_sex` (`name`,`age`,`sex`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
```

## MySQL中IS NULL、IS NOT NULL、!= 等能不能用索引？
先说结论, 能！能！能！（重要的事情说三遍）

### 1. is null
```sql
EXPLAIN SELECT * FROM user WHERE name IS NULL
```
---

1. 表中满足筛选条件的数据量较多时, 不走索引

    当表中只有1条数据`name`不为null, 其余数据均为null时

    |id|select_type|table|partitions|type|possible_keys|key|key_len|ref|rows|filtered|Extra|
    |--|--|--|--|--|--|--|--|--|--|--|--|
    |1|SIMPLE|user| $\color{#909399}{(Null)}$ | $\color{#f56c6c}{ALL}$ |idx\_name\_age\_sex| $\color{#f56c6c}{(Null)}$ | $\color{#909399}{(Null)}$ | $\color{#909399}{(Null)}$ |$\color{#f56c6c}{99}$|98.99|Using where|

2. 表中满足筛选条件的数据量较少时, 走索引

    当表中只有17条数据`name`为null时

    |id|select_type|table|partitions|type|possible_keys|key|key_len|ref|rows|filtered|Extra|
    |--|--|--|--|--|--|--|--|--|--|--|--|
    |1|SIMPLE|user| $\color{#909399}{(Null)}$ |ref| $\color{#f56c6c}{idx\_name\_age\_sex}$ | $\color{#f56c6c}{idx\_name\_age\_sex}$ |1023|const|$\color{#f56c6c}{17}$|100.00|Using index condition|

### 2. is not null

```sql
EXPLAIN SELECT * FROM user WHERE name IS NOT NULL
```
---

1. 表中满足筛选条件的数据量较多时, 不走索引

    当表中只有17条数据`name`为null, 其余数据均不为null时

    |id|select_type|table|partitions|type|possible_keys|key|key_len|ref|rows|filtered|Extra|
    |--|--|--|--|--|--|--|--|--|--|--|--|
    |1|SIMPLE|user| $\color{#909399}{(Null)}$ | $\color{#f56c6c}{ALL}$ |idx\_name\_age\_sex| $\color{#f56c6c}{(Null)}$ | $\color{#909399}{(Null)}$ | $\color{#909399}{(Null)}$ |$\color{#f56c6c}{99}$|82.83|Using where|

2. 表中满足筛选条件的数据量较少时, 走索引

    当表中只有1条数据`name`不为null时

    |id|select_type|table|partitions|type|possible_keys|key|key_len|ref|rows|filtered|Extra|
    |--|--|--|--|--|--|--|--|--|--|--|--|
    |1|SIMPLE|user| $\color{#909399}{(Null)}$ |ref| $\color{#f56c6c}{idx\_name\_age\_sex}$ | $\color{#f56c6c}{idx\_name\_age\_sex}$ |1023|const|$\color{#f56c6c}{1}$|100.00|Using index condition|

### 3. !=

```sql
EXPLAIN SELECT * FROM user WHERE name  != 'user1'
```

1. 表中为null数据较多时, 走索引

    当表中只有1条数据`name`不为null, 其余数据均为null时

    |id|select_type|table|partitions|type|possible_keys|key|key_len|ref|rows|filtered|Extra|
    |--|--|--|--|--|--|--|--|--|--|--|--|
    |1|SIMPLE|user| $\color{#909399}{(Null)}$ |range| $\color{#f56c6c}{idx\_name\_age\_sex}$ | $\color{#f56c6c}{idx\_name\_age\_sex}$ |1023| $\color{#909399}{(Null)}$ |$\color{#f56c6c}{2}$|100.00|Using index condition|


2. 表中为null数据较少时, 不走索引

    当表中只有17条数据`name`为null, 其余数据均不为null时

    |id|select_type|table|partitions|type|possible_keys|key|key_len|ref|rows|filtered|Extra|
    |--|--|--|--|--|--|--|--|--|--|--|--|
    |1|SIMPLE|user| $\color{#909399}{(Null)}$ | $\color{#f56c6c}{ALL}$ |idx\_name\_age\_sex| $\color{#f56c6c}{(Null)}$ | $\color{#909399}{(Null)}$ | $\color{#909399}{(Null)}$ |$\color{#f56c6c}{99}$|82.83|Using where|


### 4. =

```sql
EXPLAIN SELECT * FROM user WHERE name  != 'user1'
```

不管为null数据量多少, 都走索引

|id|select_type|table|partitions|type|possible_keys|key|key_len|ref|rows|filtered|Extra|
|--|--|--|--|--|--|--|--|--|--|--|--|
|1|SIMPLE|user| $\color{#909399}{(Null)}$ |ref| $\color{#f56c6c}{idx\_name\_age\_sex}$ | $\color{#f56c6c}{idx\_name\_age\_sex}$ |1023|const|$\color{#f56c6c}{1}$|100.00| $\color{#909399}{(Null)}$ |

### 综述
当索引列运行为null时, 还是能走索引的, 但具体走不走索引, 还得取决于 **走索引和不走索引哪个都成本较小**。 
注意, 这里说的走不走索引, 都是指的**非聚簇索引**。

**非聚簇索引的扫描成本**由两部分组成：
1. 读取索引记录成本
2. 回表成本 (通过非聚簇索引得到的主键到聚簇索引中查询)

**如果查询读取的非聚簇索引越多, 需要回表查询的次数就越多, 当达到一定比例后, 走索引的成本比不走索引的成本还高时, 最终的查询就不走索引了, 这也就是索引失效的原因**


## 为什么MySQL不建议使用NULL作为列默认值？

### 数据行中, null是如何存储的

InnoDB有4中行格式：
* **Redundant** : 非紧凑格式,5.0 版本之前用的行格式,目前很少使用,
* **Compact** : 紧凑格式,5.1 版本之后默认行格式,可以存储更多的数据
* **Dynamic** , **Compressed** : 和Compact类似,5.7 版本之后默认使用 Dynamic 行格式,在Compact基础上做了改进,基础设计原理没变

由于Redundant较少使用, 且Dynamic和Compressed是基于Compact的, 故这里以Compact为例。
Compact行格式如下：

<table border>
  <tr>
    <th style="text-align: center" colspan="3">存储的额外数据</th>
    <th style="text-align: center" colspan="7">存储的真实数据</th>
  <tr>
  <tr>
    <td>变长数据列的长度</td><td>NULL值的列表</td><td>记录头信息</td>
    <td>row_id <span style="color:#666">(隐藏字段)</span></td>
    <td>trx_id <span style="color:#666">(隐藏字段)</span></td>
    <td>roll_ptr <span style="color:#666">(隐藏字段)</span></td>
    <td>列1</td>
    <td>列2</td>
    <td>...</td>
    <td>列n</td>
  </tr>
</table>

由于：
1. `NULL值列表`的数量, 与允许为null的字段数量一致
    >如, 有7个字段允许为null, 则有7个`NULL值列表`
2. `NULL值列表`至少占用1字节空间, 故当数据量越大或null列越多时, 占用的存储空间越多

综上所述, 不建议允许列为null, 可使用其他默认值（如空字符串, 0等）