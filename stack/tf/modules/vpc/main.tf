variable "cidr" {
  description = "The CIDR block for the VPC"
}

variable "name" {
  description = "Name tag"
}

locals {
  availability_zones = [
    "eu-west-1a",
    "eu-west-1b",
    "eu-west-1c"
  ]
  private_subnets   = [
    "${cidrsubnet(var.cidr, 3, 1)}",
    "${cidrsubnet(var.cidr, 3, 2)}",
    "${cidrsubnet(var.cidr, 3, 3)}"
  ]
  utility_subnets    = [
    "${cidrsubnet(var.cidr, 6, 0)}",
    "${cidrsubnet(var.cidr, 6, 1)}",
    "${cidrsubnet(var.cidr, 6, 2)}"
  ]
}

/**
 * AWS
 */

provider "aws" {
  region = "eu-west-1"
}

/**
 * VPC
 */

resource "aws_vpc" "main" {
  cidr_block           = "${var.cidr}"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags {
    Name = "${var.name}"
  }
}

/**
 * DHCP
 */
resource "aws_vpc_dhcp_options" "dns_resolver" {
  domain_name         = "eu-west-1.compute.internal"
  domain_name_servers = [
    "AmazonProvidedDNS"]

  tags {
    Name = "${var.name}"
  }
}

resource "aws_vpc_dhcp_options_association" "dns_resolver" {
  vpc_id          = "${aws_vpc.main.id}"
  dhcp_options_id = "${aws_vpc_dhcp_options.dns_resolver.id}"
}

/**
 * Gateways
 */

resource "aws_internet_gateway" "main" {
  vpc_id = "${aws_vpc.main.id}"

  tags {
    Name = "${var.name}"
  }
}

resource "aws_nat_gateway" "main" {
  count         = "${length(local.private_subnets)}"
  allocation_id = "${element(aws_eip.nat.*.id, count.index)}"
  subnet_id     = "${element(aws_subnet.utility.*.id, count.index)}"
  depends_on    = [
    "aws_internet_gateway.main"]

  tags {
    Name = "${element(local.availability_zones, count.index)}.${var.name}"
  }
}

resource "aws_eip" "nat" {
  count = "${length(local.private_subnets)}"
  vpc   = true

  tags {
    Name = "${element(local.availability_zones, count.index)}.${var.name}"
  }
}

/**
 * Subnets.
 */

resource "aws_subnet" "private" {
  vpc_id            = "${aws_vpc.main.id}"
  cidr_block        = "${element(local.private_subnets, count.index)}"
  availability_zone = "${element(local.availability_zones, count.index)}"
  count             = "${length(local.private_subnets)}"

  tags {
    Name                              = "${element(local.availability_zones, count.index)}.${var.name}"
    SubnetType                        = "Private"
    "kubernetes.io/role/private-elb" = "1"
  }
}

resource "aws_subnet" "utility" {
  vpc_id            = "${aws_vpc.main.id}"
  cidr_block        = "${element(local.utility_subnets, count.index)}"
  availability_zone = "${element(local.availability_zones, count.index)}"
  count             = "${length(local.utility_subnets)}"

  tags {
    Name                     = "utility-${element(local.availability_zones, count.index)}.${var.name}"
    SubnetType               = "Utility"
    "kubernetes.io/role/elb" = "1"
  }
}

/**
 * Route tables
 */

resource "aws_route_table" "utility" {
  vpc_id = "${aws_vpc.main.id}"

  tags {
    Name                      = "${var.name}"
    "kubernetes.io/kops/role" = "public"
  }
}

resource "aws_route" "utility" {
  route_table_id         = "${aws_route_table.utility.id}"
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = "${aws_internet_gateway.main.id}"
}

resource "aws_route_table" "private" {
  count  = "${length(local.private_subnets)}"
  vpc_id = "${aws_vpc.main.id}"

  tags {
    Name                      = "private-${element(local.availability_zones, count.index)}.${var.name}"
    "kubernetes.io/kops/role" = "private-${element(local.availability_zones, count.index)}"
  }
}

resource "aws_route" "private" {
  # Create this only if using the NAT gateway service, vs. NAT instances.
  count                  = "${length(compact(local.private_subnets))}"
  route_table_id         = "${element(aws_route_table.private.*.id, count.index)}"
  destination_cidr_block = "0.0.0.0/0"
  nat_gateway_id         = "${element(aws_nat_gateway.main.*.id, count.index)}"
}

/**
 * Route associations
 */

resource "aws_route_table_association" "private" {
  count          = "${length(local.private_subnets)}"
  subnet_id      = "${element(aws_subnet.private.*.id, count.index)}"
  route_table_id = "${element(aws_route_table.private.*.id, count.index)}"
}

resource "aws_route_table_association" "utility" {
  count          = "${length(local.utility_subnets)}"
  subnet_id      = "${element(aws_subnet.utility.*.id, count.index)}"
  route_table_id = "${aws_route_table.utility.id}"
}

/**
 * Outputs
 */

output "id" {
  value = "${aws_vpc.main.id}"
}

output "cidr_block" {
  value = "${aws_vpc.main.cidr_block}"
}

output "availability_zones" {
  value = [
    "${aws_subnet.utility.*.availability_zone}"]
}

output "utility_subnets" {
  value = [
    "${aws_subnet.utility.*.id}"]
}

output "private_subnets" {
  value = [
    "${aws_subnet.private.*.id}"]
}

output "private_nats" {
  value = [
    "${aws_nat_gateway.main.*.id}"]
}