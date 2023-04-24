# 使用 docker 构建 Elasticsearch 集群

## 环境准备

1. docker 安装, 参考链接: [docker 安装](https://blog.csdn.net/zzsan/article/details/105505692)
2. vm.max_map_count 至少为 262144
   ```bash
   sudo sysctl -w vm.max_map_count=262144
   ```

## 构建网络

docker network create elastic-network

## 创建 es 节点

```bash
docker run --itd --name esn01 -p 9200:9200 -v esdata01:/usr/share/elasticsearch/data --network elastic-network -e "node.name=esn01" -e "cluster.name=liuxg-docker-cluster" -e "cluster.initial_master_nodes=esn01" -e "bootstrap.memory_lock=true" --ulimit memlock=-1:-1 -e ES_JAVA_OPTS="-Xms512m -Xmx512m" docker.elastic.co/elasticsearch/elasticsearch:7.5.0
```

```bash
docker run --itd --name esn02 -p 9201:9200 -v esdata02:/usr/share/elasticsearch/data --network elastic-network -e "node.name=esn02" -e "cluster.name=liuxg-docker-cluster" -e "discovery.seed_hosts=esn01" -e "bootstrap.memory_lock=true" --ulimit memlock=-1:-1 -e ES_JAVA_OPTS="-Xms512m -Xmx512m" docker.elastic.co/elasticsearch/elasticsearch:7.5.0
```

```bash
docker run --itd --name esn03 -p 9202:9200 -v esdata03:/usr/share/elasticsearch/data --network elastic-network -e "node.name=esn03" -e "cluster.name=liuxg-docker-cluster" -e "discovery.seed_hosts=esn01,esn02" -e "bootstrap.memory_lock=true" --ulimit memlock=-1:-1 -e ES_JAVA_OPTS="-Xms512m -Xmx512m" docker.elastic.co/elasticsearch/elasticsearch:7.5.0
```

Elasticsearch 版本信息: [Past Releases](https://www.elastic.co/cn/downloads/past-releases)

### 查看节点信息

http://localhost:9200/\_cat/nodes?v

```bash
ip         heap.percent ram.percent cpu load_1m load_5m load_15m node.role master name
172.23.0.2           29          99   9    2.03    1.19     0.76 dilm      -      esn01
172.23.0.3           20          99   9    2.03    1.19     0.76 dilm      *      esn02
172.23.0.4           35          99   9    2.03    1.19     0.76 dilm      -      esn03
```

## Kibana

```bash
docker run --itd --link esn01:elasticsearch --name kibana --network elastic-network -p 5601:5601 docker.elastic.co/kibana/kibana:7.5.0
```

> localhost:5601
