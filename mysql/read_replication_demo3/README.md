### 知识点

1. docker 容器 ip 查看
2. mysql 互为主从

### 前提

由于没有多台服务器, 所以这里改用 docker-compose 启动多个 mysql 服务, docker 安装指引可见: [debian 安装 docker](https://blog.csdn.net/zzsan/article/details/105505692)

> 我的 linux 服务器是 debian 的, 所以这里以 debian 为例
>
> 另外, 如果需要在宿主机连接 docker 的 mysql 服务, 需要在宿主机安装 mysql 服务, mysql 安装指引可见: [linux 安装 mysql](https://blog.csdn.net/zzsan/article/details/105703982)
> 或者通过`docker exec -it [容器名/容器i] /bin/bash`, 再使用 mysql 客户端连接 mysql

之前已经介绍过[mysql 主从简单部署](https://blog.csdn.net/zzsan/article/details/117304644), 该文章会在上述基础上, 实现双主热备.

### 前期准备

这里由于不像 [mysql 主从简单部署](https://blog.csdn.net/zzsan/article/details/117304644) 中所言, 通过 docker 启动容器, 改为使用 docker-compose 启动, 故需要提前准备好相关的 docker-compose 配置和 mysql 的相关配置

```
read_replication
 ├── mysql
 │     ├── master
 │     │     └── my.cnf (数据库配置文件)
 │     └── slave
 │           └── my.cnf (数据库配置文件)
 ├── clear.sh (删除镜像和容器和数据文件)
 ├── rebuild.sh (执行clear.sh的内容, 另外重新启动和创建docker-compose容器)
 └── docker-compose.yml
```

> 虽然这里是互为主从, 但是为了方便去分, 还是用 master(主服)和 slave(从服)来命名

### 启动并配置互为主从

1. 启动服务
   在 docker-compose 目录执行`docker-compose up -d`, 启动 mysql 服务
2. 查看服务状态
   `docker ps -a`, 查看服务是否正常启动, 如果启动异常, 可通过`docker logs [容器名/容器id]`查看容器日志, 以排查异常所在
3. 查看 master(主服) 和 slave(从服) 的 ip 地址, `docker inspect [容器名/容器id] | grep "IPAddress"`
4. 连接 master(主服), 查看 master(主服)的主服状态`show master status;`, 记住`file`和`position`
   > 这里 master(主服)ip 是 `177.177.3.2`, slave(从服)的 ip 地址是 `177.177.3.3`
5. 连接 slave(从服), 配置 slave(从服)的主从配置的主服

   ```sql
   # 配置从服的主从配置的主服
   change master to master_host='177.177.3.2',master_port=3306,master_user='root',master_password='123456',master_log_file='mysql-bin.000003',master_log_pos=340;
   # 启动该mysql的从服配置
   start slave;
   ```

6. 连接 slave(从服), 查看 slave(从服)的主服状态`show master status;`, 记住`file`和`position`
7. 连接 master(主服), 配置 master(主服)的主从配置的主服
   ```sql
   # 配置从服的主从配置的主服
   change master to master_host='177.177.3.3',master_port=3306,master_user='root',master_password='123456',master_log_file='mysql-bin.000003',master_log_pos=340;
   # 启动该mysql的从服配置
   start slave;
   ```

### 验证

1. 连接 master(主服), 创建 student 表
   ```sql
   CREATE TABLE IF NOT EXISTS `student`(
      `id` INT UNSIGNED AUTO_INCREMENT,
      `name` VARCHAR(100) NOT NULL,
      PRIMARY KEY ( `id` )
   ) ENGINE = InnoDB DEFAULT CHARSET = utf8;
   ```
   查看 slave(从服), 是否已经同步该表
   在 master(主服)中插入数据, 查看 slave(从服)是否已同步该插入的数据
   在 slave(从服)中插入数据, 查看 master(主服)是否已同步该插入的数据
2. 连接 slave(从服), 创建 teacher 表
   ```sql
   CREATE TABLE IF NOT EXISTS `teacher`(
      `id` INT UNSIGNED AUTO_INCREMENT,
      `name` VARCHAR(100) NOT NULL,
      PRIMARY KEY ( `id` )
   ) ENGINE = InnoDB DEFAULT CHARSET = utf8;
   ```
   查看 slave(从服), 是否已经同步该表
   在 master(主服)中插入数据, 查看 slave(从服)是否已同步该插入的数据
   在 slave(从服)中插入数据, 查看 master(主服)是否已同步该插入的数据

上述步骤验证无误后, 证明双主备份已经完成

### 脚本说明

clear.sh 删除数据文件及该 demo 相关容器
rebuild.sh 执行 clear.sh, 同时重新创建 docker-compose 相关容器
