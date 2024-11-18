package commands

import (
	"fmt"
	"os"
	"testing"
)

// TestHelperProcess isn't a real test. It's used to mock command execution.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}

	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "No command\n")
		os.Exit(2)
	}

	cmd, args := args[0], args[1:]
	switch cmd {
	case "docker":
		if args[0] == "ps" {
			fmt.Println("mysql-container")
		} else if args[0] == "inspect" {
			if args[1] == "--format" {
				switch args[2] {
				case "{{range .Config.Env}}{{println .}}{{end}}":
					fmt.Println("MYSQL_ROOT_PASSWORD=wp_password")
					fmt.Println("MYSQL_USER=wp_user")
					fmt.Println("MYSQL_DATABASE=wordpress")
				case "{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}":
					fmt.Println("172.17.0.2")
				case "{{range $p, $conf := .NetworkSettings.Ports}}{{if eq $p \"3306/tcp\"}}{{(index $conf 0).HostPort}}{{end}}{{end}}":
					fmt.Println("3306")
				}
			}
		} else if args[0] == "version" {
			if args[1] == "--format" && args[2] == "{{.Server.Version}}" {
				fmt.Println("20.10.14")
			} else {
				fmt.Println("Docker version 20.10.14, build 12345678")
			}
		} else if args[0] == "info" {
			fmt.Println("Docker info output")
		}
	case "sudo":
		if len(args) == 0 {
			fmt.Fprintf(os.Stderr, "No sudo command\n")
			os.Exit(2)
		}

		// Handle sudo commands
		sudoCmd := args[0]
		sudoArgs := args[1:]

		switch sudoCmd {
		case "cp":
			if len(sudoArgs) != 2 {
				fmt.Fprintf(os.Stderr, "Invalid cp arguments\n")
				os.Exit(1)
			}
			// Simulate successful copy
			fmt.Println("File copied successfully")
			os.Exit(0)
		case "chmod":
			if len(sudoArgs) != 2 {
				fmt.Fprintf(os.Stderr, "Invalid chmod arguments\n")
				os.Exit(1)
			}
			fmt.Println("Permissions set successfully")
			os.Exit(0)
		case "mkdir":
			if len(sudoArgs) < 2 || sudoArgs[0] != "-p" {
				fmt.Fprintf(os.Stderr, "Invalid mkdir arguments\n")
				os.Exit(1)
			}
			fmt.Println("Directory created successfully")
			os.Exit(0)
		case "ln":
			if len(sudoArgs) < 3 || sudoArgs[0] != "-s" {
				fmt.Fprintf(os.Stderr, "Invalid ln arguments\n")
				os.Exit(1)
			}
			fmt.Println("Symlink created successfully")
			os.Exit(0)
		case "systemctl":
			fmt.Println("Service managed successfully")
			os.Exit(0)
		default:
			fmt.Fprintf(os.Stderr, "Unknown sudo command: %s\n", sudoCmd)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command %q\n", cmd)
		os.Exit(2)
	}
}
