## 事务

### 特性ACID

关系性数据库需要遵循ACID规则，具体内容如下：

1. 原子性 atomicity：

   要么全部提交成功，要么全部失败回滚

2. 一致性 consistency：

   事务执行前后，数据库都必须处于一致性状态

3. 隔离性 isolation：

   一个事务所做的修改在最终提交以前，对其他事务是不可见的

4. 持久性 durability：

   事务提交后, 对数据的改变是持久的，即使数据库发生故障也不对其有任何影响。

### 4个隔离级别

1. 读未提交 Read Uncommitted

   **隔离级别最低**

   事务中的修改，即使没有提交，对其他事务也都是可见的

2. 读已提交 Read Committed

   **大多数数据库的默认隔离级别**

   事务一旦提交，该事务所作的修改对其他正在进行中的事务就是可见的。

3. 可重复读 Repeatable Read

   **MySQL的默认隔离级别**

   同一个事务中多次读取同样记录结果是一致的

4. 可序列化 Serializable

   **隔离级别最高**

   事务只能串行执行，不能并发执行

### 如果不考虑隔离性，会发生什么事呢？

1. 脏读
   事务可以读取未提交的数据，而该数据可能在未来因回滚而消失
   
2. 不可重复读
	事务内的多次查询却返回了不同的结果，这是由于在查询过程中，数据被另外一个事务修改并提交了。
	
3. 幻读
   事务在读取目标范围内的记录时，另一个事务又在该范围内插入了新的记录，当之前的事务再次读取该范围的记录时，会产生第一次读取范围时不存在的幻行
   
   
### 总结

| 隔离级别/能解决的问题 | 脏读 | 不可重复读 | 幻读 |
| --------------------- | ---- | ---------- | ---- |
| 读未提交              | -    | -          | -    |
| 读已提交              | Y    | -          | -    |
| 可重复读              | Y    | Y          | -    |
| 可序列化              | Y    | Y          | Y    |

   不可重复读和幻读比较相似, 区别在于:
   1. 解决**不可重复读**的方法是 **锁行**，解决**幻读**的方式是 **锁表**。
   2. **不可重复读**  读到其他事务**update/delete**后已提交的数据
      **幻读** 读到其他事务**insert**已经提交的数据

## 锁

## 索引

## 日志

### 日志类型

* redo log

* undo log

* binlog

* errorlog 

* slow query log

* general log

* relay log

### 谈谈redo log 、 undo log 和 binlog的异同

#### 1. 实现层级

**binlog**是mysql**服务层**实现的

**redolog**和**undolog**是**引擎层**实现的, **只存在于innodb中**，myisam引擎并没有实现, 统称为事务日志

#### 2. 用途
* redo log
   **确保事务的持久性**
   
   
   
* undo log
	* **用于回滚**
	* **MVCC**	
  
	
	
* binlog
	* **复制**：MySQL Replication在Master端开启binlog，Master把它的二进制日志传递给slaves并回放来达到master-slave数据一致的目的
	* **数据恢复**：通过mysqlbinlog工具恢复数据
	* **增量备份**

#### 3.存储内容、格式
   * redo log
     **物理日志**
     
     
     
   * undo log
     **逻辑日志**
     
  * binlog
     **逻辑日志**
     
     

#### 4. 生命周期
* redolog
  **事务开始之后**，就开始产生 redo log 日志了
  
* undolog
	**事务开始之前**，将当前事务版本生成 undo log
	
* binlog
	**事务提交的时候**，一次性将事务中的 所有sql 语句按照一定的格式记录到 binlog 中

### 两阶段提交
   将数据页加载到内存 → 修改数据 → 更新数据 → **写redolog（状态为prepare）** → 写binlog → **提交事务**（**数据写入成功后将redo log状态改为commit**）

### 其他相关文档建议

[浅谈mysql日志系统](https://blog.csdn.net/zzsan/article/details/118397623)
[腾讯工程师带你深入解析 MySQL binlog](https://zhuanlan.zhihu.com/p/33504555)
[详细分析MySQL事务日志(redo log和undo log)](https://www.cnblogs.com/f-ck-need-u/p/9010872.html)
