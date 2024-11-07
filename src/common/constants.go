package common

import (
	"os"
	"path/filepath"
)

const CurrentCliVersion = "0.5.8"

var (
	HomeDir       = os.Getenv("HOME")
	ServicesDir   = filepath.Join(HomeDir, ".ploy")
	SitesDir      = filepath.Join(ServicesDir, "sites")
	GlobalCompose = filepath.Join(ServicesDir, "docker-compose.yml")
	ProvisionsDir = filepath.Join(ServicesDir, "provisions")
	MysqlDir      = filepath.Join(ServicesDir, "database", "mysql")
	RedisDir      = filepath.Join(ServicesDir, "database", "redis")
	NginxDir      = filepath.Join(ServicesDir, "nginx")
)

func SetServicesDir(dir string)    { ServicesDir = dir }
func SetGlobalCompose(path string) { GlobalCompose = path }
func SetProvisionsDir(dir string)  { ProvisionsDir = dir }
func SetMysqlDir(dir string)       { MysqlDir = dir }
func SetRedisDir(dir string)       { RedisDir = dir }
func SetNginxDir(dir string)       { NginxDir = dir }
