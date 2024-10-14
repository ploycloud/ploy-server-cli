# Ploy CLI

Ploy CLI is a powerful tool for managing and deploying your cloud applications.

## Installation

### Prerequisites

- [Docker](https://www.docker.com/get-started)
- [Docker Compose](https://docs.docker.com/compose/install/)

### Option 1: Install Script (Recommended)

To install Ploy CLI, run:

```bash
curl -fsSL https://raw.githubusercontent.com/ploycloud/ploy-server-cli/main/install.sh | bash
```

This script will automatically download and install the latest version of Ploy CLI.

### Option 2: Manual Installation

1. Go to the [releases page](https://github.com/ploycloud/ploy-server-cli/releases) and download the latest version for your
   operating system and architecture.
2. Rename the downloaded file to `ploy`.
3. Make the file executable: `chmod +x ploy`
4. Move the file to a directory in your PATH, e.g., `sudo mv ploy /usr/local/bin/`

## Usage

```bash
ploy [command]
```

Available Commands:

### Global Services

- `ploy services start`: Start global services (MySQL, Redis, Nginx Proxy)
- `ploy services stop`: Stop global services
- `ploy services restart`: Restart global services

### Site Management

- `ploy sites start`: Start all sites
- `ploy sites stop`: Stop all sites
- `ploy sites restart`: Restart all sites

### Individual Site Operations

- `ploy start`: Start the current site
- `ploy stop`: Stop the current site
- `ploy restart`: Restart the current site
- `ploy logs [container]`: View logs from containers
- `ploy exec [container] [command]`: Execute commands inside a container

### WordPress CLI

- `ploy wp [wp-cli commands]`: Execute WP-CLI commands for the current WordPress site

### Deployment and Status

- `ploy deploy`: Deploy a repository to PloyCloud
- `ploy list`: List all deployments
- `ploy status`: Check the status of a deployment

### Miscellaneous

- `ploy version`: Display the current version of Ploy CLI
- `ploy update`: Update Ploy CLI to the latest version

For more information on a specific command, run:

```bash
ploy [command] --help
```

## Examples

1. Start global services:
   ```bash
   ploy global start
   ```

2. Deploy a repository:
   ```bash
   ploy deploy https://github.com/username/repo.git
   ```

3. Start all sites:
   ```bash
   ploy sites start
   ```

4. Execute a WP-CLI command:
   ```bash
   ploy wp plugin list
   ```

5. View logs for a specific container:
   ```bash
   ploy logs nginx
   ```

## Configuration

Ploy CLI uses a configuration file to store your API key and default region. You can set these values by creating
a `~/.ploy/config.yaml` file with the following content:

```yaml
api_key: your-api-key-here
region: us-west-2
```

Alternatively, you can set environment variables:

```bash
export PLOY_API_KEY=your-api-key-here
export PLOY_REGION=us-west-2
```

## Development

To contribute to Ploy CLI development:

1. Clone the repository:
   ```bash
   git clone https://github.com/ploycloud/ploy-server-cli.git
   cd ploy-cli
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Build the project:
   ```bash
   go build -o ploy
   ```

4. Run tests:
   ```bash
   go test ./...
   ```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Support

If you encounter any issues or have questions, please [open an issue](https://github.com/ploycloud/ploy-server-cli/issues) on
GitHub.