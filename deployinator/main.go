package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

func bringUpApp(appName string) error {
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getcwd failed: %v", err)
	}

	if err := os.Chdir("/deploy-location/" + appName); err != nil {
		return fmt.Errorf("chdir failed: %v", err)
	}
	log.Printf("$ chdir /deploy-location/%s", appName)
	defer os.Chdir(originalDir)

	cmd := exec.Command("echo", "docker-compose", "up", "-d", appName)
	log.Printf("$ docker-compose up -d %s", appName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func main() {
	contents, err := ioutil.ReadFile("/app-registry/apps.json")
	if err != nil {
		log.Fatal("could not read apps.json")
	}
	var appsRegistry map[string]struct{}
	if err = json.Unmarshal(contents, appsRegistry); err != nil {
		log.Fatal("malformed apps.json")
	}
	for appName := range appsRegistry {
		if err := bringUpApp(appName); err != nil {
			log.Printf("Warning: bringing up %s failed: %v\n", appName, err)
		}
	}

	// TODO: does docker-compose up'ing something that is already up restart?
	// err = serve HTTP
	// bring down each child
}
