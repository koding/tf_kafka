variable "ssh_public_key" {
  description = "Public key"
  default     = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC8u3tdgzNBq51ZNK0zXW1FziMU90drNgvY8uLi/zNOL1QuBwbRMNNGj/1ZyZmY+hV3VdmexA9AxsOofWEyvzUtL/hkJCmYglWGnTtIawOyDqTXi8Wjz4d00WW69zOiQqpAIAah5ejVsq9gpHslBy4amU+ExcxYoMYoz3ozccim++HkovLr9EhctfJuWvoPtrqljg4D9bn10eR0gdKNROxpnHPfX/Ge7NGcYAsvod5GsUI5zOV3lGfqJTKs+N1jxuqPVUKhoDiEimUQ4SoxBDneETdhRCZRVIQV7cwTfgw+kF58DqgTJCbwzyTyl9n7827Qi1Ha38oWhkAK+cB3uUgT cihangir@koding.com"
}

resource "aws_key_pair" "cihangir" {
  key_name   = "cihangir"
  public_key = "${var.ssh_public_key}"

  lifecycle {
    create_before_destroy = true
  }
}
