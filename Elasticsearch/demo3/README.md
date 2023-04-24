# 同步 mysql 数据到 Elasticsearch 集群

1. 使用 `docker-compose` 构建 `Elasticsearch` 集群
2. 启动 mysql, 并初始化表和数据
3. 通过 `logstash` 把 mysql 数据同步至 `Elasticsearch`

## 启动 mysql, 初始化表和数据

将需初始化的 mysql, 放于 docker 容器内目录 `docker-entrypoint-initdb.d`

```yml
# docker-compose.yml
volumes:
  - ./test.sql:/docker-entrypoint-initdb.d/test.sql
```

> demo 中提供了示例的 sql 文件 [test.sql](./test.sql)

## 通过 logstash 把 mysql 数据同步至 Elasticsearch

1. 下载 mysql 的 connector 的 jar 包
   访问[mysql 官网](https://downloads.mysql.com/archives/c-j/), 选择对应版本, `Operating System` 选择 `Platform Independent`, 下载 zip, 解压获取里面的 jar 文件

   > demo 中提供了示例的 jar 文件 [mysql-connector-j-8.0.31.jar](./mysql-connector-j-8.0.31.jar)

2. 将 jar 包放于`logstash/logstash-core/lib/jars/` 目录下
   ```yml
   # docker-compose.yml
   volumes:
     - ./mysql-connector-j-8.0.31.jar:/usr/share/logstash/logstash-core/lib/jars/mysql-connector-j-8.0.31.jar
   ```
3. 编写 logstash 配置文件

   ```yml
   input {
      jdbc {
         # 数据库连接信息
         jdbc_connection_string => "jdbc:mysql://mysql1:3306/test"
         # 数据库连接账号
         jdbc_user => "test"
         # 数据库连接密码
         jdbc_password => "test"
         jdbc_validate_connection => true
         # 指定连接驱动
         jdbc_driver_class => "com.mysql.cj.jdbc.Driver"
         # sql及sql转义
         parameters => { "creatorId" => "1" }
         statement => "SELECT * FROM test_role WHERE creator_id = :creatorId"
         #上面运行结果的保存位置
         # last_run_metadata_path => "/usr/share/logstash/result/jdbc-position.txt"
         #记录最后一次运行的结果
         record_last_run => true
         # 同步定时任务, 了解更多可搜索cron表达式
         schedule => " * * * * * *"
         jdbc_paging_enabled => true
         jdbc_page_size => 50000
      }
   }

   filter {
      mutate {
         # 同步时重命名列
         rename => {
            "creator_id" => "[creator__id]"
         }
      }
   }

   output {
      stdout {
      }

      elasticsearch {
         index => "test"
         # 指定数据表中列名为id的列作为文档的id
         document_id => "%{id}"
         document_type => "role"
         hosts => ["http://es:9200"]
      }
   }
   ```

4. 挂载 logstash 配置文件
   ```yml
   volumes:
     - ./logstash.conf:/usr/share/logstash/pipeline/logstash.conf
   ```

启动后, logstash 会根据配置的 cron 表达式对应循环调度, **如果配置不正确, logstash 尝试导入一次后, 就会自动停止**

# 常见问题解决方案

1. `com.mysql.cj.jdbc.Driver not loaded. Are you sure you've included the correct jdbc driver in :jdbc_driver_library?`
   将 jar 包放于`logstash/logstash-core/lib/jars/` 目录下即可

2. 同步 mysql 数据时, 同步表主键作为 es 数据的 id
   ```yaml
   output {
      stdout {
      }
      elasticsearch {
         index => "test"
         # 指定数据表中列名为id的列作为文档的id
         document_id => "%{id}"
         # 文档类型
         document_type => "role"
         hosts => ["http://es:9200"]
         # hosts => ["http://es:9200", "http://es2:9200"]
      }
   }
   ```
   document_id 指定数据表的字段即可
