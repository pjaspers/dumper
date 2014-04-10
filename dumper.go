package main

import (
	"fmt"
	"log"

	"io/ioutil"
	"gopkg.in/yaml.v1"
)

func pg_dump(l interface {}) {
	k := l.(map[interface{}]interface{})
	// fmt.Println("Adapter : ", k["adapter"])
	// fmt.Println("Host    : ", k["host"])
	// fmt.Println("Username:", k["username"])
	// fmt.Println("Password:", k["password"])

  fmt.Printf("PGPASSWORD=%s pg_dump -Fc --no-acl --no-owner --clean -U %s -h %s %s > %s.dump", k["password"], k["username"], k["host"], k["database"], "bla")
}

func main() {
	dat, err := ioutil.ReadFile("/Users/pjaspers/development/rails/casablanco/config/database.yml")
	// fmt.Print(string(dat))

	m := make(map[interface{}]interface{})
	// fmt.Printf("--- t:\n%v\n\n", t.development)
	err = yaml.Unmarshal([]byte(dat), &m)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Println("Development:")
	pg_dump(m["development"])
	fmt.Println("\nStaging:")
	pg_dump(m["staging"])
	fmt.Println("\nProduction:")
	pg_dump(m["production"])
}
