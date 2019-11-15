#!/bin/sh

set -ex

while test ! -e "${NEBTMP}/postgres_init_done"
do
    sleep 1
done
sleep 1
