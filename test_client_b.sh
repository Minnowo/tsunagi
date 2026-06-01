#!/bin/bash

addr1="127.0.0.1:7471"
addr2="127.0.0.1:7472"

go run main.go client fmsg --addr1 "${addr1}" --addr2 "tcp://${addr2}" --msg "hello"


