#!/bin/bash

KAFKA_RELEASE=kafka_${SCALA_VERSION}-${KAFKA_VERSION}
KAFKA_HOME=/opt/kafka
KAFKA_USER=kafka

# install java
JAVA_RPM=jre-8u92-linux-x64.rpm
wget --no-cookies --no-check-certificate \
--header "Cookie: gpw_e24=http%3A%2F%2Fwww.oracle.com%2F; oraclelicense=accept-securebackup-cookie" \
"http://download.oracle.com/otn-pub/java/jdk/8u92-b14/$$JAVA_RPM"

sudo yum localinstall -y $$JAVA_RPM
rm -rf $$JAVA_RPM

# download kafka
curl -sS https://archive.apache.org/dist/kafka//${KAFKA_VERSION}/$KAFKA_RELEASE.tgz -o /opt/kafka.tgz
cd /opt && tar zxf kafka.tgz && rm kafka.tgz
mv /opt/kafka_* $$KAFKA_HOME

# set permissions for kafka
mkdir /var/{lib,log}/kafka
chown -R $$KAFKA_USER:$$KAFKA_USER /var/{lib,log}/kafka
