package main

import (
	"fmt"
	"log"

	"golang.org/x/sys/windows/registry"
)

const targetKey = "EnableUlps"

func main() {
	processKey(`SYSTEM`)

	fmt.Println("Done")
	fmt.Scanln()
}

func processKey(baseKey string) {
	subKeys, err := workOn(baseKey)

	if err != nil {
		return
	}

	if len(subKeys) > 0 {
		for _, subKey := range subKeys {
			processKey(baseKey + `\` + subKey)
		}
	}
}

func workOn(key string) ([]string, error) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, key, registry.ALL_ACCESS)
	if err != nil {
		return nil, err
	}
	defer k.Close()

	subValueNames, err := k.ReadValueNames(-1)
	if err != nil {
		return nil, err
	}

	for _, valueName := range subValueNames {
		if valueName == targetKey {
			log.Println("Found value in key: " + key)

			k.SetDWordValue(targetKey, 0)
		}
	}

	subKeys, serr := k.ReadSubKeyNames(-1)
	if serr != nil {
		return nil, err
	}

	return subKeys, nil
}
