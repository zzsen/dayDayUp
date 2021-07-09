### 前提
由于没有多台服务器, 所以这里改用docker-compose启动多个mysql服务, docker安装指引可见: [debian安装docker](https://blog.csdn.net/zzsan/article/details/105505692)
> 我的linux服务器是debian的, 所以这里以debian为例
> 
另外, 需要在宿主机连接docker的mysql服务, 需要在宿主机安装mysql服务, mysql安装指引可见: [linux安装mysql](https://blog.csdn.net/zzsan/article/details/105703982)

这里我的虚拟机的ip为: **192.168.226.140**

之前已经介绍过[mysql主从简单部署](https://blog.csdn.net/zzsan/article/details/117304644), 该文章会在上述基础上, 实现双主热备.

### 前期准备
这里由于不像 [mysql主从简单部署](https://blog.csdn.net/zzsan/article/details/117304644) 中所言, 通过docker启动容器, 改为使用docker-compose启动, 故需要提前准备好相关的docker-compose配置和mysql的相关配置
```
read_replication
 ├── mysql
 │     ├── master
 │     │     └── my.cnf (数据库配置文件)
 │     └── slave
 │           └── my.cnf (数据库配置文件)
 └── docker-compose.yml
```
> 虽然这里是互为主从, 但是为了方便去分, 还是用master(主服)和slave(从服)来命名

### 启动并配置互为主从
1. 启动服务
	在docker-compose目录执行`docker-compose up -d`, 启动mysql服务
2. 查看服务状态
	`docker ps -a`, 查看服务是否正常启动, 如果启动异常, 可通过`docker logs [容器名/容器id]`查看容器日志, 已排查异常所在
3. 连接master(主服), 查看master(主服)的主服状态`show master status;`, 记住`file`和`position`	
4. 连接slave(从服), 配置slave(从服)的主从配置的主服
	```sql
	# 配置从服的主从配置的主服
	change master to master_host='192.168.226.140',master_port=3307,master_user='root',master_password='123456',master_log_file='mysql-bin.000003',master_log_pos=340;
	# 启动该mysql的从服配置
	start slave;
	```
 

5. 连接slave(从服), 查看slave(从服)的主服状态`show master status;`, 记住`file`和`position`	
6. 连接master(主服), 配置master(主服)的主从配置的主服
	```sql
	# 配置从服的主从配置的主服
	change master to master_host='192.168.226.140',master_port=3308,master_user='root',master_password='123456',master_log_file='mysql-bin.000003',master_log_pos=340;
	# 启动该mysql的从服配置
	start slave;
	```
 
### 验证
1. 连接master(主服), 创建student表
	```sql
	CREATE TABLE IF NOT EXISTS `student`(
   	   `id` INT UNSIGNED AUTO_INCREMENT,
   	   `name` VARCHAR(100) NOT NULL,
   	   PRIMARY KEY ( `id` )
	) ENGINE = InnoDB DEFAULT CHARSET = utf8;
	```
	查看slave(从服), 是否已经同步该表
	在master(主服)中插入数据, 查看slave(从服)是否已同步该插入的数据
	在slave(从服)中插入数据, 查看master(主服)是否已同步该插入的数据
2. 连接slave(从服), 创建teacher表
	```sql
	CREATE TABLE IF NOT EXISTS `teacher`(
   	   `id` INT UNSIGNED AUTO_INCREMENT,
   	   `name` VARCHAR(100) NOT NULL,
   	   PRIMARY KEY ( `id` )
	) ENGINE = InnoDB DEFAULT CHARSET = utf8;
	```
	查看slave(从服), 是否已经同步该表
	在master(主服)中插入数据, 查看slave(从服)是否已同步该插入的数据
	在slave(从服)中插入数据, 查看master(主服)是否已同步该插入的数据

上述步骤验证无误后, 证明双主备份已经完成


	