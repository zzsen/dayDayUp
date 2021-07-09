
FROM mysql

ENV MYSQL_DATABASE test
# 创建test用户
ENV  MYSQL_USER sen
# 设置test用户的密码
ENV  MYSQL_PASSWORD zzsen
ENV  MYSQL_ROOT_PASSWORD 123456
# 设置时区
ENV  TZ Asia/Shanghai


VOLUME ./my.cnf /etc/mysql/my.cnf
VOLUME ./db/ /var/lib/mysql
VOLUME ./db-files/ /var/lib/mysql-files

EXPOSE 3306