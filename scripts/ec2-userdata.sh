#!/bin/bash

# Function to log messages with levels
log() {
    local level=$1
    shift
    echo "$(date '+%Y-%m-%d %H:%M:%S') - [$level] - $*" | tee -a /var/log/ploy-setup.log
}

log "INFO" "Starting Ploy server setup"

# Update and upgrade the system
log "INFO" "Updating and upgrading the system"
if ! apt-get update && apt-get upgrade -y; then
    log "ERROR" "Failed to update and upgrade the system"
    exit 1
fi

# Install Docker and Docker Compose
log "INFO" "Installing Docker and Docker Compose"
if ! apt-get install -y docker.io docker-compose; then
    log "ERROR" "Failed to install Docker and Docker Compose"
    exit 1
fi

# Start and enable Docker service
log "INFO" "Starting and enabling Docker service"
if ! systemctl start docker && systemctl enable docker; then
    log "ERROR" "Failed to start and enable Docker service"
    exit 1
fi

# Create and configure swap space
log "INFO" "Creating and configuring swap space"
if ! dd if=/dev/zero of=/swapfile bs=1M count=2048 || ! chmod 600 /swapfile || ! mkswap /swapfile || ! swapon /swapfile; then
    log "ERROR" "Failed to create and configure swap space"
    exit 1
fi
echo '/swapfile none swap sw 0 0' | tee -a /etc/fstab
echo 'vm.sappiness=10' | tee -a /etc/sysctl.conf
sysctl -p

# Create and configure ploy user
log "INFO" "Creating and configuring ploy user"
if ! useradd -m -s /bin/bash ploy || ! usermod -aG ubuntu ploy || ! usermod -aG docker ploy; then
    log "ERROR" "Failed to create and configure ploy user"
    exit 1
fi
echo 'ploy ALL=(ALL) NOPASSWD: /usr/bin/docker' | tee -a /etc/sudoers
mkdir -p /home/ploy/sites
chown -R ploy:ploy /home/ploy/sites

# Install and configure unattended-upgrades
log "INFO" "Installing and configuring unattended-upgrades"
if ! apt-get install -y unattended-upgrades; then
    log "ERROR" "Failed to install unattended-upgrades"
    # Continue with the script instead of exiting
else
    log "INFO" "Successfully installed unattended-upgrades"
    
    # Configure unattended-upgrades non-interactively
    if ! echo 'Unattended-Upgrade::Allowed-Origins {
        "${distro_id}:${distro_codename}";
        "${distro_id}:${distro_codename}-security";
    };' | sudo tee /etc/apt/apt.conf.d/50unattended-upgrades > /dev/null; then
        log "ERROR" "Failed to configure unattended-upgrades"
    else
        log "INFO" "Successfully configured unattended-upgrades"
    fi
    
    # Enable unattended-upgrades
    if ! echo 'APT::Periodic::Update-Package-Lists "1";
APT::Periodic::Unattended-Upgrade "1";' | sudo tee /etc/apt/apt.conf.d/20auto-upgrades > /dev/null; then
        log "ERROR" "Failed to enable unattended-upgrades"
    else
        log "INFO" "Successfully enabled unattended-upgrades"
    fi
fi

# Install Ploy Server CLI
log "INFO" "Installing Ploy Server CLI"
if ! curl -fsSL https://raw.githubusercontent.com/ploycloud/ploy-server-cli/main/install.sh | bash; then
    log "ERROR" "Failed to install Ploy Server CLI"
    exit 1
fi

# Install Nginx Proxy using Ploy CLI
log "INFO" "Installing Nginx Proxy using Ploy CLI"
if ! runuser -l ploy -c 'ploy services install nginx-proxy'; then
    log "ERROR" "Failed to install Nginx Proxy using Ploy CLI"
    exit 1
fi

log "INFO" "Ploy server setup completed"
