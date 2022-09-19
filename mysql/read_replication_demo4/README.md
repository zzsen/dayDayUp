### 前言

前面的 demo, 都是通过 docker-compose 启动多个 mysql, 然后分别连上对应的 mysql 容器执行指令, 完成主从配置. 本 demo 涉及 mysql 主从的内容相对较少, 主要是在之前的基础上做了调优, 如果对 mysql 主从感兴趣的, 可以查看前面的 demo

### 改进

#### 简化 docker-compose

这里启用多个 mysql 容器, 每个容器都需要加上对应的环境变量(MYSQL_ROOT_PASSWORD 等), 这部分作为示例, 都是重复的, 故写进同一个 dockerFile, 减少 docker-compose 内容

#### 固定每个容器的 ip

固定容器 ip, 即可减少通过指令获取容器 ip 信息等操作

1. 配置网桥

```yml
networks:
  read_replication_demo4: # 网桥名
    driver: bridge
    ipam:
      config:
        - subnet: 172.177.0.0/16
```

2. 配置容器 ip

```yml
networks:
  read_replication_demo4: # 同上面的网桥名
    ipv4_address: 172.177.0.3
```

#### 使用 docker-entrypoint-initdb.d 执行脚本

利用 docker 官方提供的初始目录, 该目录下的 sql 和 sh 和 bat 会再容器启动时自动执行

```
read_replication
 ├── mysql
 │     ├── master.sh (主服初始化时执行的脚本)
 │     ├── slave.sh (从服初始化时执行的脚本)
 │     ├── master
 │     │     └── my.cnf (数据库配置文件)
 │     └── slave
 │           └── my.cnf (数据库配置文件)
 ├── Dockerfile (镜像DockerFile)
 ├── clear.sh (删除镜像和容器和数据文件)
 ├── rebuild.sh (执行clear.sh的内容, 另外重新启动和创建docker-compose容器)
 └── docker-compose.yml
```

docker-compose 文件中, 通过配置 volume 挂载脚本至路径`docker-entrypoint-initdb.d`下

```yml
- ./mysql/master.sh:/docker-entrypoint-initdb.d/master.sh #挂载脚本
- ./mysql/slave.sh:/docker-entrypoint-initdb.d/slave.sh #挂载脚本
```
