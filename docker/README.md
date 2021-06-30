### 常用指令

1. 查看docker容器日志

   `docker logs [容器名/容器id]`

2. 查看容器信息

   `docker inspect [容器名/容器id]`

3. 进入容器执行指令

   `docker exec [容器名/容器id] /bin/bash`



### 常见问题

#### docker-compose创建的自定义网络, 和其他网段冲突, 影响其他网段正常访问

##### 方案1(不推荐)

1. 把容器down掉, `docker-compose down`
2. 重新启动, `docker-compose up -d`

不推荐原因: 每次down了重新up, 创建新的自定义网络, 依然存在网络冲突的风险



##### 方案2(推荐)

在docker-compose.yml配置文件中明确的指定subnet和gateway

推荐原因: 每次重启或者down了再up, subnet和gateway不会变化, 需要调整时, 也方便

