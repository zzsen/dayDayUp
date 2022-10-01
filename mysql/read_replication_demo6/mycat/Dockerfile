# 设置镜像的 base 镜像，这里我们使用 centos 系统镜像
FROM docker.io/centos
# jdk 文件
COPY jdk-8u341-linux-x64.tar.gz /usr/local/
RUN tar -zxvf /usr/local/jdk-8u341-linux-x64.tar.gz -C /usr/local/
ENV JAVA_HOME=/usr/local/jdk1.8.0_341
ENV PATH=$PATH:$JAVA_HOME/bin
ENV CLASSPATH=.:$JAVA_HOME/lib/dt.jar:$JAVA_HOME/lib/tools.jar

# mycat 文件
COPY Mycat-server-1.6.7.4-release-20200105164103-linux.tar.gz /usr/local/
# 设置环境变量
RUN tar -zxvf /usr/local/Mycat-server-1.6.7.4-release-20200105164103-linux.tar.gz -C /usr/local/
ENV MYCAT_HOME=/usr/local/mycat
# 执行最终命令，启动 mycat
RUN source /etc/profile
RUN source ~/.bash_profile
CMD ["/usr/local/mycat/bin/mycat", "console"]
