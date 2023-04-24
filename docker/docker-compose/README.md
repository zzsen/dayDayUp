### 安装

#### 下载

官方文档: https://docs.docker.com/compose/install/

1. github 安装

   ```bash
   curl -SL https://github.com/docker/compose/releases/download/v2.14.0/docker-compose-linux-x86_64 -o /usr/local/bin/docker-compose
   ```

   > 可访问[官网](https://docs.docker.com/compose/install/other/)查看最新版安装指令

2. DaoCloud 安装
   方式一安装可能不稳定, 可通过 DaoCloud 安装
   ```bash
   curl -L https://get.daocloud.io/docker/compose/releases/download/v2.14.0/docker-compose-`uname -s`-`uname -m` > /usr/local/bin/docker-compose
   ```
   > 可访问[DaoCloud]查看版本

#### 授权

```bash
sudo chmod +x /usr/local/bin/docker-compose
```

#### 查看版本

```bash
docker-compose --version
```

### docker-compose.yml 文件详解

#### demo

举个简单例子, 这里使用 docker-compose 启动个 mysql

```yml
# 描述 Compose 文件的版本信息
version: "3.8"
# 定义服务，可以多个
services:
  mysql_service: # 服务名称
    command: # 覆盖容器启动后默认执行的命令
      - --default-authentication-plugin=mysql_native_password
      - --character-set-server=utf8mb4
      - --collation-server=utf8mb4_unicode_ci
    environment:
      - MYSQL_DATABASE=test
      - MYSQL_USER=test # 创建test用户
      - MYSQL_PASSWORD=test # 设置test用户的密码
      - MYSQL_ROOT_PASSWORD=123456
      - TZ=Asia/Shanghai # 设置时区
    image: mysql:8.0.30 # 创建容器时所需的镜像
    container_name: mysql_container # 容器名称，默认为"工程名称_服务条目名称_序号"
    ports: # 宿主机与容器的端口映射关系
      - 3307:3306 # 左边宿主机端口:右边容器端口
    restart: always # 容器总是重新启动
    networks: # 配置容器连接的网络，引用顶级 networks 下的条目
      - mysql-net

# 定义网络，可以多个。如果不声明，默认会创建一个网络名称为"工程名称_default"的 bridge 网络
networks:
  mysql-net: # 一个具体网络的条目名称
    name: mysql-net # 网络名称，默认为"工程名称_网络条目名称"
    driver: bridge # 网络模式，默认为 bridge
```

```bash
docker-compose up -d # 后台启动
```

#### version

compose 文件的版本信息, 对照表可见[官方文档](https://docs.docker.com/compose/compose-file/compose-versioning/)

#### services

用于定义服务, 可以多个, 然后在每个服务中, 声明所需的镜像、环境变量、端口参数等, 和使用`docker run`一样
demo 中的内容, 使用`docker run`的指令如下:

```bash
docker run -itd --name mysql_service \
--network elastic-network \
-p 3307:3306 -e MYSQL_DATABASE=test -e MYSQL_USER=test -e MYSQL_PASSWORD=test -e MYSQL_ROOT_PASSWORD=test -e TZ=Asia/Shanghai \
mysql:8.0.30 \
--default-authentication-plugin=mysql_native_password --character-set-server=utf8mb4 --collation-server=utf8mb4_unicode_ci
```

#### image

镜像的标签或 id, 优先使用本地镜像, 本地不存在, 则去远程拉取(默认: https://hub.docker.com/)

#### build

并非所有镜像都是拿来就可以直接用的, 部分镜像可基于 Dockerfile 构建, 这里用 golang 微服务举个例子, Dockerfile 内容如下:

```
# 镜像
FROM golang:1.19
# 拷贝文件至目录
COPY . /src
# 指定工作目录
WORKDIR /src
# 安装依赖包
RUN go mod tidy
# 打包文件
RUN go build -o goBuild
# 镜像启动入口
CMD ["/src/goBuild"]
```

```yml
# 描述 Compose 文件的版本信息
version: "3.8"
# 定义服务
services:
  go_service: # 服务名称
    build: . # 相对当前 docker-compose.yml 文件所在目录，基于名称为 Dockerfile 的文件构建镜像
    container_name: go_service # 容器名称，默认为"工程名称_服务条目名称_序号"
    ports: # 宿主机与容器的端口映射关系
      - "8080:8080" # 左边宿主机端口:右边容器端口
```

更多关于 build 的内容, 可见[官方文档](https://docs.docker.com/compose/compose-file/build/)

#### depends_on

指定依赖容器, 会在依赖容器启动后再启动, 具体使用场景如: 如果未启用 redis 和 db 服务, 启动 web 服务会报错而退出

```yml
services:
  web:
    build: .
    depends_on:
      - db
      - redis
  redis:
    image: redis
  db:
    image: mysql
```

#### ports

容器对外暴露的端口，格式：左边宿主机端口:右边容器端口。

```yml
ports:
  - "80:80"
  - "8080:8080"
```

#### expose

容器暴露的端口不映射到宿主机，只允许能被连接的服务访问。

```yml
expose:
  - "80"
  - "8080"
```

#### restart

容器重启策略，简单的理解就是 Docker 重启以后容器要不要一起启动：

`no`：默认的重启策略，在任何情况下都不会重启容器；
`always`：容器总是重新启动，除非容器被移除了；
`on-failure`：容器非正常退出时，才会重启容器；
`unless-stopped`：容器总是重新启动，除非容器被停止（手动或其他方式），那么 Docker 重启时容器则不会启动。

```yml
services:
  mysql:
    image: mysql:8.0.30
    container_name: mysql
    ports:
      - "3306:3306"
    restart: always
```

#### environment

环境变量。可以使用数组也可以使用字典。bool 值（true、false、yes、no）需要用引号括起来，以确保 YML 解析器不会将它们转换为 `True` 或 `False`。

```yml
environment:
  RACK_ENV: development
  SHOW: "true"
  USER_INPUT:
```

或者以下格式：

```yml
environment:
  - RACK_ENV=development
  - SHOW=true
  - USER_INPUT
```

#### env_file

从文件中获取环境变量, 可以是单个文件, 也可以是多个文件, 可以用绝对路径, 也可以用相对路径

```yml
env_file: .env
```

```yml
env_file:
  - ./a.env
  - ./b.env
```

更多关于 env_file 的内容, 可见[官方文档](https://docs.docker.com/compose/compose-file/#env_file)

#### command

覆盖镜像声明的默认指令

```yml
command: echo "helloworld"
```

也可以是一个列表。

```yml
command: ["echo", "helloworld"]
```

#### volumes

数据卷，用于实现目录挂载，支持指定目录挂载、匿名挂载、具名挂载。

1. `指定目录挂载`<br>
   格式: `左边宿主机目录:右边容器目录`，或 `左边宿主机目录:右边容器目录:读写权限`

2. `匿名挂载`<br>
   格式：`容器目录`，或 `容器目录:读写权限`
3. `具名挂载`<br>
   格式：`数据卷条目名称:容器目录`，或 `数据卷条目名称:容器目录:读写权限`

```yml
# 版本信息
version: "3.8"
# 定义服务
services:
  mysql: # 服务名称
    image: mysql:8.0.30 # 镜像
    container_name: mysql8 # 容器名称
    ports: # 端口映射关系
      - "3306:3306" # 左边宿主机端口:右边容器端口
    environment: # 创建容器时所需的环境变量
      MYSQL_ROOT_PASSWORD: 1234
    volumes:
      # 绝对路径
      - "/mydata/docker_mysql/data:/var/lib/mysql"
      # 相对路径，相对当前 docker-compose.yml 文件所在目录
      - “./conf:/etc/mysql/conf.d“
      # 匿名挂载，匿名挂载只需要写容器目录即可，容器外对应的目录会在 /var/lib/docker/volume 中生成
      - "/var/lib/mysql"
      # 具名挂载，就是给数据卷起了个名字，容器外对应的目录会在 /var/lib/docker/volume 中生成
      - "mysql-data-volume:/var/lib/mysql"

# 定义数据卷，可以多个
volumes:
  mysql-data-volume: # 一个具体数据卷的条目名称
    name: mysql-data-volume # 数据卷名称
```

#### 更多

更多内容可见[官方文档](https://docs.docker.com/compose/compose-file/)

### 常用指令

1. 创建并启动所有服务的容器 `docker-compose up -d`
2. 停止并删除所有服务的容器 `docker-compose down -v`
3. 查看服务容器的输出日志 `docker-compose logs`
4. 列出工程中所有服务的容器 `docker-compose ps`
5. 在指定服务容器上执行命令 `docker-compose run nginx echo "helloworld"`
6. 进入服务容器 `docker-compose exec nginx bash`
7. 启动/重启/暂停/恢复/停止 某个/全部服务 `docker-compose start/restart/pause/unpause/stop [容器名/容器id, 不填则全部容器]`

更多内容可见[官方文档](https://docs.docker.com/compose/reference/)

### 定制版本

在 docker-compose 同级目录下, 建个文件`.env`, 在文件内添加以下内容:

```
MYSQL_VERSION=8.0.29
```

`docker-compose.yml`中, image 可以用`.env`文件中的变量

```yml
version: "3"

services:
  mysql:
    container_name: mysql
    image: mysql:${MYSQL_VERSION}
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
      - ./mysql/master/db/:/var/lib/mysql # 低版本的mysql数据存于此
      - ./mysql/master/db/:/var/lib/mysql-files # 高版本的mysql数据存于此
    restart: always
```

如果需要更换版本, 修改 env 文件即可

### 常见问题

#### docker-compose 创建的自定义网络, 和其他网段冲突, 影响其他网段正常访问

##### 方案 1(不推荐)

1. 把容器 down 掉, `docker-compose down`
2. 重新启动, `docker-compose up -d`

不推荐原因: 每次 down 了重新 up, 创建新的自定义网络, 依然存在网络冲突的风险

##### 方案 2(推荐)

在 docker-compose.yml 配置文件中明确的指定 subnet 和 gateway

推荐原因: 每次重启或者 down 了再 up, subnet 和 gateway 不会变化, 需要调整时, 也方便
