package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConstants(t *testing.T) {
	assert.NotEmpty(t, CurrentCliVersion)
	assert.Equal(t, "/home/ploy", HomeDir)
	assert.Equal(t, HomeDir+"/.ploy", ServicesDir)
	assert.Equal(t, ServicesDir+"/docker-compose.yml", GlobalCompose)
	assert.Equal(t, HomeDir+"/.provisions", ProvisionsDir)
	assert.Equal(t, ServicesDir+"/database/mysql", MysqlDir)
	assert.Equal(t, ServicesDir+"/database/redis", RedisDir)
	assert.Equal(t, ServicesDir+"/nginx", NginxDir)
}
