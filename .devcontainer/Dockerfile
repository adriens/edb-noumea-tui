# Codespace/Devcontainer Dockerfile
# Installs Task CLI globally for all users
FROM mcr.microsoft.com/devcontainers/go:latest


# Install Node.js (LTS) and npm
RUN curl -fsSL https://deb.nodesource.com/setup_lts.x | bash - && \
	apt-get install -y nodejs && \
	npm install -g @go-task/cli

# ...other setup steps as needed...
