package main

import (
	"fmt"
	"log"
	"strings"

	"io/ioutil"
	"gopkg.in/yaml.v1"
	"flag"
	"os"

	"path/filepath"
	"time"
	"regexp"
)


func green(s string) (o string){
	s = "\033[32m" + s + "\033[0m"
	return s
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s <environment>\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(2)
}

func mysql_dump(l interface {}, name string) {
	k := l.(map[interface{}]interface{})
  password := fmt.Sprintf("Password: %s", k["password"])
	host, ok := k["host"]
	if !ok {
		host = "localhost"
	}
  command := fmt.Sprintf("mysqldump -u %s -p -h %s %s > %s.sql", k["username"], host, k["database"], name)
	fmt.Printf("%s\n\n%s", password, command)
}

func pg_dump(l interface {}, name string) {
	k := l.(map[interface{}]interface{})
	host, ok := k["host"]
	if !ok {
		host = "localhost"
	}
  fmt.Printf("PGPASSWORD=%s pg_dump -Fc --no-acl --no-owner --clean -U %s -h %s %s > %s.dump", k["password"], k["username"], host, k["database"], name)
}

func main() {
	flag.Usage = usage
	var force bool
	var path string

	flag.BoolVar(&force, "F", false, "Show restore operation")
	flag.StringVar(&path, "p", "", "Path to yaml (otherwise config/database.yml")
	flag.Parse()
	environment := "development"
	if strings.TrimSpace(flag.Arg(0)) != "" {
		environment = strings.TrimSpace(flag.Arg(0))
	}
	if path == "" {
		current_dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
		path = current_dir
	}
	file := filepath.Join(path, "config", "database.yml")
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "No config file found: %s\n", path)
		os.Exit(2)
	}
	name := fmt.Sprintf("%s_%s_%s", filepath.Base(filepath.Dir(path)), environment[0:3], time.Now().Format("20060102"))
	// fmt.Printf(name)
	// fmt.Printf("Environment: %s Force: %t Path %s ", environment, force, path)
	dat, err := ioutil.ReadFile(file)
	// fmt.Print(string(dat))

	m := make(map[interface{}]interface{})
	err = yaml.Unmarshal([]byte(dat), &m)
	if err != nil {
		log.Fatalf("Error parsing yaml: %v", err)
	}
	adapter := fmt.Sprintf("%s", (m[environment].(map[interface{}]interface{}))["adapter"])
	pg := regexp.MustCompile("postgres")
	mysql := regexp.MustCompile("mysql")
	foundPg := pg.MatchString(adapter)
	foundMysql := mysql.MatchString(adapter)

	fmt.Printf("%s\n\n", green("Dump:"))
	if (foundPg) {
		pg_dump(m[environment], name)
	}
	if (foundMysql) {
		mysql_dump(m[environment], name)
	}
}
