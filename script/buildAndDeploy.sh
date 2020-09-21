#!/bin/bash -xe

go build .
systemctl restart tomatobot
