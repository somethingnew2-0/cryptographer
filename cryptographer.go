package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path"

	"github.com/cenkalti/backoff"
	"github.com/fsouza/go-dockerclient"
	"github.com/somethingnew2-0/crypt/config"
)

func getopt(name, def string) string {
	if env := os.Getenv(name); env != "" {
		return env
	}
	return def
}

func assert(err error) {
	if err != nil {
		log.Fatal("cryptographer: ", err)
	}
}

func retry(fn func() error) error {
	return backoff.Retry(fn, backoff.NewExponentialBackOff())
}

func NewConfigManager(uri *url.URL, kr io.Reader) (config.ConfigManager, error) {
	factory := map[string]func([]string, io.Reader) (config.ConfigManager, error){
		"consul": config.NewConsulConfigManager,
		"etcd":   config.NewEtcdConfigManager,
	}[uri.Scheme]
	if factory == nil {
		log.Fatal("cryptographer: unrecognized registry backend: ", uri.Scheme)
	}
	log.Println("cryptographer: Using", uri.Scheme, "key value backend at", uri)
	uri.Scheme = "http"
	machines := []string{uri.String()}
	return factory(machines, kr)
}

func main() {
	flag.Parse()

	secrets := getopt("SECRETS_DIR", "/secrets")

	keyRing, err := os.Open(getopt("KEY_RING", "/var/usr/keyring.gpg"))
	defer keyRing.Close()
	assert(err)

	client, err := docker.NewClient(getopt("DOCKER_HOST", "unix:///var/run/docker.sock"))
	assert(err)

	uri, err := url.Parse(flag.Arg(0))
	assert(err)

	manager, err := NewConfigManager(uri, keyRing)
	assert(err)

	containers, err := client.ListContainers(docker.ListContainersOptions{})
	for _, container := range containers {
		fmt.Println("ID: ", container.ID)
		entries, err := manager.List(fmt.Sprintf("/%s", container.ID))
		if err != nil {
			log.Println("cryptographer: can't find directory in config manager for container: ", container.ID, err)
			continue
		}

		containerSecrets := path.Join(secrets, container.ID)
		err = os.MkdirAll(containerSecrets, 0600)
		if err != nil {
			log.Println("cryptographer: error creating container secrets dir: ", container.ID, err)
			continue
		}

		for _, entry := range entries {
			value, err := manager.Get(fmt.Sprintf("/%s/%s", container.ID, entry))
			if err != nil {
				log.Println("cryptographer: error creating container secrets dir: ", container.ID, err)
				continue
			}
			err = ioutil.WriteFile(path.Join(containerSecrets, entry), value, 0600)
			assert(err)
		}
	}
}
