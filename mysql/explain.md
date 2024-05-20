# explain执行计划
explain是mysql中一关键字, 用于**查看执行计划**, 模拟执行器执行sql查询语句, 从而分析sql语句或表结构的性能瓶颈或优化方向。

## explain用途
可分析得到以下信息：

1. 表读取顺序
2. 数据读取操作的操作类型
3. 可使用的索引
4. 实际使用的索引
5. 表间引用
6. 遍历数据行数

## explain字段
explain结果返回以下字段

|id|select_type|table|partitions|type|possible_keys|key|key_len|ref|rows|filtered|Extra|
|--|--|--|--|--|--|--|--|--|--|--|--|

### 实例
下列说明中的查询, 以该表结构为例
``` sql
CREATE TABLE `user` (
  `id` int NOT NULL COMMENT 'id',
  `name` varchar(255) COLLATE utf8mb4_bin NOT NULL COMMENT '姓名',
  `age` int NOT NULL COMMENT '年龄',
  `sex` tinyint(1) NOT NULL COMMENT '性别',
  `phone` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL COMMENT '电话',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_phone` (`phone`),
  KEY `idx_name` (`name`),
  KEY `idx_name_age_sex` (`name`,`age`,`sex`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
```

### id

`id`为select的序列号, 有几个select, 就有几个id

* id值不同：如果是只查询, id的序号会递增, **id值越大优先级越高, 越先被执行**；
* id值相同：从上往下依次执行；
* id列为null：表示这是一个结果集, 不需要使用它来进行查询。

### select_type 查询类型

1. simple

    表示查询中不包含union操作或子查询

    ``` sql
    explain select * from user where id = 1
    ```

    |id|select_type|table|partitions|type|possible_keys|key|key_len|ref|rows|filtered|Extra|
    |--|--|--|--|--|--|--|--|--|--|--|--|
    |1| $\color{#f56c6c}{SIMPLE}$ |user| $\color{#909399}{(Null)}$ |const|PRIMARY|PRIMARY|4|const|1|100.00| $\color{#909399}{(Null)}$ |

2. primary

    需要union或者含子查询的select, 位于最外层查询的select_type即为primary, 有且只有一个
    
    ``` sql
    explain select * from user where id = 1 union select * from user where id = 2
    ```

    |id|select_type|table|partitions|type|possible_keys|key|key_len|ref|rows|filtered|Extra|
    |--|--|--|--|--|--|--|--|--|--|--|--|
    |1| $\color{#f56c6c}{PRIMARY}$ |user| $\color{#909399}{(Null)}$ |const|PRIMARY|PRIMARY|4|const|1|100.00| $\color{#909399}{(Null)}$ |
    |2|UNION|user| $\color{#909399}{(Null)}$ |const|PRIMARY|PRIMARY|4|const|1|100.00| $\color{#909399}{(Null)}$ |
    | $\color{#909399}{(Null)}$ |UNION RESULT|<union1,2>| $\color{#909399}{(Null)}$ |ALL| $\color{#909399}{(Null)}$ | $\color{#909399}{(Null)}$ | $\color{#909399}{(Null)}$ | $\color{#909399}{(Null)}$ | $\color{#909399}{(Null)}$ | $\color{#909399}{(Null)}$ |Using temporary|

3. subquery
    
    子查询

    ``` sql
    explain select * from user where id = (select id from user where id = 1)
    ```

    |id|select_type|table|partitions|type|possible_keys|key|key_len|ref|rows|filtered|Extra|
    |--|--|--|--|--|--|--|--|--|--|--|--|
    |1|PRIMARY|user| $\color{#909399}{(Null)}$ |const|PRIMARY|PRIMARY|4|const|1|100.00| $\color{#909399}{(Null)}$ |
    |2| $\color{#f56c6c}{SUBQUERY}$ |user| $\color{#909399}{(Null)}$ |const|PRIMARY|PRIMARY|4|const|1|100.00| using index|
    

4. union

    需要union的select
    
    ``` sql
    explain select * from user where id = 1 union select * from user where id = 2
    ```

    |id|select_type|table|partitions|type|possible_keys|key|key_len|ref|rows|filtered|Extra|
    |--|--|--|--|--|--|--|--|--|--|--|--|
    |1|PRIMARY|user| $\color{#909399}{(Null)}$ |const|PRIMARY|PRIMARY|4|const|1|100.00| $\color{#909399}{(Null)}$ |
    |2| $\color{#f56c6c}{UNION}$ |user| $\color{#909399}{(Null)}$ |const|PRIMARY|PRIMARY|4|const|1|100.00| $\color{#909399}{(Null)}$ |
    | $\color{#909399}{(Null)}$ |UNION RESULT|<union1,2>| $\color{#909399}{(Null)}$ |ALL| $\color{#909399}{(Null)}$ | $\color{#909399}{(Null)}$ | $\color{#909399}{(Null)}$ | $\color{#909399}{(Null)}$ | $\color{#909399}{(Null)}$ | $\color{#909399}{(Null)}$ |Using temporary|

5. union result

    从union表获取结果的select, 由于不参与查询, 故id为null
    
    ``` sql
    explain select * from user where id = 1 union select * from user where id = 2
    ```

    |id|select_type|table|partitions|type|possible_keys|key|key_len|ref|rows|filtered|Extra|
    |--|--|--|--|--|--|--|--|--|--|--|--|
    |1|PRIMARY|user| $\color{#909399}{(Null)}$ |const|PRIMARY|PRIMARY|4|const|1|100.00| $\color{#909399}{(Null)}$ |
    |2|UNION|user| $\color{#909399}{(Null)}$ |const|PRIMARY|PRIMARY|4|const|1|100.00| $\color{#909399}{(Null)}$ |
    | $\color{#909399}{(Null)}$ | $\color{#f56c6c}{UNION}$ $\color{#f56c6c}{RESULT}$ |<union1,2>| $\color{#909399}{(Null)}$ |ALL| $\color{#909399}{(Null)}$ | $\color{#909399}{(Null)}$ | $\color{#909399}{(Null)}$ | $\color{#909399}{(Null)}$ | $\color{#909399}{(Null)}$ | $\color{#909399}{(Null)}$ |Using temporary|

### table 当前查询正在访问的数据表

* 如果查询使用了别名, 则table显示的是别名
* 如果不涉及对数据表的操作, 则table为null
* 如果结果为<X,Y>, XY为执行计划的id, 表示结果来自该查询

### type 查询范围

从好到坏以此为:
```
system > const > eq_ref >ref > range > index > ALL
```

1. system

    表中只有一行数据（等于系统表）, 这是const 类型的特例, 平时不会出现, 可以忽略不计。

2. const
    
    **使用唯一索引或者主键**, 表示通过索引一次就找到了, const用于比较primary key 或者 unique索引。因为只需匹配一行数据, 所有很快。如果将主键置于where列表中, mysql就能将该查询转换为一个const。

3. eq_ref

    唯一性索引扫描, 对于每个索引键, 表中只有一行数据与之匹配。常见于主键或唯一索引扫描。

    ``` sql
    explain select * from user a1,user a2  where a1.id = a2.id
    ```

    |id|select_type|table|partitions|type|possible_keys|key|key_len|ref|rows|filtered|Extra|
    |--|--|--|--|--|--|--|--|--|--|--|--|
    |1|SIMPLE|a1| $\color{#909399}{(Null)}$ |ALL|PRIMARY| $\color{#909399}{(Null)}$ | $\color{#909399}{(Null)}$ | $\color{#909399}{(Null)}$ |99|100.00| $\color{#909399}{(Null)}$ |
    |2|SIMPLE|a2| $\color{#909399}{(Null)}$ | $\color{#f56c6c}{eq\_ref}$ |PRIMARY|PRIMARY|4| $\color{#f56c6c}{test.a1.id}$ |1|100.00| $\color{#909399}{(Null)}$ |

4. ref

    **非唯一性索引扫描**, 返回匹配某个单独值的所有行。本质也是一种索引。

    ``` sql
    explain select * from user where name = 'user1'
    ```

    |id|select_type|table|partitions|type|possible_keys|key|key_len|ref|rows|filtered|Extra|
    |--|--|--|--|--|--|--|--|--|--|--|--|
    |1|SIMPLE|user| $\color{#909399}{(Null)}$ | $\color{#f56c6c}{ref}$ |idx_name,idx_name_age_sex|idx_name|1022|const|1|100.00| $\color{#909399}{(Null)}$ |

5. range

    索引范围扫描, 常见于使用>,<,between ,in ,like等运算符的查询中。
    
    ``` sql
    explain select * from user where id in (1,2)
    ```

    |id|select_type|table|partitions|type|possible_keys|key|key_len|ref|rows|filtered|Extra|
    |--|--|--|--|--|--|--|--|--|--|--|--|
    |1|SIMPLE|user| $\color{#909399}{(Null)}$ | $\color{#f56c6c}{range}$ |PRIMARY|PRIMARY|4| $\color{#909399}{(Null)}$ |2|100.00|Using where|

6. index

    **索引全表扫描, 把索引树从头到尾扫一遍**。index与ALL区别为index类型只遍历索引树。这通常比ALL快, 因为索引文件通常比数据文件小。(也就是说虽然all和Indx都是读全表, 但index是从索引中读取的, 而all是从硬盘中读的)

7. all
    
    **全表扫描**（Index与ALL虽然都是读全表, 但index是从索引中读取, 而ALL是从硬盘读取）

8. NULL

    MySQL在优化过程中分解语句, 执行时甚至不用访问表或索引。

### possible_keys 查询可能使用到的索引

### key 查询实际使用的索引

    select_type为index_merge时, 这里可能出现两个以上的索引, 其他的select_type这里只会出现一个。

### key_len 用于处理查询的索引长度

用于处理查询的索引长度, 表示索引中使用的**字节数**。通过这个值, 可以得出一个多列索引里实际使用了哪一部分。
    
> key_len显示的值为索引字段的最大可能长度, 并非实际使用长度, 即key_len是根据表定义计算而得, 不是通过表内检索出的。另外, key_len只计算where条件用到的索引长度, 而排序和分组就算用到了索引, 也不会计算到key_len中。

### ref

显示索引的哪一列被使用了, 如果可能的话, 是一个常数。哪些列或常量被用于查找索引列上的值


``` sql
explain select * from user a1,user a2  where a1.id = a2.id
```

|id|select_type|table|partitions|type|possible_keys|key|key_len|ref|rows|filtered|Extra|
|--|--|--|--|--|--|--|--|--|--|--|--|
|1|SIMPLE|a1| $\color{#909399}{(Null)}$ |ALL|PRIMARY| $\color{#909399}{(Null)}$ | $\color{#909399}{(Null)}$ | $\color{#909399}{(Null)}$ |99|100.00| $\color{#909399}{(Null)}$ |
|2|SIMPLE|a2| $\color{#909399}{(Null)}$ | eq_ref |PRIMARY|PRIMARY|4| $\color{#f56c6c}{test.a1.id}$ |1|100.00| $\color{#909399}{(Null)}$ |


### rows

表示MySQL根据表统计信息及索引选用情况, 大致估算的找到所需的目标记录所需要读取的行数, 不是精确值。

### filtered

filtered列返回的是根据条件筛选的百分比值, 最大值为100, 意味着没有对其进行筛选。数值越小, 说明筛选量越大。rows表示大致遍历的数量, rows x filtered 表示表连接的行数。例如, 当rows为100, filtered为50时, 与表连接的行数为100 * 50% = 500
> The filtered column indicates an estimated percentage of table rows filtered by the table condition. The maximum value is 100, which means no filtering of rows occurred. Values decreasing from 100 indicate increasing amounts of filtering. rows shows the estimated number of rows examined and rows × filtered shows the number of rows joined with the following table. For example, if rows is 1000 and filtered is 50.00 (50%), the number of rows to be joined with the following table is 1000 × 50% = 500.

### extra
额外信息。

|参数|说明|
|--|--|
|Using index|查询覆盖了索引, 不需要读取数据文件, **从索引树（索引文件）中即可获得信息**。如果同时出现using where, 表明索引被用来执行索引键值的查找, 没有using where, 表明索引用来读取数据而非执行查找动作。这是MySQL服务层完成的, 但无需再回表查询记录。|
|Using index condition|**查询的列未被索引覆盖, where筛选条件是索引的前导列**。这是MySQL 5.6出来的新特性, 叫做“索引条件推送”。简单说一点就是MySQL原来在索引上是不能执行如like这样的操作的, 但是现在可以了, 这样减少了不必要的IO操作, 但是只能用在二级索引上。|
|Using filesort|MySQL有两种方式可以生成有序的结果, 通过**排序操作**或者**使用索引**, 当Extra中出现了Using filesort 说明MySQL使用了后者, 但注意虽然叫filesort但并不是说明就是用了文件来进行排序, 只要可能排序都是在内存里完成的。大部分情况下利用索引排序更快, 所以一般这时也要考虑优化查询了。使用文件完成排序操作, 这是可能是ordery by, group by语句的结果, 这可能是一个CPU密集型的过程, 可以通过选择合适的索引来改进性能, 用索引来为查询结果排序。|
|Using temporary|用临时表保存中间结果, 常用于GROUP BY 和 ORDER BY操作中, 一般看到它说明查询需要优化了, 就算避免不了临时表的使用也要尽量避免硬盘临时表的使用。|
|Not exists|MYSQL优化了LEFT JOIN, 一旦它找到了匹配LEFT JOIN标准的行,  就不再搜索了。|
|Using where|使用了WHERE从句来限制哪些行将与下一张表匹配或者是返回给用户。注意：Extra列出现Using where表示MySQL服务器将存储引擎返回服务层以后再应用WHERE条件过滤。|
|Using join buffer|使用了连接缓存：Block Nested Loop, 连接算法是块嵌套循环连接;Batched Key Access, 连接算法是批量索引连接|
|impossible where|where子句的值总是false, 不能用来获取任何元组|
|select tables optimized away|在没有GROUP BY子句的情况下, 基于索引优化MIN/MAX操作, 或者对于MyISAM存储引擎优化COUNT(*)操作, 不必等到执行阶段再进行计算, 查询执行计划生成的阶段即完成优化。|
|distinct|优化distinct操作, 在找到第一匹配的元组后即停止找同样值的动作|

## exlpain的局限性：
* EXPLAIN不会告诉关于触发器、存储过程的信息或用户自定义函数对查询的影响情况；
* EXPLAIN不考虑各种Cache；
* EXPLAIN不能显示MySQL在执行查询时所作的优化工作；
* 部分统计信息是估算的, 并非精确值；
* EXPALIN只能解释SELECT操作, 其他操作要重写为SELECT后查看。