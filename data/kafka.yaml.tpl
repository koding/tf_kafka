#cloud-config

repo_update: true
repo_upgrade: all

write_files:
  - path: /tmp/install_kafka.sh
    permissions: "0755"
    encoding: b64
    content: |
      ${kafka_install_script_b64}

  - path: /etc/init.d/kafka
    permissions: "0755"
    encoding: b64
    content: |
      ${kafka_init_script_b64}

runcmd:
 - /tmp/install_kafka.sh
 - chkconfig --add kafka
 - service kafka start

users:
  - name: kafka
    primary-group: kafka
    lock_passwd: true
    system: true
