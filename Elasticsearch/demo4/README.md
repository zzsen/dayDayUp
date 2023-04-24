# 同步 mysql 数据到 Elasticsearch 集群

1. 使用 `docker-compose` 构建 `Elasticsearch` 集群
2. 设置 ES 登录账密

## 设置 ES 登录账密

1. docker-compose 的环境变量添加`- xpack.security.enabled=true` 或 修改 es 的 elasticsearch.yml, 添加`xpack.security.enabled=true`

   kibana 容器的环境变量中的 ELASTICSEARCH_USERNAME 和 ELASTICSEARCH_PASSWORD 设置为待会准备设置的账密, 否则 kibana 连不上 es, 会提示服务不可用

2. es 的 bin 目录下，执行设置账密指令 `./elasticsearch-setup-passwords interactive`

   ```bash
   docker exec -it es01 bash
   bin/elasticsearch-setup-passwords interactive
   # 依次给elastic, apm_system, kibana, logstash_system, beats_system, remote_monitoring_user设置登录密码, 这里设置的密码为es_elastic
   ```

3. 设置密码即可
