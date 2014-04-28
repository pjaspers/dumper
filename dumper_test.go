package main

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestGreen(t *testing.T) {
	assert.Equal(t, green("some string"), "\033[32msome string\033[0m")
}

func TestGetConfigSetsDefaults(t *testing.T) {
	var data = `
dev:
  key: "bla"
`
	config, _ := get_config([]byte(data), "dev")
	assert.Equal(t, "localhost", config.Host)
}

func TestConfigShortAdapter(t *testing.T) {
	var examples = map[string]string{
		"postgresql": "pg",
		"postgres":   "pg",
		"sqlite3":    "sqlite",
		"sqlite":     "sqlite",
		"mysql2":     "mysql",
		"mysql":      "mysql",
		"weird":      "weird",
	}
	for k, v := range examples {
		short := (&DbConfig{Adapter: k}).ShortAdapter()
		assert.Equal(t, short, v)
	}
}

func TestPGDumpSkipsPasswordIfNoneFound(t *testing.T) {
	k := DbConfig{}
	r := pg_dump(k, "bla")
	assert.False(t, strings.Contains(r, "PGPASSWORD"))
}

func TestPGDumpSetsHostIfFound(t *testing.T) {
	k := DbConfig{Host: "somehost"}
	r := pg_dump(k, "bla")
	assert.True(t, strings.Contains(r, "-h somehost"))
}

func TestGetEnvironmentWithoutArgumnet(t *testing.T) {
	assert.Equal(t, "development", get_environment(""))
}

func TestGetEnvironmentWithArgument(t *testing.T) {
	assert.Equal(t, "staging", get_environment("staging"))
}

func TestIfNoPathSuppliedGetCurrentDir(t *testing.T) {
	currentDir = func() (s string) { return "/some/path/to/current/dir" }
	path, _ := get_yaml_path("")
	assert.Equal(t, "/some/path/to/current/dir/config/database.yml", path)
}

func TestIfPathSuppliedFindYaml(t *testing.T) {
	currentDir = func() (s string) { return "/this/dir" }
	file_exists = func(path string) (e error) { return nil }
	path, _ := get_yaml_path("/some/other/dir")
	assert.Equal(t, "/some/other/dir/config/database.yml", path)
}

func TestIfYamlGivenUseAllTheYamls(t *testing.T) {
	currentDir = func() (s string) { return "" }
	file_exists = func(path string) (e error) { return nil }
	path, _ := get_yaml_path("ding.yml")
	assert.Equal(t, "ding.yml", path)
}

func TestPGDumpWithAllData(t *testing.T) {
	k := DbConfig{
		Host:     "somehost",
		Username: "Franz",
		Password: "Bob",
		Database: "box",
	}
	expected := "PGPASSWORD=Bob pg_dump -Fc --no-acl --no-owner --clean -U Franz -h somehost box > franz_bob.dump"
	assert.Equal(t, expected, pg_dump(k, "franz_bob"))
}
