
FROM mysql:8.0.29

ENV MYSQL_DATABASE test
# 创建test用户
ENV  MYSQL_USER sen
# 设置test用户的密码
ENV  MYSQL_PASSWORD zzsen
ENV  MYSQL_ROOT_PASSWORD 123456

#设置脚本中定义的用于同步的账号
ENV  MASTER_SYNC_USER sync_admin
#允许同步账号的host地址
ENV  ALLOW_HOST 177.177.4.%
# 设置时区
ENV  TZ Asia/Shanghai

EXPOSE 3306