# DDL、DML和DQL
 
 **DDL**（Data Define Languge）：数据定义语言, 用于**库和表**的创建、修改、删除
 
 **DML** (Data Manipulate Language)：数据操纵语言, 用于**添加、删除、修改数
    据库记录**, 并检查数据完整性
 
 **DQL**（Data Query Language）：数据查询语言, 用来**查询数据库中表的记录(数据)**

# DDL
  
**DDL**（Data Define Languge）：数据定义语言, 用于**库和表**的创建、修改、删除

## 数据库操作

1. 创建数据库
    ``` sql
    -- 创建数据库
    CREATE DATABASE 【databaseName】;
    -- 判断不存在, 再创建
    CREATE DATABASE IF NOT EXISTS 【databaseName】;
    -- 指定字符集
    CREATE DATABASE 【databaseName】 CHARACTER SET 【charSetName】;
    ```

2. 查询数据库

    ``` sql
    -- 查询所有数据库的名称
    SHOW DATABASES;
    -- 查询某个数据库的创建语句
    SHOW CREATE DATABASE 【databaseName】;
    ```

3. 修改数据库

    ``` sql
    -- 修改数据库的字符集
    ALTER DATABASE 【databaseName】 CHARACTER SET 【charSetName】;
    ```

4. 删除数据库

    ``` sql
    -- 删除数据库
    DROP DATABASE 【databaseName】;
    -- 判断存在, 再删除
    DROP DATABASE IF EXISTS 【databaseName】;
    ```

5. 查询当前正在使用的数据库

    ``` sql
    -- 查询当前正在使用的数据库名称
    SELECT DATABASE();
    ```

6. 使用数据库

    ``` sql
    -- 使用数据库
    USE 【databaseName】;
    ```

## 数据表操作
1. 创建数据表

    ``` sql
    CREATE TABLE 【tableName】(
      【列名1】 【数据类型1】 【字段约束】,
      【列名2】 【数据类型2】 【字段约束】,
      ....
      【列名n】 【数据类型n】 【字段约束】
    );
    ```

2. 修改数据表

    ``` sql
    -- 修改表名
    ALTER TABLE 【oldTableName】 RENAME TO 【newTableName】
    -- 修改表字符集
    ALTER TABLE 【tableName】 CHARACTER SET 【charSetName】;
    -- 修改列名称/类型
    ALTER TABLE 【tableName】 CHANGE 【oldColumnName】 【newColumnName】 【newDataType】;
    alter table 【tableName】 modify 【columnName】 【newDataType】;
    ```

3. 表的删除
    ``` sql
    DROP TABLE 【tableName】;
    -- 判断存在, 再删除
    DROP TABLE IF EXISTS 【tableName】;
    ```

## 约束操作

约束: 对表中的数据进行限定, 保证数据的正确性、有效性和完整性。    

|约束类型|type|说明|
|--|--|--|
|主键约束|primary key|非空且唯一, 一张表只能有一个字段为主键, 主键就是表中记录的唯一标识|
|非空约束|not null|该列的值不能为null|
|唯一约束|unique|该列的值不能重复|
|外键约束|foreign key|表与表的关系|

1. 主键约束 primary key

    ``` sql
    -- 创建表时, 添加主键
    CREATE TABLE 【tableName】(
    id INT PRIMARY KEY, --给id添加主键约束
    -- ...
    );

    -- 创建表后, 添加主键        
    ALTER TABLE 【tableName】 MODIFY 【columnName】 INT PRIMARY KEY;

    -- 删除主键
    ALTER TABLE 【tableName】 DROP PRIMARY KEY;

    -- 创建表时, 添加主键约束, 并完成主键自增长
    CREATE TABLE 【tableName】(
    id INT PRIMARY KEY AUTO_INCREMENT, -- 给id添加主键约束, 并设置为自动增长
    -- ...
    );

    -- 删除自动增长
    ALTER TABLE 【tableName】 MODIFY 【columnName】 INT;

    -- 创建表后, 添加自动增长
    ALTER TABLE 【tableName】 MODIFY 【columnName】 INT AUTO_INCREMENT;
    ```

2. 非空约束 not null

    ``` sql
    -- 创建表时, 添加非空约束
    CREATE TABLE 【tableName】(
      -- ...
      【notNullColumn】 【dataType】 NOT NULL,
      -- ...
    );

    -- 创建表后, 添加非空约束
    ALTER TABLE 【tableName】 MODIFY 【columnName】 【dataType】 NOT NULL;

    -- 删除非空约束
    ALTER TABLE 【tableName】 MODIFY 【columnName】 【dataType】;
    ```

3. 唯一约束 unique

    ``` sql
    -- 创建表时，添加唯一约束
    CREATE TABLE 【tableName】(
      -- ...
      【uniqueColumn】 【dataType】 UNIQUE,
      -- ...
    );

    -- 删除唯一约束
    ALTER TABLE 【tableName】 DROP INDEX 【columnName】;

    -- 创建表后，添加唯一约束
    ALTER TABLE 【tableName】 MODIFY 【uniqueColumn】 【dataType】 UNIQUE;
    ```

4. 外键约束 foreign key

    ``` sql
    -- 创建表时，可以添加外键
    CREATE TABLE 【tableName】(
    ....
    外键列
    CONSTRAINT 【foreignKeyName】 FOREIGN KEY (【foreignKeyColumn】) REFERENCES 主表名称(主表列名称)
    );

    -- 删除外键
    ALTER TABLE 【tableName】 DROP FOREIGN KEY 外键名称;

    -- 创建表后，添加外键
    ALTER TABLE 【tableName】 ADD CONSTRAINT 【foreignKeyName】 FOREIGN KEY (【foreignKeyColumn】) REFERENCES 主表名称(主表列名称); 
    ```

# DML

**DML** (Data Manipulate Language)：数据操纵语言, 用于**添加、删除、修改数
    据库记录**, 并检查数据完整性

1. 添加数据
    ```sql
    -- 添加单行数据：
    -- 方式1：
    INSERT INTO 【tableName】(【columnName1】,【columnName2】,...【columnNameN】) VALUES(【value1】,【value2】,...【valueN】);
    -- 方式2：
    INSERT INTO 【tableName】 SET 【columnName1】=【value1】,【columnName2】=【value2】,...【columnNameN】=【valueN】;

    -- 添加多行数据：
    INSERT INTO 【tableName】(columnName1】,【columnName2】,...【columnNameN】)
    VALUES(【value1】,【value2】,...【valueN】)，
    ...........
    VALUES(【value1】,【value2】,...【valueN】);
    ```

2. 删除数据
    ``` sql
    -- 删除单行数据：
    DELETE FROM 【tableName】 WHERE 【condition】

    -- 删除表中全部数据：
    DELETE FROM 【tableName】
    TRUNCATE TABLE 表名
    ```
DELETE与TRUNCATE区别？
* 1.truncate不能加where条件，而delete可以加where条件
* 2.truncate的效率高
* 3.truncate 删除带自增长的列的表后，如果再插入数据，数据从1开始
* 4.delete 删除带自增长列的表后，如果再插入数据，数据从上一次的断点处开始
* 5.truncate删除不能回滚，delete删除可以回滚


3. 修改数据
    ``` sql
    -- 修改单表
    UPDATE 【tableName】 SET 【columnName1】 = 【value1】, 【columnName2】 = 【value2】,... WHERE 【condition】;

    -- 修改多表
    UPDATE 【tableName1】 别名1,【tableName2】 别名2 SET 【columnName1】=【value1】，【columnName2】=【value2】 WHERE 【condition】
    ```

# DQL

**DQL**（Data Query Language）：数据查询语言, 用来**查询数据库中表的记录(数据)**

    ``` sql
    -- 查询表记录
    SELECT * FROM 【tableName】 WHERE 【condition】;

    -- 查询表中多个字段
    SELECT 【columnName1】,【columnName2】,...【columnNameN】 FROM 【tableName】【condition】;

    -- 去除表中重复值
    SELECT DISTINCT(【columnName】) FROM 【tableName】;
    SELECT DISTINCT 【columnName】 FROM 【tableName】;
    ```