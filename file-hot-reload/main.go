package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Mode      string `yaml:"mode"`
	CacheSize int    `yaml:"cacheSize"`
}

var (
	config     *Config
	configLock sync.RWMutex
)

func main() {
	// load config for the first time
	err := loadConfig()
	if err != nil {
		fmt.Println("load config failed: ", err.Error())
		os.Exit(1)
	}

	// start config reloader
	go configReloader()

	fmt.Println(fmt.Sprintf("config: %+v", getConfig()))
	time.Sleep(1 * time.Second)

	// start reload trigger
	trigger()
}

func configReloader() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGUSR2)

	for {
		<-sigChan
		fmt.Println("received signal to reload config")
		err := loadConfig()
		if err != nil {
			fmt.Println("reload config failed: ", err.Error())
		}
	}
}

func trigger() {
	counter := 0
	for {
		fmt.Println("send signal to reload config")

		pid := syscall.Getpid()
		proc, procErr := os.FindProcess(pid)
		if procErr != nil {
			fmt.Println(fmt.Sprintf("find process with pid %d failed: %s", pid, procErr.Error()))
			continue
		}
		sigErr := proc.Signal(syscall.SIGUSR2)
		if sigErr != nil {
			fmt.Println(fmt.Sprintf("send signal to process with pid %d failed: %s", pid, sigErr.Error()))
			continue
		}

		time.Sleep(1 * time.Second)

		fmt.Println(fmt.Sprintf("config: %+v", getConfig()))

		if counter > 10 {
			break
		}

		counter++

		time.Sleep(10 * time.Second)
	}
}

func loadConfig() error {
	fmt.Println("load config from yaml file")

	file, fileErr := ioutil.ReadFile("./file-hot-reload/config.yaml")
	if fileErr != nil {
		return fileErr
	}

	var newCfg Config
	yamlErr := yaml.Unmarshal(file, &newCfg)
	if yamlErr != nil {
		return yamlErr
	}
	// fmt.Println(fmt.Sprintf("[DEBUG] NEW config loaded: %+v", newCfg))

	configLock.Lock()
	defer configLock.Unlock()
	config = &newCfg
	fmt.Println("load config completed")
	return nil
}

func getConfig() *Config {
	fmt.Println("get config struct")

	configLock.RLock()
	defer configLock.RUnlock()
	return config
}
