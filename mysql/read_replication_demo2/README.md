### 相关依赖或者知识点
1. [mysql简单主从部署demo](https://github.com/zzsen/dayDayUp/tree/master/mysql/read_replication_demo1)
2. docker-compose相关基础

### 相关文档链接
1. [mysql简单主从部署demo](https://github.com/zzsen/dayDayUp/tree/master/mysql/read_replication_demo1)
2. [mysql主从简单部署](https://blog.csdn.net/zzsan/article/details/117304644)

### 如何运行
1. 启动容器
    >检出代码后, 进入到项目目录, 执行`docker-compose up -d`启动docker容器
2. 主从简单部署
    >[mysql主从简单部署](https://blog.csdn.net/zzsan/article/details/117304644)

### 常见问题
#### docker网桥和其他网络冲突
每次把docker-compose down了再up, 都会重新创建网桥, 默认自动创建
可通过指定networks配置, 固定每次创建的网桥
```
networks:
  read_replication_demo2:
    driver: bridge
    ipam:
      config:
        - subnet: 172.177.0.0/16
```