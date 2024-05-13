# mysql with 的用法 (含 with recursive)

## 相关基础

### AS 用法

as 在 mysql 中用来给列/表起别名
如:

```sql
-- 给列起别名, 把列为name的别名命名为student_name
select name as student_name from student;
-- 给表起别名, 把表student的别名命名为data_list
select * from student as data_list;
-- 给查询结果/表达式起别名
select length(name) as name_length from student;
```

### UNION 用法

union 用于结合多个 sql 查询结果于单个结果集中

```sql
query_expression_body UNION [ALL | DISTINCT] query_block
    [UNION [ALL | DISTINCT] query_expression_body]
    [...]
```

```sql
mysql> SELECT 1, 2;
+---+---+
| 1 | 2 |
+---+---+
| 1 | 2 |
+---+---+

mysql> SELECT 'a', 'b';
+---+---+
| a | b |
+---+---+
| a | b |
+---+---+

mysql> SELECT 1, 2 UNION SELECT 'a', 'b';
+---+---+
| 1 | 2 |
+---+---+
| 1 | 2 |
| a | b |
+---+---+
```

## with (Common Table Expressions)

Common Table Expressions（CTE）是一个命名的临时结果集，存在于单个语句的范围内，以后该临时结果集可以在该语句中引用, 甚至可能多次引用。
语法:

```
with_clause:
    WITH [RECURSIVE]
        cte_name [(col_name [, col_name] ...)] AS (subquery)
        [, cte_name [(col_name [, col_name] ...)] AS (subquery)] ...
```

示例:

```sql
WITH
  cte1 AS (SELECT a, b FROM table1),
  cte2 AS (SELECT c, d FROM table2)
SELECT b, d FROM cte1 JOIN cte2
WHERE cte1.a = cte2.c;
```

```sql
--下面两个查询等价
WITH cte (col1, col2) AS
(
  SELECT 1, 2
  UNION ALL
  SELECT 3, 4
)
SELECT col1, col2 FROM cte;

WITH cte AS
(
  SELECT 1 AS col1, 2 AS col2
  UNION ALL
  SELECT 3, 4
)
SELECT col1, col2 FROM cte;
```

### 用法

1. 在 select, update, delete 语句前
   ```sql
   WITH ... SELECT ...
   WITH ... UPDATE ...
   WITH ... DELETE ...
   ```
2. 在子查询前

   ```sql
   SELECT ... WHERE id IN (WITH ... SELECT ...) ...
   SELECT * FROM (WITH ... SELECT ...) AS dt ...
   ```

3. 于含 select 的语句的 select 前

   ```sql
   INSERT ... WITH ... SELECT ...
   REPLACE ... WITH ... SELECT ...
   CREATE TABLE ... WITH ... SELECT ...
   CREATE VIEW ... WITH ... SELECT ...
   DECLARE CURSOR ... WITH ... SELECT ...
   EXPLAIN ... WITH ... SELECT ...
   ```

4. 同一级别只允许一个 WITH 子句

   ```sql
   -- 错误示范
   WITH cte1 AS (...) WITH cte2 AS (...) SELECT ...
   -- 正确示范1
   WITH cte1 AS (...), cte2 AS (...) SELECT ...
   -- 正确示范2, 语句中可含有多个with, 前提是他们都在不同的层级
   WITH cte1 AS (SELECT 1)
       SELECT * FROM (WITH cte2 AS (SELECT 2) SELECT * FROM cte2 JOIN cte1) AS dt;
   ```

5. 一个 with 语句能定义一个或多个 CTE, 但每个 CTE 在该语句中都是唯一的
   ```sql
   -- 错误示范, 两个cte命名都是cte1
   WITH cte1 AS (...), cte1 AS (...) SELECT ...
   -- 正确示范
   WITH cte1 AS (...), cte2 AS (...) SELECT ...
   ```

## 递归 CTE (Recursive Common Table Expressions)

### 示例 1 递归增长

```sql
WITH RECURSIVE cte (n) AS
(
  SELECT 1
  UNION ALL
  SELECT n + 1 FROM cte WHERE n < 5
)
SELECT * FROM cte;
```

上述 sql 输出结果如下:

```sql
+------+
| n    |
+------+
|    1 |
|    2 |
|    3 |
|    4 |
|    5 |
+------+
```

该 sql 可以分成两部分, 一部分是`非递归部分`, 用于初始化行数据:

```sql
SELECT 1
```

另一部分是`递归部分`:

```sql
SELECT n + 1 FROM cte WHERE n < 5
```

等价于以下代码:

```javascript
(function test(a) {
  console.log(a);
  a++;
  if (a <= 5) {
    test(a);
  }
})(1);

// 1
// 2
// 3
// 4
// 5
```

### 示例 2 递归字符串拼接

```sql
WITH RECURSIVE cte AS
(
  SELECT 1 AS n, 'abc' AS str
  UNION ALL
  SELECT n + 1, CONCAT(str, str) FROM cte WHERE n < 3
)
SELECT * FROM cte;
```

在`非严格模式`下, 输出以下内容:

```sql
+------+------+
| n    | str  |
+------+------+
|    1 | abc  |
|    2 | abc  |
|    3 | abc  |
+------+------+
```

`严格模式`下, 则会报错: `ERROR 1406 (22001): Data too long for column 'str' at row 1`

> 定义`str`列时, 用`abc`定义, 该操作同时定义了长度为`length(abc)`, 故拼接时, 会超出长度
> 将上述 sql 调整一下

```sql
WITH RECURSIVE cte AS
(
  SELECT 1 AS n, CAST('abc' AS CHAR(20)) AS str
  UNION ALL
  SELECT n + 1, CONCAT(str, str) FROM cte WHERE n < 3
)
SELECT * FROM cte;
```

即可正常输出:

```sql
+------+--------------+
| n    | str          |
+------+--------------+
|    1 | abc          |
|    2 | abcabc       |
|    3 | abcabcabcabc |
+------+--------------+
```

### 限制 CTE 循环次数

输入以下 sql 时, 会提示`Recursive query aborted after 1048577 iterations. Try increasing @@cte_max_recursion_depth to a larger value.`

```sql
WITH RECURSIVE cte (n) AS
(
  SELECT 1
  UNION ALL
  SELECT n + 1 FROM cte
)
SELECT * FROM cte;
```

默认情况下, `cte_max_recursion_depth`的值为 1000, 会限制 CTE 的循环次数, 可以通过修改`cte_max_recursion_depth`修改循环次数上限.

通过修改`cte_max_recursion_depth`修改循环次数上限后, 可通过`limit`限制上限.

> `cte_max_recursion_depth`>=`limit`
> 故当 limit 较大时, 需先修改`cte_max_recursion_depth`, 否则较大的`limit`不生效

```sql
WITH RECURSIVE cte (n) AS
(
  SELECT 1
  UNION ALL
  SELECT n + 1 FROM cte LIMIT 10000
)
SELECT * FROM cte;
```

### 斐波那契数列

```sql
WITH RECURSIVE fibonacci (n, fib_n, next_fib_n) AS
(
  SELECT 1, 0, 1
  UNION ALL
  SELECT n + 1, next_fib_n, fib_n + next_fib_n
    FROM fibonacci WHERE n < 10
)
SELECT * FROM fibonacci;
```

```sql
+------+-------+------------+
| n    | fib_n | next_fib_n |
+------+-------+------------+
|    1 |     0 |          1 |
|    2 |     1 |          1 |
|    3 |     1 |          2 |
|    4 |     2 |          3 |
|    5 |     3 |          5 |
|    6 |     5 |          8 |
|    7 |     8 |         13 |
|    8 |    13 |         21 |
|    9 |    21 |         34 |
|   10 |    34 |         55 |
+------+-------+------------+
```

### 日期序列生成

```sql
mysql> SELECT * FROM sales ORDER BY date, price;
+------------+--------+
| date       | price  |
+------------+--------+
| 2022-01-03 | 100.00 |
| 2022-01-03 | 200.00 |
| 2022-01-06 |  50.00 |
| 2022-01-08 |  10.00 |
| 2022-01-08 |  20.00 |
| 2022-01-08 | 150.00 |
| 2022-01-17 |   5.00 |
+------------+--------+
```

求每日总`sales`时

```sql
mysql> SELECT date, SUM(price) AS sum_price
       FROM sales
       GROUP BY date
       ORDER BY date;
+------------+-----------+
| date       | sum_price |
+------------+-----------+
| 2022-01-10 |    300.00 |
| 2022-01-13 |     50.00 |
| 2022-01-15 |    180.00 |
| 2022-01-17 |      5.00 |
+------------+-----------+
```

这样产生的结果, 中间会缺少部分日期.
先写个 sql, 根据日期, 输出中间的日期列表:

```sql
WITH RECURSIVE dates (date) AS
(
  SELECT MIN(date) FROM sales
  UNION ALL
  SELECT date + INTERVAL 1 DAY FROM dates
  WHERE date + INTERVAL 1 DAY <= (SELECT MAX(date) FROM sales)
)
SELECT * FROM dates;
```

```sql
+------------+
| date       |
+------------+
| 2022-01-10 |
| 2022-01-11 |
| 2022-01-12 |
| 2022-01-13 |
| 2022-01-14 |
| 2022-01-15 |
| 2022-01-16 |
| 2022-01-17 |
+------------+
```

结合上述 sql:

```sql
WITH RECURSIVE dates (date) AS
(
  SELECT MIN(date) FROM sales
  UNION ALL
  SELECT date + INTERVAL 1 DAY FROM dates
  WHERE date + INTERVAL 1 DAY <= (SELECT MAX(date) FROM sales)
)
SELECT dates.date, COALESCE(SUM(price), 0) AS sum_price
FROM dates LEFT JOIN sales ON dates.date = sales.date
GROUP BY dates.date
ORDER BY dates.date;
```

```sql
+------------+-----------+
| date       | sum_price |
+------------+-----------+
| 2022-01-10 |    300.00 |
| 2022-01-11 |      0.00 |
| 2022-01-12 |      0.00 |
| 2022-01-13 |     50.00 |
| 2022-01-14 |      0.00 |
| 2022-01-15 |    180.00 |
| 2022-01-16 |      0.00 |
| 2022-01-17 |      5.00 |
+------------+-----------+
```

### 分层数据遍历

简单写个 sql, 建表并插入数据

```sql
CREATE TABLE employees (
  id         INT PRIMARY KEY NOT NULL,
  name       VARCHAR(100) NOT NULL,
  manager_id INT NULL,
  INDEX (manager_id),
FOREIGN KEY (manager_id) REFERENCES employees (id)
);
INSERT INTO employees VALUES
(117, "Zzs", NULL),      # zzs is the boss (manager_id is NULL)
(198, "John", 117),      # John has ID 198 and reports to 117 (zzs)
(692, "Tarek", 117),
(29, "Pedro", 198),
(4610, "Sarah", 29),
(72, "Pierre", 29),
(123, "Adil", 692);
```

此时数据库内数据如下:

```sql
mysql> SELECT * FROM employees ORDER BY id;
+------+---------+------------+
| id   | name    | manager_id |
+------+---------+------------+
|   29 | Pedro   |        198 |
|   72 | Pierre  |         29 |
|  117 | Zzs     |       NULL |
|  123 | Adil    |        692 |
|  198 | John    |        117 |
|  692 | Tarek   |        117 |
| 4610 | Sarah   |         29 |
+------+---------+------------+
```

通过以下 sql, 查询出管理链路:

```sql
WITH RECURSIVE employee_paths (id, name, path) AS
(
  SELECT id, name, CAST(id AS CHAR(200))
    FROM employees
    WHERE manager_id IS NULL
  UNION ALL
  SELECT e.id, e.name, CONCAT(ep.path, ',', e.id)
    FROM employee_paths AS ep JOIN employees AS e
      ON ep.id = e.manager_id
)
SELECT * FROM employee_paths ORDER BY path;
```

查询结果如下:

```sql
+------+---------+-----------------+
| id   | name    | path            |
+------+---------+-----------------+
|  117 | Zzs     | 117             |
|  198 | John    | 117,198         |
|   29 | Pedro   | 117,198,29      |
| 4610 | Sarah   | 117,198,29,4610 |
|   72 | Pierre  | 117,198,29,72   |
|  692 | Tarek   | 117,692         |
|  123 | Adil    | 117,692,123     |
+------+---------+-----------------+
```

## 参考文档

[WITH (Common Table Expressions)](https://dev.mysql.com/doc/refman/8.2/en/with.html)
