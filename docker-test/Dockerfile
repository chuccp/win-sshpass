FROM ubuntu:22.04

# Prevent interactive prompts during package installation
ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y \
    openssh-server \
    lrzsz \
    sudo \
    rsync \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Create testuser with password and sudo access
RUN useradd -m -s /bin/bash testuser && \
    echo "testuser:testpass" | chpasswd && \
    usermod -aG sudo testuser && \
    mkdir -p /home/testuser/.ssh && \
    chown testuser:testuser /home/testuser/.ssh && \
    chmod 700 /home/testuser/.ssh

# Also set password for root
RUN echo "root:rootpass" | chpasswd

# Pre-populate a testdata directory for README examples and tests
RUN mkdir -p /tmp/testdata /home/testuser/testdata && \
    echo "hello from testdata" > /tmp/testdata/hello.txt && \
    echo "test file content" > /tmp/testdata/test.txt && \
    dd if=/dev/urandom of=/tmp/testdata/binary.bin bs=1024 count=100 2>/dev/null && \
    chmod -R 777 /tmp/testdata && \
    cp /tmp/testdata/* /home/testuser/testdata/ && \
    chown -R testuser:testuser /home/testuser/testdata

# Create a larger file for resume/transfer testing
RUN dd if=/dev/urandom of=/tmp/testdata/large_file.bin bs=1024 count=1024 2>/dev/null && \
    chmod 777 /tmp/testdata/large_file.bin

# Create test upload directory
RUN mkdir -p /tmp/upload /home/testuser/upload && \
    chmod 777 /tmp/upload && \
    chown testuser:testuser /home/testuser/upload

# Configure SSH: allow password auth + key auth
RUN mkdir -p /var/run/sshd && \
    sed -i 's/#PasswordAuthentication yes/PasswordAuthentication yes/' /etc/ssh/sshd_config && \
    sed -i 's/#PasswordAuthentication no/PasswordAuthentication yes/' /etc/ssh/sshd_config && \
    sed -i 's/PasswordAuthentication no/PasswordAuthentication yes/' /etc/ssh/sshd_config && \
    sed -i 's/#PubkeyAuthentication yes/PubkeyAuthentication yes/' /etc/ssh/sshd_config && \
    sed -i 's/#UsePAM yes/UsePAM yes/' /etc/ssh/sshd_config && \
    sed -i 's/UsePAM no/UsePAM yes/' /etc/ssh/sshd_config && \
    sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config && \
    sed -i 's/PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config && \
    echo "AllowUsers testuser root" >> /etc/ssh/sshd_config

# Pre-deploy test key for key-based authentication tests
# (private key: docker-test/test_key, generated with `sshpass keygen -out docker-test/test_key`)
RUN echo 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIIARCAtaU97qC0uKlKSmstudohxT27CFyI7LFgduNsXJ docker-test-key' \
        >> /home/testuser/.ssh/authorized_keys && \
    chown testuser:testuser /home/testuser/.ssh/authorized_keys && \
    chmod 600 /home/testuser/.ssh/authorized_keys && \
    mkdir -p /root/.ssh && chmod 700 /root/.ssh && \
    echo 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIIARCAtaU97qC0uKlKSmstudohxT27CFyI7LFgduNsXJ docker-test-key' \
        >> /root/.ssh/authorized_keys && \
    chmod 600 /root/.ssh/authorized_keys

# Generate host keys
RUN ssh-keygen -A

# Allow sshd privilege-separation user to read /etc/shadow (required for PAM password auth)
RUN usermod -aG shadow sshd

EXPOSE 22
CMD ["/usr/sbin/sshd", "-D", "-o", "LogLevel=DEBUG"]
