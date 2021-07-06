
FROM percona/percona-xtradb-cluster

ENV MYSQL_DATABASE test
# 创建test用户
ENV  MYSQL_USER sen
# 设置test用户的密码
ENV  MYSQL_PASSWORD zzsen
ENV  MYSQL_ROOT_PASSWORD 123456
ENV  XTRABACKUP_PASSWORD 123456
# 设置时区
ENV  TZ Asia/Shanghai
# 集群名
ENV CLUSTER_NAME PXC


VOLUME ./my.cnf /etc/mysql/my.cnf
# 低版本的mysql数据存于此
VOLUME ./db/ /var/lib/mysql
# 高版本的mysql数据存于此
VOLUME ./db/ /var/lib/mysql-files

EXPOSE 3306