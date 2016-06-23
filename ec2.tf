resource "aws_instance" "kafka" {
  count         = "${length(split(",", var.private_ips))}"
  ami           = "${var.ami_id}"
  instance_type = "${var.aws_instance_type}"
  key_name      = "${aws_key_pair.cihangir.key_name}"

  ebs_optimized = "${var.ebs_optimized}"
  subnet_id = "${var.aws_subnet_subnet_id}"

  # (Optional) A list of security group IDs to associate with.
  vpc_security_group_ids = ["${split(",", var.vpc_security_group_ids)}"]

  user_data = "${element(template_file.kafka_cloud_init_file.*.rendered, count.index)}"

  private_ip = "${element(split(",", var.private_ips), count.index)}"

  tags {
    Name = "${format("${var.name}-%03d", count.index + 1)}"
  }
}
