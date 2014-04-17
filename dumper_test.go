package main

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestGreen(t *testing.T) {
	assert.Equal(t, green("some string"), "\033[32msome string\033[0m")
}

func TestFetchWithExistingKey(t *testing.T) {
	k := map[string]interface{}{"a": "key"}
	assert.Equal(t, "key", fetch(k, "a", "not key"))
}

func TestFetchWithoutExistingKey(t *testing.T) {
	k := map[string]interface{}{"a": "key"}
	assert.Equal(t, "not key", fetch(k, "x", "not key"))
}

func TestPGDumpSkipsPasswordIfNoneFound(t *testing.T) {
	k := map[string]interface{}{}
	r := pg_dump(k, "bla")
	assert.False(t, strings.Contains(r, "PGPASSWORD"))
}

func TestPGDumpSetsLocalhostIfNoHostFound(t *testing.T) {
	k := map[string]interface{}{}
	r := pg_dump(k, "bla")
	assert.True(t, strings.Contains(r, "-h localhost"))
}

func TestPGDumpSetsHostIfFound(t *testing.T) {
	k := map[string]interface{}{"host": "somehost"}
	r := pg_dump(k, "bla")
	assert.True(t, strings.Contains(r, "-h somehost"))
}

func TestPGDumpWithAllData(t *testing.T) {
	k := map[string]interface{}{
		"host":     "somehost",
		"username": "Franz",
		"password": "Bob",
		"database": "box",
	}
	expected := "PGPASSWORD=Bob pg_dump -Fc --no-acl --no-owner --clean -U Franz -h somehost box > franz_bob.dump"
	assert.Equal(t, expected, pg_dump(k, "franz_bob"))
}