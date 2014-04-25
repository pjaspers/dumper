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
	Adapter string
	Host string
	Database string
	Username string
	Password string
}

// If no host is set, assume localhost.
func (self *DbConfig) SetDefaults() {
	if self.Host == "" {
		self.Host = "localhost"
	}
}

// These get used later on to find the correct methods to call for dumping or
// restoring (`pg_dump`)
func (self *DbConfig) ShortAdapter() (short string){
	if (regexp.MustCompile("postgres").MatchString(self.Adapter)){ return "pg"}
	if (regexp.MustCompile("mysql").MatchString(self.Adapter))	 { return "mysql"}
	if (regexp.MustCompile("sqlite").MatchString(self.Adapter))	 { return "sqlite"}

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

func sqlite3_dump(l interface{}, name string) {
	fmt.Printf("sqlite3 db/development.sqlite3 .dump > dump")
}

func mysql_dump(config DbConfig, name string) (out string){
	command := fmt.Sprintf("mysqldump -u %s -p -h %s %s > %s.sql", config.Username, config.Host, config.Database, name)
	if len(config.Password) > 0 {
		command = fmt.Sprintf("Password: %s\n\n%s", config.Password, command)
	}
	return command
}

func mysql_restore(config DbConfig, name string) (out string){
	command := fmt.Sprintf("mysql -u %s -p -h %s %s < %s.sql", config.Username, config.Host, config.Database, name)
	if len(config.Password) > 0 {
		command = fmt.Sprintf("Password: %s\n\n%s", config.Password, command)
	}
	return command
}

func pg_dump(config DbConfig, name string) (out string) {
	username := config.Username
	hostname := config.Host
	command := fmt.Sprintf("pg_dump -Fc --no-acl --no-owner --clean -U %s -h %s %s > %s.dump", username, hostname, config.Database, name)
	if len(config.Password) > 0 {
		command = fmt.Sprintf("PGPASSWORD=%s %s", config.Password, command)
	}
	return command
}

func pg_restore(config DbConfig, name string) (out string){
	username := config.Username
	hostname := config.Host
	command := fmt.Sprintf("pg_restore --verbose --clean --no-acl --no-owner -h %s -U %s -d %s %s.dump", hostname, username, config.Database, name)
	if len(config.Password) > 0 {
		command = fmt.Sprintf("PGPASSWORD=%s %s", config.Password, command)
	}
	return command
}

func get_environment(argument string) (environment string){
	environment = "development"
	if strings.TrimSpace(argument) != "" {
		environment = strings.TrimSpace(argument)
	}
	return environment
}

func get_config(yamlData []byte, environment string) (config DbConfig, err error){
	m := make(map[string]DbConfig)
	err = yaml.Unmarshal(yamlData, &m)
	if err != nil {
		return m[environment], err
	}
	config, ok := m[environment]
	if !ok {
		var keys[]string
		for key, _ := range m {
			keys = append(keys, key)
		}
		err := fmt.Errorf("No such environment found. Use one of: %s", keys)
		return m[environment], err
	}
	config.SetDefaults()
	return config, nil
}

var currentDir = func() string {
	p,_ := filepath.Abs(filepath.Dir(os.Args[0]))
	return p
}

var file_exists = func(path string) (err error) {
	_, err = os.Stat(path)
	return err
}

func get_yaml_path (path string) (out string, err error) {
	if path == "" {
		path = currentDir()
	}
	if filepath.Ext(path) != ".yml" {
		path = filepath.Join(path, "config", "database.yml")
	}
	err = file_exists(path)
	return path, err
}

func main() {
	flag.Usage = usage
	var force bool
	var path string

	flag.BoolVar(&force, "F", false, "Show restore operation")
	flag.StringVar(&path, "p", "", "Path to yaml (otherwise config/database.yml)")
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
	if err != nil {
		printError(fmt.Sprintf("%s", err))
		os.Exit(2)
	}
	dumpers := map[string]func(DbConfig, string) string {
		"pg": pg_dump,
		"mysql": mysql_dump,
	}
	restorers := map[string]func(DbConfig, string) string {
		"pg": pg_restore,
		"mysql": mysql_restore,
	}
	f, ok := dumpers[config.ShortAdapter()]
	if ok {
		dump := f(config, name)
		fmt.Printf("%s\n\n", green("Dump:"))
		fmt.Printf("%s\n", dump)
		if force {
			fmt.Printf("\n%s\n\n", red("Restore:"))
			fmt.Printf("%s", restorers[config.ShortAdapter()](config, name))
		}
	} else {
		printError(fmt.Sprintf("Sorry, don't know how to export %s", config.Adapter))
	}
}
