### 前言

前面几个 demo, 实现了 mysql 的主从备份, 双主备份等, 这里开始研究 mysql 高可用和稳定性的问题

### 思路

基于之前的主从备份的 demo, 外加 nginx 的 upstream 的 backup 实现代理分发

```sql
# 添加stream模块，实现tcp反向代理
stream {
    proxy_timeout 30m;
    upstream mysql-slave-cluster{
      #docker-compose.yml里面会配置固定mysql-slave的ip地址，这里就填写固定的ip地址
      server 177.177.5.2:3306 weight=1;
      server 177.177.5.3:3306 weight=1 backup; #备用数据库，当上面的数据库挂掉之后，才会使用此数据库，也就是如果上面的数据库没有挂，则所有的流量都很转发到上面的主库
    }
    server {
      listen  0.0.0.0:3307;
      proxy_pass mysql-slave-cluster;
    }
}
```

### 文件路径和说明

```sh
.
├── clear.sh              # 删除数据文件及该 demo 相关容器
├── docker-compose.yml
├── mysql
│   ├── master.sh         # 主服创建同步用户的脚本
│   └── slave.sh          # 修改从服的主服脚本
├── nginx                 # nginx配置
│   └── nginx.conf
├── README.md
└── rebuild.sh            # 执行 clear.sh, 同时重新创建 docker-compose 相关容器
```
