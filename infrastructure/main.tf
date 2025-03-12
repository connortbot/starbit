provider "aws" {
  region = "us-east-2"  # or your preferred region
  profile = "personal"
}

# Security group for the game server
resource "aws_security_group" "game" {
  name        = "starbit-game"
  description = "Security group for Starbit game server"

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 50051
    to_port     = 50051
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 50052
    to_port     = 50052
    protocol    = "udp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# EC2 instance
resource "aws_instance" "game" {
  ami           = "ami-04f167a56786e4b09"  # Ubuntu 24.04 LTS (Noble Numbat) in us-east-2
  instance_type = "t2.micro"

  vpc_security_group_ids = [aws_security_group.game.id]
  key_name              = var.key_name

  user_data = <<-EOF
              #!/bin/bash
              apt-get update
              apt-get install -y git
              EOF

  tags = {
    Name = "starbit-server"
  }
}

# Variables
variable "key_name" {
  description = "Name of your SSH key pair in AWS"
  type        = string
}

# Output the public IP
output "public_ip" {
  value = aws_instance.game.public_ip
}