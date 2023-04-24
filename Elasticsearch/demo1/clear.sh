#!/bin/bash

rm ./esdata01*
rm ./esdata02*
rm ./esdata03*

docker rm -f esn01
docker rm -f esn02
docker rm -f esn03