module "bastion_key_pair" {
  source  = "cloudposse/key-pair/aws"
  version = "0.16.1"

  name                = "${var.namespace}-bastion-key"
  attributes          = ["ssh", "key"]
  ssh_public_key_path = "/tmp/"
  generate_ssh_key    = true

  tags = local.tags
}

resource "aws_security_group" "bastion" {
  name        = "bastion_sg"
  description = "Allow SSH inbound traffic"
  vpc_id      = module.vpc.vpc_id

  ingress {
    description = "SSH access"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = local.tags
}

module "ec2_bastion" {
  source  = "cloudposse/ec2-bastion-server/aws"
  version = "0.25.0"

  name          = "${var.namespace}-bastion"
  key_name      = module.bastion_key_pair.key_name
  instance_type = "t3a.micro"
  user_data_base64 = base64encode(templatefile("templates/bastion_userdata.sh.tpl", {
    capectl_version = var.capectl_version
  }))

  ami_filter = {
    name = ["ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"]
  }
  ami_owners = ["099720109477"]

  vpc_id                      = module.vpc.vpc_id
  subnets                     = module.vpc.public_subnets
  security_groups             = [aws_security_group.bastion.id]
  associate_public_ip_address = true

  tags = local.tags
}
