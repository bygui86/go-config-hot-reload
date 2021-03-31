package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	consulApi "github.com/hashicorp/consul/api"
)

type Config struct {
	Mode      string `json:"mode"`
	CacheSize int    `json:"cacheSize"`
}

var (
	config                *Config
	configLock            sync.RWMutex
	consulClient          *consulApi.KV
	lastKVPairModifyIndex uint64
)

func main() {
	// create consul client
	initErr := initConsulClient()
	if initErr != nil {
		fmt.Println("init consul client failed: ", initErr.Error())
		os.Exit(1)
	}

	// load config for the first time
	loadErr := loadConfig()
	if loadErr != nil {
		fmt.Println("load config failed: ", loadErr.Error())
		os.Exit(1)
	}

	// start config reloader
	go configReloader()

	fmt.Println(fmt.Sprintf("config: %+v", getConfig()))
	time.Sleep(1 * time.Second)

	// start reload trigger
	trigger()
}

func initConsulClient() error {
	cfg := consulApi.DefaultConfig()

	client, err := consulApi.NewClient(cfg)
	if err != nil {
		return err
	}

	consulClient = client.KV()
	return nil
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
	consulKey := "/samples/app"

	fmt.Println(fmt.Sprintf("load config from consul with key %s", consulKey))

	kvPair, metadata, fetchErr := fetchConfigFromConsul(consulKey)
	if fetchErr != nil {
		return fetchErr
	}

	if kvPair == nil {
		return fmt.Errorf("no KV pair found for %s", consulKey)
	}

	fmt.Println(fmt.Sprintf("KV pair: %+v - Metadata: %+v", kvPair, metadata))

	if kvPair.ModifyIndex > lastKVPairModifyIndex {
		var newCfg Config
		yamlErr := json.Unmarshal(kvPair.Value, &newCfg)
		if yamlErr != nil {
			return yamlErr
		}
		// fmt.Println(fmt.Sprintf("[DEBUG] NEW config reloaded: %+v", newCfg))

		configLock.Lock()
		defer configLock.Unlock()
		config = &newCfg
		lastKVPairModifyIndex = kvPair.ModifyIndex
		fmt.Println("load config completed")

	} else {
		fmt.Println("no config updates to load")
	}

	return nil
}

func fetchConfigFromConsul(key string) (*consulApi.KVPair, *consulApi.QueryMeta, error) {
	fmt.Println(fmt.Sprintf("fetch config from consul with key %s", key))

	return consulClient.Get(key, &consulApi.QueryOptions{})
}

func getConfig() *Config {
	fmt.Println("get config struct")

	configLock.RLock()
	defer configLock.RUnlock()
	return config
}
