#!/bin/bash
redis-cli -p 6380 DEL bigintset
for i in {1..600}; do
  redis-cli -p 6380 SADD bigintset $i > /dev/null
done
redis-cli -p 6380 SCARD bigintset
redis-cli -p 6380 SETTYPE bigintset