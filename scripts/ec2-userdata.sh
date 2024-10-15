#!/bin/bash

# Update and upgrade the system
apt-get update && apt-get upgrade -y

# Install Docker and Docker Compose
apt-get install -y docker.io docker-compose

# Start and enable Docker service
systemctl start docker
systemctl enable docker

# Create and configure swap space
dd if=/dev/zero of=/swapfile bs=1M count=2048
chmod 600 /swapfile
mkswap /swapfile
swapon /swapfile
echo '/swapfile none swap sw 0 0' | tee -a /etc/fstab
echo 'vm.sappiness=10' | tee -a /etc/sysctl.conf
sysctl -p

# Create and configure ploy user
useradd -m -s /bin/bash ploy
usermod -aG ubuntu ploy
usermod -aG docker ploy
echo 'ploy ALL=(ALL) NOPASSWD: /usr/bin/docker' | tee -a /etc/sudoers
mkdir -p /home/ploy/sites
chown -R ploy:ploy /home/ploy/sites

# Install and configure UFW (Uncomplicated Firewall)
#apt-get install -y ufw
#ufw default deny incoming
#ufw default allow outgoing
#ufw allow ssh
#ufw allow http
#ufw allow https
#echo "y" | ufw enable

# Install and configure unattended-upgrades
apt-get install -y unattended-upgrades
dpkg-reconfigure -plow unattended-upgrades

# Install Ploy Server CLI
curl -fsSL https://raw.githubusercontent.com/cloudoploy/ploy-server-cli/main/install.sh | bash
