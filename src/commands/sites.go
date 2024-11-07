package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/ploycloud/ploy-server-cli/src/common"
	"github.com/ploycloud/ploy-server-cli/src/docker"
	"github.com/spf13/cobra"
)

var SitesCmd = &cobra.Command{
	Use:   "sites",
	Short: "Manage all sites",
	Long:  `Start, stop, or restart all sites on the server.`,
}

var logBasePath = "/var/log"

var execSudo = func(name string, arg ...string) *exec.Cmd {
	// Add -n flag to prevent password prompt
	args := append([]string{"-n"}, arg...)
	cmd := exec.Command("sudo", args...)

	// Set SUDO_ASKPASS to /bin/true to handle password prompts
	cmd.Env = append(os.Environ(), "SUDO_ASKPASS=/bin/true")

	return cmd
}

func init() {
	SitesCmd.AddCommand(sitesStartCmd)
	SitesCmd.AddCommand(sitesStopCmd)
	SitesCmd.AddCommand(sitesRestartCmd)
	SitesCmd.AddCommand(sitesNewCmd)

	// Add flags for the new command
	sitesNewCmd.Flags().String("type", "", "Site type (e.g., wp)")
	sitesNewCmd.Flags().String("domain", "", "Domain or subdomain")
	sitesNewCmd.Flags().String("db_source", "", "Database source (internal or external)")
	sitesNewCmd.Flags().String("db_host", "", "Database host (required for external DB)")
	sitesNewCmd.Flags().String("db_port", "", "Database port (required for external DB)")
	sitesNewCmd.Flags().String("db_name", "", "Database name (required for external DB)")
	sitesNewCmd.Flags().String("db_user", "", "Database user (required for external DB)")
	sitesNewCmd.Flags().String("db_password", "", "Database password (required for external DB)")
	sitesNewCmd.Flags().String("scaling_type", "", "Scaling type (dynamic or static)")
	sitesNewCmd.Flags().Int("replicas", 1, "Number of replicas")
	sitesNewCmd.Flags().Int("max_replicas", 0, "Maximum number of replicas (required for dynamic scaling)")
	sitesNewCmd.Flags().String("webhook", "", "Webhook URL for progress updates (optional)")
	sitesNewCmd.Flags().String("site_id", "", "Unique identifier for the site (optional)")
	sitesNewCmd.Flags().String("hostname", "", "Hostname for the site (optional)")
	sitesNewCmd.Flags().String("php_version", "8.3", "PHP version for WordPress (default: 8.3)")
}

var sitesStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start all sites",
	Long:  `Start all sites on the server.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting all sites...")
		startAllSites()
	},
}

var sitesStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop all sites",
	Long:  `Stop all sites on the server.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Stopping all sites...")
		stopAllSites()
	},
}

var sitesRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart all sites",
	Long:  `Restart all sites on the server.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Restarting all sites...")
		stopAllSites()
		startAllSites()
	},
}

var sitesNewCmd = &cobra.Command{
	Use:   "new",
	Short: "Launch a new site",
	Long:  `Launch a new site with specified parameters`,
	Run:   runNewSite,
}

var getDockerComposeTemplate = docker.GetDockerComposeTemplate

func startAllSites() {
	sitesDir := common.HomeDir
	foundSite := false

	entries, err := os.ReadDir(sitesDir)
	if err != nil {
		color.Red("Error reading directory %s: %v\n", sitesDir, err)
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}

			path := filepath.Join(sitesDir, entry.Name())
			composePath := filepath.Join(path, "docker-compose.yml")
			if _, err := os.Stat(composePath); err == nil {
				color.Yellow("Starting site in %s\n", filepath.Base(path))
				err := docker.RunCompose(composePath, "up", "-d")

				if nil != err {
					continue
				}

				foundSite = true
			}
		}
	}

	if !foundSite {
		fmt.Println("No sites found to start.")
	}
}

func stopAllSites() {
	sitesDir := common.HomeDir
	foundSite := false

	entries, err := os.ReadDir(sitesDir)
	if err != nil {
		color.Red("Error reading directory %s: %v\n", sitesDir, err)
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}

			path := filepath.Join(sitesDir, entry.Name())
			composePath := filepath.Join(path, "docker-compose.yml")
			if _, err := os.Stat(composePath); err == nil {
				color.Yellow("Stopping site in %s\n", filepath.Base(path))
				err := docker.RunCompose(composePath, "down")

				if nil != err {
					continue
				}

				foundSite = true
			}
		}
	}

	if !foundSite {
		fmt.Println("No sites found to stop.")
	}
}

func runNewSite(cmd *cobra.Command, args []string) {
	// Get all flags
	siteType, _ := cmd.Flags().GetString("type")
	domain, _ := cmd.Flags().GetString("domain")
	dbSource, _ := cmd.Flags().GetString("db_source")
	scalingType, _ := cmd.Flags().GetString("scaling_type")
	replicas, _ := cmd.Flags().GetInt("replicas")
	maxReplicas, _ := cmd.Flags().GetInt("max_replicas")
	webhook, _ := cmd.Flags().GetString("webhook")
	siteID, _ := cmd.Flags().GetString("site_id")
	hostname, _ := cmd.Flags().GetString("hostname")
	phpVersion, _ := cmd.Flags().GetString("php_version")

	// Set default domain if not provided
	if domain == "" {
		if hostname != "" {
			domain = hostname + ".localhost"
		} else {
			domain = "site.localhost"
		}
	}

	// Check nginx-proxy status and install if needed
	if err := setupNginxProxy(webhook); err != nil {
		color.Red("Error setting up nginx-proxy: %v", err)
		return
	}

	// Create nginx configuration
	if err := createNginxConfig(domain, webhook); err != nil {
		color.Red("Error creating nginx configuration: %v", err)
		return
	}

	// Validate and prompt for missing required fields
	siteType = promptIfEmpty(siteType, "Enter site type (wp):", "wp")
	domain = promptIfEmpty(domain, "Enter domain or subdomain:", "")
	dbSource = promptIfEmpty(dbSource, "Enter database source (internal/external):", "internal")

	var dbHost, dbPort, dbName, dbUser, dbPassword string
	if dbSource == "external" {
		dbHost, _ = cmd.Flags().GetString("db_host")
		dbPort, _ = cmd.Flags().GetString("db_port")
		dbName, _ = cmd.Flags().GetString("db_name")
		dbUser, _ = cmd.Flags().GetString("db_user")
		dbPassword, _ = cmd.Flags().GetString("db_password")

		dbHost = promptIfEmpty(dbHost, "Enter database host:", "")
		dbPort = promptIfEmpty(dbPort, "Enter database port:", "")
		dbName = promptIfEmpty(dbName, "Enter database name:", "")
		dbUser = promptIfEmpty(dbUser, "Enter database user:", "")
		dbPassword = promptIfEmpty(dbPassword, "Enter database password:", "")
	}

	scalingType = promptIfEmpty(scalingType, "Enter scaling type (dynamic/static):", "static")
	if scalingType == "dynamic" && maxReplicas == 0 {
		maxReplicas = promptInt("Enter maximum number of replicas:", replicas)
	}

	// Validate inputs
	if err := validateInputs(siteType, domain, dbSource, scalingType, replicas, maxReplicas); err != nil {
		color.Red("Error: %v", err)
		return
	}

	// Check and setup MySQL if needed
	if dbSource == "internal" {
		if err := setupInternalMySQL(); err != nil {
			color.Red("Error setting up internal MySQL: %v", err)
			return
		}
	}

	// Launch the site
	if err := launchSite(
		siteType, domain, dbSource, dbHost, dbPort, dbName, dbUser, dbPassword, scalingType, replicas, maxReplicas,
		siteID, hostname, phpVersion, webhook,
	); err != nil {
		color.Red("Error launching site: %v", err)
		return
	}

	color.Green("Site launched successfully!")

	// Send webhook if provided
	if webhook != "" {
		sendWebhook(webhook, "Site launched successfully")
	}
}

func promptIfEmpty(value, prompt, defaultValue string) string {
	if value == "" {
		fmt.Print(prompt + " ")
		fmt.Scanln(&value)
		if value == "" {
			value = defaultValue
		}
	}
	return value
}

func promptInt(prompt string, minValue int) int {
	var value int
	for {
		fmt.Print(prompt + " ")
		fmt.Scanln(&value)
		if value >= minValue {
			break
		}
		fmt.Printf("Value must be at least %d\n", minValue)
	}
	return value
}

func validateInputs(siteType, domain, dbSource, scalingType string, replicas, maxReplicas int) error {
	if siteType != "wp" {
		return errors.New("invalid site type")
	}
	if domain == "" {
		return errors.New("domain is required")
	}
	if dbSource != "internal" && dbSource != "external" {
		return errors.New("invalid database source")
	}
	if scalingType != "dynamic" && scalingType != "static" {
		return errors.New("invalid scaling type")
	}
	if replicas < 1 {
		return errors.New("replicas must be at least 1")
	}
	if scalingType == "dynamic" && maxReplicas < replicas {
		return errors.New("max_replicas must be greater than or equal to replicas")
	}
	return nil
}

func setupInternalMySQL() error {
	// Check if MySQL is running
	output, err := execCommand("ploy", "services", "status", "mysql").Output()
	if err != nil || !strings.Contains(string(output), "running") {
		// Install MySQL
		cmd := execCommand("ploy", "services", "install", "mysql")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install MySQL: %v", err)
		}
	}
	return nil
}

// Add this function to help with variable replacement
func replaceTemplateVariables(content string, vars map[string]string) string {
	result := content
	for key, value := range vars {
		result = strings.ReplaceAll(result, "${"+key+"}", value)
	}
	return result
}

func launchSite(
	siteType, domain, dbSource, dbHost, dbPort, dbName, dbUser, dbPassword, scalingType string,
	replicas, maxReplicas int, siteID, hostname, phpVersion string, webhook string,
) error {
	// Start logging
	if err := createSiteLog(hostname, "Starting site creation process"); err != nil {
		return fmt.Errorf("failed to create site log: %v", err)
	}

	// Set default domain if not provided
	if domain == "" {
		if hostname != "" {
			domain = hostname + ".localhost"
		} else {
			domain = "site.localhost"
		}
		createSiteLog(hostname, fmt.Sprintf("Using default domain: %s", domain))
	}

	// Create nginx configuration first
	createSiteLog(hostname, "Creating nginx configuration...")
	if err := createNginxConfig(domain, webhook); err != nil {
		createSiteLog(hostname, fmt.Sprintf("Failed to create nginx configuration: %v", err))
		return fmt.Errorf("failed to create nginx configuration: %v", err)
	}
	createSiteLog(hostname, "Nginx configuration created successfully")

	// Get MySQL details if using internal database
	if dbSource == "internal" {
		createSiteLog(hostname, "Checking MySQL status...")
		sendWebhook(webhook, "Checking MySQL status...")
		mysqlStatus, err := checkMySQLStatus()
		if err != nil {
			createSiteLog(hostname, fmt.Sprintf("Error checking MySQL status: %v", err))
			sendWebhook(webhook, fmt.Sprintf("Error checking MySQL status: %v", err))
			return err
		}

		if !mysqlStatus {
			createSiteLog(hostname, "MySQL is not running. Installing MySQL...")
			sendWebhook(webhook, "MySQL is not running. Installing MySQL...")
			if err := installMySQL("default_user", "default_password", "3306"); err != nil {
				createSiteLog(hostname, fmt.Sprintf("Error installing MySQL: %v", err))
				sendWebhook(webhook, fmt.Sprintf("Error installing MySQL: %v", err))
				return err
			}
			createSiteLog(hostname, "MySQL installed successfully")
			sendWebhook(webhook, "MySQL installed successfully.")
		}

		createSiteLog(hostname, "Fetching MySQL details...")
		sendWebhook(webhook, "Fetching MySQL details...")
		mysqlDetails, err := getServiceDetails("mysql")
		if err != nil {
			createSiteLog(hostname, fmt.Sprintf("Error fetching MySQL details: %v", err))
			sendWebhook(webhook, fmt.Sprintf("Error fetching MySQL details: %v", err))
			return err
		}
		dbHost = mysqlDetails["Host"]
		dbPort = mysqlDetails["Port"]
		dbName = mysqlDetails["Database"]
		dbUser = mysqlDetails["User"]
		dbPassword = mysqlDetails["Password"]
		createSiteLog(hostname, "MySQL details fetched successfully")
	}

	// Choose the appropriate Docker Compose template
	templateFilename := docker.WPComposeStaticTemplate
	if scalingType == "dynamic" {
		templateFilename = docker.WPComposeDynamicTemplate
	}

	// Fetch the Docker Compose template from GitHub
	templateContent, err := getDockerComposeTemplate(templateFilename)
	if err != nil {
		return fmt.Errorf("failed to fetch Docker Compose template: %v", err)
	}

	// If phpVersion is not provided, use the default
	if phpVersion == "" {
		phpVersion = "8.3"
	}

	// Create variables map for replacement
	vars := map[string]string{
		"PHP_VERSION":           phpVersion,
		"HOSTNAME":              hostname,
		"SITE_ID":               siteID,
		"DOMAIN":                domain,
		"WORDPRESS_DB_HOST":     dbHost,
		"WORDPRESS_DB_USER":     dbUser,
		"WORDPRESS_DB_PASSWORD": dbPassword,
		"WORDPRESS_DB_NAME":     dbName,
	}

	// Replace variables in the template
	composeContent := string(templateContent)
	for key, value := range vars {
		if value == "" {
			continue // Skip empty values
		}
		placeholder := "${" + key + "}"
		composeContent = strings.ReplaceAll(composeContent, placeholder, value)
	}

	// Create the site directory
	siteDir := filepath.Join(common.SitesDir, hostname)
	if err := os.MkdirAll(siteDir, 0755); err != nil {
		return fmt.Errorf("failed to create site directory: %v", err)
	}

	// Write the Docker Compose file
	composeFileName := fmt.Sprintf("docker-compose-wp-php%s.yml", phpVersion)
	composeFilePath := filepath.Join(siteDir, composeFileName)
	if err := os.WriteFile(composeFilePath, []byte(composeContent), 0644); err != nil {
		return fmt.Errorf("failed to write docker-compose file: %v", err)
	}

	// Launch the containers
	if os.Getenv("PLOY_TEST_ENV") != "true" {
		cmd := execCommand("docker-compose", "-f", composeFilePath, "up", "-d")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to launch containers: %v", err)
		}
	}

	createSiteLog(hostname, "Site launched successfully")
	sendWebhook(webhook, "Site launched successfully!")
	return nil
}

func checkMySQLStatus() (bool, error) {
	cmd := execCommand("ploy", "services", "status", "mysql")
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return strings.Contains(string(output), "is running"), nil
}

func sendWebhook(url, message string) {
	if url == "" {
		return
	}
	payload, _ := json.Marshal(map[string]string{"message": message})
	resp, err := http.Post(url, "application/json", strings.NewReader(string(payload)))
	if err != nil {
		color.Yellow("Failed to send webhook: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		color.Yellow("Webhook response status: %s", resp.Status)
	}
}

func setupNginxProxy(webhook string) error {
	sendWebhook(webhook, "Checking nginx-proxy status...")

	// Check if nginx-proxy is running
	cmd := execCommand("ploy", "services", "status", "nginx-proxy")
	output, err := cmd.Output()
	if err != nil || !strings.Contains(string(output), "is running") {
		sendWebhook(webhook, "Installing nginx-proxy...")

		// Install nginx-proxy
		cmd = execCommand("ploy", "services", "install", "nginx-proxy")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install nginx-proxy: %v", err)
		}
		sendWebhook(webhook, "nginx-proxy installed successfully")
	}

	return nil
}

var nginxBasePath = "/etc/nginx"

func createNginxConfig(domain string, webhook string) error {
	sendWebhook(webhook, "Creating nginx configuration...")

	// Create container name based on domain
	containerName := strings.ReplaceAll(domain, ".", "-")

	configContent := fmt.Sprintf(
		`server {
	listen 80;
	server_name %s;
	
	location / {
		proxy_pass http://%s:80;
		proxy_set_header Host $host;
		proxy_set_header X-Real-IP $remote_addr;
		proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
		proxy_set_header X-Forwarded-Proto $scheme;
		proxy_redirect off;
		proxy_buffering off;
		
		# Add WebSocket support
		proxy_http_version 1.1;
		proxy_set_header Upgrade $http_upgrade;
		proxy_set_header Connection "upgrade";
	}
}`, domain, containerName,
	)

	// Create nginx sites directory if it doesn't exist
	nginxSitesDir := filepath.Join(nginxBasePath, "sites-available")
	nginxEnabledDir := filepath.Join(nginxBasePath, "sites-enabled")

	// First, try to create the directories with sudo
	cmd := execSudo("sh", "-c", fmt.Sprintf("mkdir -p %s %s", nginxSitesDir, nginxEnabledDir))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create nginx directories: %v", err)
	}

	// Write nginx configuration using a temporary file
	tempFile, err := os.CreateTemp("", "nginx-conf-")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.WriteString(configContent); err != nil {
		return fmt.Errorf("failed to write to temporary file: %v", err)
	}
	tempFile.Close()

	// Move the temporary file to the nginx sites-available directory using sudo
	configPath := filepath.Join(nginxSitesDir, domain+".conf")
	cmd = execSudo("sh", "-c", fmt.Sprintf("mv %s %s && chown root:root %s && chmod 644 %s",
		tempFile.Name(), configPath, configPath, configPath))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install nginx configuration: %v", err)
	}

	// Create symlink in sites-enabled using sudo
	enabledPath := filepath.Join(nginxEnabledDir, domain+".conf")
	cmd = execSudo("sh", "-c", fmt.Sprintf("rm -f %s && ln -s %s %s",
		enabledPath, configPath, enabledPath))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to enable nginx configuration: %v", err)
	}

	// Skip nginx reload in test environment
	if os.Getenv("PLOY_TEST_ENV") != "true" {
		// Reload nginx using sudo
		cmd = execSudo("systemctl", "reload", "nginx")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to reload nginx: %v", err)
		}
	}

	sendWebhook(webhook, "Nginx configuration created and enabled")
	return nil
}

func createSiteLog(hostname, message string) error {
	// Create log directory with sudo if needed
	logDir := filepath.Join(logBasePath, "sites", hostname)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		// Try with sudo if regular mkdir fails
		cmd := execSudo("mkdir", "-p", logDir)
		if err := cmd.Run(); err != nil {
			// If sudo fails, try to create directory as current user
			if err := os.MkdirAll(logDir, 0755); err != nil {
				return fmt.Errorf("failed to create log directory: %v", err)
			}
		}
		// Set permissions
		cmd = execSudo("chmod", "755", logDir)
		if err := cmd.Run(); err != nil {
			// If sudo fails, try to set permissions as current user
			if err := os.Chmod(logDir, 0755); err != nil {
				return fmt.Errorf("failed to set log directory permissions: %v", err)
			}
		}
	}

	// Use deploy.log instead of creating.log
	logFile := filepath.Join(logDir, "deploy.log")

	// Try to open file directly first
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		// If direct access fails, try with sudo
		cmd := execSudo("touch", logFile)
		if err := cmd.Run(); err != nil {
			// If sudo fails, try to create file as current user
			f, err = os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return fmt.Errorf("failed to create log file: %v", err)
			}
		}

		// Set permissions
		cmd = execSudo("chmod", "644", logFile)
		if err := cmd.Run(); err != nil {
			// If sudo fails, try to set permissions as current user
			if err := os.Chmod(logFile, 0644); err != nil {
				return fmt.Errorf("failed to set log file permissions: %v", err)
			}
		}

		// Try opening again after setting permissions
		f, err = os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file after setting permissions: %v", err)
		}
	}
	defer f.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	_, err = fmt.Fprintf(f, "[%s] %s\n", timestamp, message)
	return err
}
