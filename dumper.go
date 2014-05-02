package main

import (
	"fmt"
	"strings"

	"flag"
	"gopkg.in/yaml.v1"
	"io/ioutil"
	"os"

	"path/filepath"
	"regexp"
	"time"
)

// # Database Config struct.
//
// Each Rails `database.yml` environment block has the following structure, so
// it maps to this struct.
type DbConfig struct {
	Adapter        string
	Host           string
	Database       string
	Username       string
	Password       string
	ExcludedTables []string
}

// If no host is set, assume localhost.
func (self *DbConfig) SetDefaults() {
	if self.Host == "" {
		self.Host = "localhost"
	}
}

func (self *DbConfig) SetExcludedTables(ignored []string) {
	if len(ignored) > 0 {
		self.ExcludedTables = ignored
	}
}

func (self *DbConfig) HasExcludedTables() bool {
	return len(self.ExcludedTables) > 0
}

func (self *DbConfig) ExcludedTablesWithFlag(flag string) string {
	values := make([]string, 0, len(self.ExcludedTables))
	for _, table := range self.ExcludedTables {
		values = append(values, fmt.Sprintf("%s=%s", flag, table))
	}
	return fmt.Sprintf("%s", strings.Join(values, " "))
}

// These get used later on to find the correct methods to call for dumping or
// restoring (`pg_dump`)
func (self *DbConfig) ShortAdapter() (short string) {
	if regexp.MustCompile("postgres").MatchString(self.Adapter) {
		return "pg"
	}
	if regexp.MustCompile("mysql").MatchString(self.Adapter) {
		return "mysql"
	}
	if regexp.MustCompile("sqlite").MatchString(self.Adapter) {
		return "sqlite"
	}

	return self.Adapter
}

// # Utilities
//
// Returns a green string in ANSI codes
func green(in string) (out string) {
	in = "\033[32m" + in + "\033[0m"
	return in
}

func red(in string) (out string) {
	in = "\033[31m" + in + "\033[0m"
	return in
}

func printError(message string) {
	fmt.Println(red(fmt.Sprintf("\n\t%s", message)))
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s <environment>\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(2)
}

//
// ## SQLite
//

func sqlite_dump(config DbConfig, name string) (out string) {
	command := fmt.Sprintf("sqlite3 %s .dump > %s.dump", config.Database, name)
	return command
}

func sqlite_restore(config DbConfig, name string) (out string) {
	command := fmt.Sprintf("sqlite3 %s < %s.dump", config.Database, name)
	return command
}

//
// ## MySQL
//

func mysql_dump(config DbConfig, name string) (out string) {
	command := ""
	if config.HasExcludedTables() {
		dump_structure := fmt.Sprintf("mysqldump -u %s -p -h %s --no-data %s > %s.sql", config.Username, config.Host, config.Database, name)
		exclude_command := config.ExcludedTablesWithFlag("--ignore-table")
		dump_without_tables := fmt.Sprintf("mysqldump -u %s -p -h %s %s %s > %s.sql", config.Username, config.Host, exclude_command, config.Database, name)
		command = fmt.Sprintf("%s && %s", dump_structure, dump_without_tables)
	} else {
		command = fmt.Sprintf("mysqldump -u %s -p -h %s %s > %s.sql", config.Username, config.Host, config.Database, name)
	}

	if len(config.Password) > 0 {
		command = fmt.Sprintf("Password: %s\n\n%s", config.Password, command)
	}
	return command
}

func mysql_restore(config DbConfig, name string) (out string) {
	command := fmt.Sprintf("mysql -u %s -p -h %s %s < %s.sql", config.Username, config.Host, config.Database, name)
	if len(config.Password) > 0 {
		command = fmt.Sprintf("Password: %s\n\n%s", config.Password, command)
	}
	return command
}

//
// ## PostgreSQL
//

func pg_dump(config DbConfig, name string) (out string) {
	username := config.Username
	hostname := config.Host
	command := ""
	if config.HasExcludedTables() {
		exclude_command := config.ExcludedTablesWithFlag("--exclude-table-data")
		command = fmt.Sprintf("pg_dump -Fc --no-acl --no-owner --clean -U %s -h %s %s %s > %s.dump", username, hostname, exclude_command, config.Database, name)
	} else {
		command = fmt.Sprintf("pg_dump -Fc --no-acl --no-owner --clean -U %s -h %s %s > %s.dump", username, hostname, config.Database, name)
	}
	if len(config.Password) > 0 {
		command = fmt.Sprintf("PGPASSWORD=%s %s", config.Password, command)
	}
	return command
}

func pg_restore(config DbConfig, name string) (out string) {
	username := config.Username
	hostname := config.Host
	command := fmt.Sprintf("pg_restore --verbose --clean --no-acl --no-owner -h %s -U %s -d %s %s.dump", hostname, username, config.Database, name)
	if len(config.Password) > 0 {
		command = fmt.Sprintf("PGPASSWORD=%s %s", config.Password, command)
	}
	return command
}

func get_environment(argument string) (environment string) {
	environment = "development"
	if strings.TrimSpace(argument) != "" {
		environment = strings.TrimSpace(argument)
	}
	return environment
}

func get_config(yamlData []byte, environment string) (config DbConfig, err error) {
	m := make(map[string]DbConfig)
	err = yaml.Unmarshal(yamlData, &m)
	if err != nil {
		return m[environment], err
	}
	config, ok := m[environment]
	if !ok {
		var keys []string
		for key, _ := range m {
			keys = append(keys, key)
		}
		err := fmt.Errorf("No such environment found. Use one of: %s", keys)
		return m[environment], err
	}
	config.SetDefaults()
	return config, nil
}

// Tries to get the directory from which the app is called
var currentDir = func() string {
	p, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	return p
}

// Simple check to see if a file exists, mainly added so it can be
// stubbed in the tests.
var file_exists = func(path string) (err error) {
	_, err = os.Stat(path)
	return err
}

func get_yaml_path(path string) (out string, err error) {
	if path == "" {
		path = currentDir()
	}
	if filepath.Ext(path) != ".yml" {
		path = filepath.Join(path, "config", "database.yml")
	}
	err = file_exists(path)
	return path, err
}

type ignored []string

func (i *ignored) String() string {
	return fmt.Sprint(*i)
}

// Set is the method to set the flag value, part of the flag.Value interface.
// Set's argument is a string to be parsed to set the flag.
// It's a comma-separated list, so we split it.
func (i *ignored) Set(value string) error {
	for _, dt := range strings.Split(value, ",") {
		*i = append(*i, dt)
	}
	return nil
}

func main() {
	flag.Usage = usage
	var force bool
	var path string
	var ignoreFlag ignored
	flag.BoolVar(&force, "F", false, "Show restore operation")
	flag.StringVar(&path, "p", "", "Path to yaml (otherwise config/database.yml)")
	flag.Var(&ignoreFlag, "i", "comma-separated list of tables to ignore")
	flag.Parse()

	environment := get_environment(flag.Arg(0))
	yamlFile, err := get_yaml_path(path)
	if err != nil {
		printError(fmt.Sprintf("Couldn't find a database.yml to parse."))
		os.Exit(2)
	}
	name := fmt.Sprintf("%s_%s_%s", filepath.Base(filepath.Dir(filepath.Dir(yamlFile))), environment[0:3], time.Now().Format("20060102"))
	yamlData, err := ioutil.ReadFile(yamlFile)
	if err != nil {
		printError(fmt.Sprintf("Couldn't read the yaml: %v", err))
		os.Exit(2)
	}
	config, err := get_config(yamlData, environment)
	config.SetExcludedTables(ignoreFlag)
	if err != nil {
		printError(fmt.Sprintf("%s", err))
		os.Exit(2)
	}
	dumpers := map[string]func(DbConfig, string) string{
		"pg":     pg_dump,
		"mysql":  mysql_dump,
		"sqlite": sqlite_dump,
	}
	restorers := map[string]func(DbConfig, string) string{
		"pg":     pg_restore,
		"mysql":  mysql_restore,
		"sqlite": sqlite_restore,
	}
	f, ok := dumpers[config.ShortAdapter()]
	if ok {
		dump := f(config, name)
		fmt.Printf("%s\n\n", green("Dump:"))
		fmt.Printf("%s\n", dump)
		if force {
			fmt.Printf("\n%s\n\n", red("Restore:"))
			fmt.Printf("%s\n", restorers[config.ShortAdapter()](config, name))
		}
	} else {
		printError(fmt.Sprintf("Sorry, don't know how to export %s", config.Adapter))
	}
}
