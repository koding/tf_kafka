resource "template_file" "kafka_cloud_init_file" {
  template = "${file("cloud_init/kafka.yaml")}"

  vars = {
    region           = "${var.region}"
  }
}
