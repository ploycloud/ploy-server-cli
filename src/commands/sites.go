package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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
	dbHost, _ := cmd.Flags().GetString("db_host")
	dbPort, _ := cmd.Flags().GetString("db_port")
	dbName, _ := cmd.Flags().GetString("db_name")
	dbUser, _ := cmd.Flags().GetString("db_user")
	dbPassword, _ := cmd.Flags().GetString("db_password")
	scalingType, _ := cmd.Flags().GetString("scaling_type")
	replicas, _ := cmd.Flags().GetInt("replicas")
	maxReplicas, _ := cmd.Flags().GetInt("max_replicas")
	webhook, _ := cmd.Flags().GetString("webhook")

	// Validate and prompt for missing required fields
	siteType = promptIfEmpty(siteType, "Enter site type (wp):", "wp")
	domain = promptIfEmpty(domain, "Enter domain or subdomain:", "")
	dbSource = promptIfEmpty(dbSource, "Enter database source (internal/external):", "internal")

	if dbSource == "external" {
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
	if err := launchSite(siteType, domain, dbSource, dbHost, dbPort, dbName, dbUser, dbPassword, scalingType, replicas, maxReplicas); err != nil {
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

func launchSite(siteType, domain, dbSource, dbHost, dbPort, dbName, dbUser, dbPassword, scalingType string, replicas, maxReplicas int) error {
	// Get MySQL details if using internal database
	if dbSource == "internal" {
		mysqlDetails, err := getServiceDetails("mysql")
		if err != nil {
			return err
		}
		dbHost = mysqlDetails["Host"]
		dbPort = mysqlDetails["Port"]
		dbName = mysqlDetails["Database"]
		dbUser = mysqlDetails["User"]
		dbPassword = mysqlDetails["Password"]
	}

	// Choose the appropriate Docker Compose template
	templatePath := filepath.Join(os.Getenv("HOME"), "docker", "wp", "wp-compose-static.yml")
	if scalingType == "dynamic" {
		templatePath = filepath.Join(os.Getenv("HOME"), "docker", "wp", "wp-compose-dynamic.yml")
	}

	// Read the Docker Compose template
	templateContent, err := ioutil.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read Docker Compose template: %v", err)
	}

	// Replace placeholders in the template
	composeContent := strings.ReplaceAll(string(templateContent), "${DB_HOST}", dbHost)
	composeContent = strings.ReplaceAll(composeContent, "${DB_PORT}", dbPort)
	composeContent = strings.ReplaceAll(composeContent, "${DB_NAME}", dbName)
	composeContent = strings.ReplaceAll(composeContent, "${DB_USER}", dbUser)
	composeContent = strings.ReplaceAll(composeContent, "${DB_PASSWORD}", dbPassword)
	composeContent = strings.ReplaceAll(composeContent, "${DOMAIN}", domain)
	composeContent = strings.ReplaceAll(composeContent, "${REPLICAS}", fmt.Sprintf("%d", replicas))

	// Write the Docker Compose file
	composeFilePath := filepath.Join(os.Getenv("HOME"), domain, "docker-compose.yml")
	if err := os.MkdirAll(filepath.Dir(composeFilePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}
	if err := ioutil.WriteFile(composeFilePath, []byte(composeContent), 0644); err != nil {
		return fmt.Errorf("failed to write docker-compose file: %v", err)
	}

	// Launch the site using docker-compose
	cmd := execCommand("docker-compose", "-f", composeFilePath, "up", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to launch site: %v", err)
	}

	return nil
}

func sendWebhook(url, message string) {
	payload, _ := json.Marshal(map[string]string{"message": message})
	resp, err := http.Post(url, "application/json", strings.NewReader(string(payload)))
	if err != nil {
		color.Yellow("Failed to send webhook: %v", err)
		return
	}
	defer resp.Body.Close()
}
