### 相关依赖或者知识点

1. [mysql 简单主从部署 demo](https://github.com/zzsen/dayDayUp/tree/master/mysql/read_replication_demo1)
2. docker-compose 相关基础
3. 使用 docker 网桥避免网络冲突

### 相关文档链接

1. [mysql 简单主从部署 demo](https://github.com/zzsen/dayDayUp/tree/master/mysql/read_replication_demo1)
2. [mysql 主从简单部署](https://blog.csdn.net/zzsan/article/details/117304644)

### 如何运行

1. 启动容器
   > 检出代码后, 进入到项目目录, 执行`docker-compose up -d`启动 docker 容器
2. 主从简单部署
   > [mysql 主从简单部署](https://blog.csdn.net/zzsan/article/details/117304644)

### 常见问题

#### docker 网桥和其他网络冲突

每次把 docker-compose down 了再 up, 都会重新创建网桥, 默认自动创建
可通过指定 networks 配置, 固定每次创建的网桥

```
networks:
  read_replication_demo2:
    driver: bridge
    ipam:
      config:
        - subnet: 177.177.2.0/166
```

### 脚本说明

clear.sh 删除数据文件及该 demo 相关容器
rebuild.sh 执行 clear.sh, 同时重新创建 docker-compose 相关容器
