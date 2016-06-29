resource "template_file" "kafka_cloud_init_file" {
  template = "${file("data/kafka.yaml.tpl")}"

  vars = {
    kafka_install_script_b64      = "${base64encode(template_file.kafka_install.rendered)}"
    kafka_init_script_b64         = "${base64encode(file("data/kafka.init"))}"
    kafka_broker_id_generator_b64 = "${base64encode(file("data/brokerid.sh"))}"
    kafka_server_properties_b64   = "${base64encode(file("data/server.properties"))}"
    region                        = "${var.region}"
  }
}

resource "template_file" "kafka_install" {
  template = "${file("data/kafka_install.sh.tpl")}"

  vars = {
    SCALA_VERSION      = "${var.scala_version}"
    KAFKA_VERSION      = "${var.kafka_version}"
    ZOOKEEPER_CONN_STR = "${join(",",formatlist("%s:%s", split(",", var.zookeeper_ips), var.zookeeper_port))}"
  }
}
