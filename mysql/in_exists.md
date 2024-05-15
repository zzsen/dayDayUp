# MySQL in和exists查询对比
## sql实例
**外表**：tableA
**内表**：tableB
**IN**：`select * from tableA where tableA.id IN ( select A_Id from tableB )`
**EXISTS**：`select * from tableA where EXISTS ( select * from tableB where tableB.A_Id = tableA.id )`

## 执行原理
in的执行原理，是把外表和内表做hash连接，**主要用到外表索引**
exists的执行原理，是对外表做loop循环，**主要用到内表索引**

## In 和 Exists 分析
>以下部分主要引用自https://blog.csdn.net/kk123k/article/details/80614956，稍作改动，如有侵权，烦请联系我删除

### In 查询
第1条查询其实就相当于or语句，假设B表有A_Id分别为id1，id2 ... idN这N条记录，那么上面语句可以等价转化成：
```sql
select * from tableA where tableA.id = id1 or tableA.id = id2 or ... or tableA.id = idN;
```
**主要是用到了A表的索引**，B表的大小对查询效率影响不大



### Exists 查询
第2条查询，可以理解为类似下列流程
```javascript
function GetExists () {
	var resultSet = []
	for(var i = 0; i< ResultA.length; i++){
		//从tableA逐条获取记录
		var dataA = getId(tableA, i)
		if(tableB.A_Id === dataA.id){
			result.push(dataA)
		}
	}
	return resultSet
}
```

**主要是用到了B表的索引**，A表大小对查询效率影响不大


## Not In 和 Not Exists 分析
```sql
select * from tableA where tableA.id NOT IN ( select A_Id from tableB )
```
```sql
select * from tableA where not exists ( select * from tableB where tableB.A_Id = tableA.id )
```
### Not In 查询
类似，第1条查询相当于and语句，假设B表有A_Id分别为id1，id2 ... idN这N条记录，那么上面语句可以等价转化成：
```sql
select * from tableA where tableA.id != id1 and tableA.id != id2 and ... and tableA.id != idN
```

not in是个范围查询，由于! =不能使用任何索引，故A表的每条记录，都要在B表里遍历一次，查看B表里是否存在这条记录，而not exists还是和上面一样，用了B的索引，所以**无论什么情况，not exists都比not in效率高**

总结：
**如果查询的两个表大小相当，那么用in和exists效率差别不大**。

如果两个表中一个较小，一个较大，则子查询表大的用exists，子查询表小的用in。
例如：有表A （**小**表）和表B（**大**表）

1：

`select * from A where cc in (select cc from B)` 效率低，用到了A表上cc列的索引；
 
`select * from A where exists(select cc from B where cc=A.cc)` 效率高，用到了B表上cc列的索引。

2：

`select * from B where cc in (select cc from A) `效率高，用到了B表上cc列的索引；
 
`select * from B where exists(select cc from A where cc=B.cc)` 效率低，用到了A表上cc列的索引。

not in 和not exists如果查询语句使用了not in 那么内外表都进行全表扫描，没有用到索引；而not extsts 的子查询依然能用到表上的索引。**所以无论那个表大，用not exists都比not in要快**。