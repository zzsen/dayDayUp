#!/bin/bash
docker-compose down 
echo '\033[36m docker-compose down, done\033[0m'

rm -rf master/data/*
rm -rf slave1/data/*
rm -rf slave2/data/*
echo '\033[36m data rm, done, done\033[0m'

docker ps -a