#!/bin/bash


scp config.yaml $1:subserver/
ssh $1 docker restart subserver

