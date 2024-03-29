#!/bin/bash
#定义用于同步的用户名
MASTER_SYNC_USER=${MASTER_SYNC_USER:-sync_admin}
echo "MASTER_SYNC_USER $MASTER_SYNC_USER"
#定义用于同步的用户密码
MASTER_SYNC_PASSWORD=${MASTER_SYNC_PASSWORD:-123456}
echo "MASTER_SYNC_PASSWORD $MASTER_SYNC_PASSWORD"
#定义用于登录mysql的用户名
ADMIN_USER=${ADMIN_USER:-root}
echo $ADMIN_USER
#定义用于登录mysql的用户密码
ADMIN_PASSWORD=${ADMIN_PASSWORD:-123456}
echo $ADMIN_PASSWORD
#定义运行登录的host地址
ALLOW_HOST=${ALLOW_HOST:-%}
echo $ALLOW_HOST
#定义创建账号的sql语句
CREATE_USER_SQL="CREATE USER '$MASTER_SYNC_USER'@'$ALLOW_HOST' IDENTIFIED BY '$MASTER_SYNC_PASSWORD';"
echo $CREATE_USER_SQL
#定义赋予同步账号权限的sql,这里设置两个权限，REPLICATION SLAVE，属于从节点副本的权限，REPLICATION CLIENT是副本客户端的权限，可以执行show master status语句
GRANT_PRIVILEGES_SQL="GRANT REPLICATION SLAVE,REPLICATION CLIENT ON *.* TO '$MASTER_SYNC_USER'@'$ALLOW_HOST';"
echo $GRANT_PRIVILEGES_SQL
#定义刷新权限的sql
FLUSH_PRIVILEGES_SQL="FLUSH PRIVILEGES;"
#执行sql
test="mysql -u$ADMIN_USER -p$ADMIN_PASSWORD -e $CREATE_USER_SQL $GRANT_PRIVILEGES_SQL $FLUSH_PRIVILEGES_SQL"
echo $test
mysql -u"$ADMIN_USER" -p"$ADMIN_PASSWORD" -e "$CREATE_USER_SQL $GRANT_PRIVILEGES_SQL $FLUSH_PRIVILEGES_SQL"
echo mysql -u"$ADMIN_USER" -p"$ADMIN_PASSWORD" -e "$CREATE_USER_SQL $GRANT_PRIVILEGES_SQL $FLUSH_PRIVILEGES_SQL"