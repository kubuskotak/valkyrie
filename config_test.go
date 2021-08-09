package valkyrie

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type constants struct {
	App struct {
		Name         string         `yaml:"name"`
		Port         int            `yaml:"port"`
		ReadTimeout  int            `yaml:"read_timeout"`
		WriteTimeout int            `yaml:"write_timeout"`
		Timezone     string         `yaml:"timezone"`
		Debug        bool           `yaml:"debug"`
		Env          string         `yaml:"env"`
		SecretKey    string         `yaml:"secret_key"`
		ExpireIn     *time.Duration `yaml:"expire_in"`
	} `yaml:"App"`

	DB struct {
		DsnMain string `yaml:"dsn_main" env:"DSN_MAIN"`
	}
}

func TestConfig(t *testing.T) {
	var cfg constants
	err := Config(ConfigOpts{
		Config:    &cfg,
		Filenames: []string{"app.test.yaml"},
		Paths:     []string{".", "./config"},
	})

	assert.Error(t, err)
	assert.Equal(t, 8778, cfg.App.Port)
}

func TestConfigPathFail(t *testing.T) {
	var cfg constants
	err := Config(ConfigOpts{
		Config:    &cfg,
		Filenames: []string{"app.test.yaml"},
		Paths:     []string{"./config"},
	})

	assert.Error(t, err)
}

func TestConfigEnv(t *testing.T) {
	defer os.Clearenv()
	var cfg constants
	val := "host=localhost port=5999 user=root password=root123 dbname=dbroot sslmode=disable"

	err := Config(ConfigOpts{
		Config:    &cfg,
		Filenames: []string{"app.test.yaml"},
		Paths:     []string{".", "./config"},
	})

	assert.Error(t, err)
	assert.Equal(t, val, cfg.DB.DsnMain)
}

func TestConfigFunc(t *testing.T) {
	defer os.Clearenv()
	var cfg constants
	val := "host=localhost port=5999 user=root password=root123 dbname=dbroot sslmode=disable"
	_ = os.Setenv("DSN_MAIN", val)

	err := Config(ConfigOpts{
		Config:    &cfg,
		Filenames: []string{"app.test.yaml"},
		Paths:     []string{".", "./config"},
	})

	assert.Error(t, err)
	assert.Equal(t, val, cfg.DB.DsnMain)
}
