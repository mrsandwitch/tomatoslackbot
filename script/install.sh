#!/bin/bash -xe

systemdConfigDir="/etc/systemd/system"
multiUserTargetWant="multi-user.target.wants"
botServiceName="tomatobot.service"

cp "./conf/${botServiceName}" ${systemdConfigDir}

systemctl enable tomatobot
systemctl daemon-reload
