### 相关依赖或者知识点
1. docker相关基础
2. docker-compose相关基础
3. mysql简单主从部署

### 相关文档链接
1. [debian安装docker](https://blog.csdn.net/zzsan/article/details/105505692)
2. [mysql主从简单部署](https://blog.csdn.net/zzsan/article/details/117304644)

### 如何运行
#### 启动容器
    >检出代码后, 进入到项目目录, 执行`docker-compose up -d`启动docker容器
#### 主从简单部署
  1. 登录主服, 查看主服master状态 `show master status;`

  2. 登录从服mysql，设置与主服务器相关的配置参数

       ```sql
        # 设置与主服务器相关的配置参数
        change master to master_host='192.168.226.137',master_port=3307,master_user='replicate',master_password='123456',master_log_file='mysql-bin.000003',master_log_pos=340;
        # 启动从服
        start slave;
       ```

       > master_host: docker的地址, 不能写127.0.0.1

       > master_user: 在主库创建的用户, 步骤8中创建的用户
       
       > master_port: 主库的端口, 默认3306

       > master_log_pos: 主库show master status;查询出的Position
  
  详见: [mysql主从简单部署](https://blog.csdn.net/zzsan/article/details/117304644)


### 常见问题
#### 出现同步错误后, 后续同步不执行
若在主从同步的过程中，出现其中一条语句同步失败报错了，则后面的语句也肯定不能同步成功了。例如，主库有一个数据库，而从库并没有，然而，在主库执行了删除这个数据库的操作，那么从库没有这么数据库就肯定删除不了，从而报错了。在此时的从数据库的数据同步就失败了，因此后面的同步语句就无法继续执行。

这里提供的解决方法有两种：

1. 在从数据库中，使用SET全局sql_slave_skip_counter来跳过事件，跳过这一个错误，然后执行从下一个事件组开始。
#在从数据库上操作
	```
	mysql > stop slave;
	mysql > set global sql_slave_skip_counter=1;
	mysql > start slave;
	```

2. 想办法(例如, 数据库迁移等)令从库与主库的数据结构和数据都一致了之后，再来恢复主从同步的操作。
	
	>`start slave;`

#### 重新创建容器后, 新建的容器一直在重启

修改docker镜像版本时, 把容器remove了重建创建, 但是新建的容器却在启动后马上退出了
发现报错: `InnoDB: Table flags are 0 in the data dictionary but the flags in file ./ibdatal are 0x4800!`
原来是我remove的容器在宿主机有生成对应的db文件, 这里不需要保留, 直接删除即可, 删除完在`docker start 容器名`即可

#### secure_file_priv
启动容器后, 报以下错误:
`failed to access directory for --secure-file-priv. please make sure that directory exists and is accessible by MYSQL Server. Supplied value: /var/lib/mysql-files`
这里主要是mysql8.0+的文件路径从原来的`/var/lib/mysql`调整为了`/var/lib/mysql-files`
解决上述报错的方法有:
1. 不调整docker-compose内容, 在代码库设置里, 加上 `secure_file_priv=/var/lib/mysql`
	  ```
    [mysqld]
    log-bin=mysql-bin
    server-id=101
    secure_file_priv=/var/lib/mysql

  	```
2. 修改docker-compose内容, 文件路径映射加上 `- ./db/mysql/master/:/var/lib/mysql-files`
    ```
    mysql_master:
    container_name: read_replication_demo1_mysql_master
    image: mysql
    command:
      - --default-authentication-plugin=mysql_native_password
      - --character-set-server=utf8mb4
      - --collation-server=utf8mb4_unicode_ci
    ports:
      - 3307:3306
    environment:
      - MYSQL_DATABASE=test
      - MYSQL_USER=sen # 创建test用户
      - MYSQL_PASSWORD=zzsen # 设置test用户的密码
      - MYSQL_ROOT_PASSWORD=123456
      - TZ=Asia/Shanghai # 设置时区
    volumes:
      - ./mysql/master/my.cnf:/etc/mysql/my.cnf
      - ./db/mysql/master/:/var/lib/mysql-files # 高版本的mysql数据存于此
    restart: always
    networks:
      - read_replication_demo1
    ```