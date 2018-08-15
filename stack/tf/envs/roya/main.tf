terraform = {
  required_version = ">= 0.9.3"

  backend "s3" {
    bucket = "dev-okro-io"
    key    = "terraform-state/roya-dev-okro-io.tfstate"
    region = "eu-west-1"
  }
}

module "k8s_vpc" {
  source = "../../modules/vpc"
  name   = "k8s.roya.dev.okro.io"
  cidr   = "172.24.0.0/16"
}

output "vpc_id" {
  value = "${module.k8s_vpc.id}"
}

output "vpc_cidr" {
  value = "${module.k8s_vpc.cidr_block}"
}

output "utility_subnets" {
  value = "${formatlist(
    "id=%s az=%s",
    module.k8s_vpc.utility_subnets,
    module.k8s_vpc.availability_zones
  )}"
}

output "private_subnets" {
  value = "${formatlist(
    "id=%s az=%s nat=%s",
    module.k8s_vpc.private_subnets,
    module.k8s_vpc.availability_zones,
    module.k8s_vpc.private_nats
  )}"
}