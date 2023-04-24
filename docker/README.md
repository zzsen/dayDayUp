### 常用指令

1. 查看 docker 容器日志 `docker logs [容器名/容器id]`

2. 查看容器信息 `docker inspect [容器名/容器id]`

3. 进入容器执行指令 `docker exec [容器名/容器id] /bin/bash`

4. 查看容器状态 `docker ps -a`

5. 查看镜像列表 `docker image ls`

6. 移除镜像 `docker image rm [镜像名/镜像id]`

7. 查看容器列表 `docker container ls -a`

更多内容可见[官方文档](https://docs.docker.com/engine/reference/commandline/docker/)

### 常见问题

#### docker 使用空间

1. 查看 docker 硬盘使用情况

   ```shell
   docker system df
   ```

   ```bash
   TYPE            TOTAL     ACTIVE    SIZE      RECLAIMABLE
   Images          7         7         3.762GB   444MB (11%)
   Containers      11        8         2.079GB   2.055GB (98%)
   Local Volumes   18        3         2.56GB    2.56GB (99%)
   Build Cache     0         0         0B        0B
   ```

2. 删除所有未使用过的镜像

   ```bash
   docker image prune -a
   ```

3. 删除所有停止的容器

   ```bash
   docker container prune
   docker rm -f $(docker ps -aq)
   ```

4. 删除未使用的数据卷

   ```bash
   docker volume prune
   ```

5. 删除没有使用过的网络
   ```bash
   docker network prune
   ```
6. 删除所有未使用过的资源

   ```bash
   docker system prune
   ```

   更多可见[官方文档](https://docs.docker.com/config/pruning/)

#### /var/lib/docker/volumes 目录报错

清理 docker 使用空间时, 多手把挂载的数据卷`/var/lib/docker/volumes`也给删了, 启动容器时, 报错信息如下:

```bash
docker: Error response from daemon: open /var/lib/docker/volumes/esdata01/_data: no such file or directory.
```

解决方案:
重启 docker 服务 `systemctl restart docker`
