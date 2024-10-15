#!/bin/bash

# Function to log messages
log() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" | tee -a /var/logs/ploy-setup.log
}

log "Starting Ploy server setup"

# Update and upgrade the system
log "Updating and upgrading the system"
apt-get update && apt-get upgrade -y

# Install Docker and Docker Compose
log "Installing Docker and Docker Compose"
apt-get install -y docker.io docker-compose

# Start and enable Docker service
log "Starting and enabling Docker service"
systemctl start docker
systemctl enable docker

# Create and configure swap space
log "Creating and configuring swap space"
dd if=/dev/zero of=/swapfile bs=1M count=2048
chmod 600 /swapfile
mkswap /swapfile
swapon /swapfile
echo '/swapfile none swap sw 0 0' | tee -a /etc/fstab
echo 'vm.sappiness=10' | tee -a /etc/sysctl.conf
sysctl -p

# Create and configure ploy user
log "Creating and configuring ploy user"
useradd -m -s /bin/bash ploy
usermod -aG ubuntu ploy
usermod -aG docker ploy
echo 'ploy ALL=(ALL) NOPASSWD: /usr/bin/docker' | tee -a /etc/sudoers
mkdir -p /home/ploy/sites
chown -R ploy:ploy /home/ploy/sites

# Install and configure unattended-upgrades
log "Installing and configuring unattended-upgrades"
apt-get install -y unattended-upgrades
dpkg-reconfigure -plow unattended-upgrades

# Install Ploy Server CLI
log "Installing Ploy Server CLI"
curl -fsSL https://raw.githubusercontent.com/cloudoploy/ploy-server-cli/main/install.sh | bash

# Install Nginx Proxy using Ploy CLI
log "Installing Nginx Proxy using Ploy CLI"
sudo -u ploy ploy services install nginx-proxy

log "Ploy server setup completed"
