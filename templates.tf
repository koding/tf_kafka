resource "template_file" "kafka_cloud_init_file" {
  template = "${file("data/kafka.yaml.tpl")}"

  vars = {
    kafka_install_script_b64    = "${base64encode(template_file.kafka_install.rendered)}"
    kafka_init_script_b64       = "${base64encode(file("data/kafka.init"))}"
    region                      = "${var.region}"
  }
}

resource "template_file" "kafka_install" {
  template = "${file("data/kafka_install.sh.tpl")}"

  vars = {
    SCALA_VERSION = "${var.scala_version}"
    KAFKA_VERSION = "${var.kafka_version}"
  }
}
