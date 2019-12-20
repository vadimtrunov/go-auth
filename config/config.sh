#!/usr/bin/env bash
ENV='local'

while getopts e: option
do
 case "${option}"
 in
 e) ENV=${OPTARG};;
 esac
done

if [ ${ENV} == 'local' ]; then
    FILE='config/docker-compose-local.yml'
elif [ ${ENV} == 'dev' ]; then
    FILE='config/docker-compose-dev.yml'
else
    echo 'Environment not found.'
    exit 1
fi
