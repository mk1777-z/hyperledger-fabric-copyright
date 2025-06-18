package conf

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

var CategoryEncode map[string]int
var CategoryDecode map[int]string

func InitCategory(configPath string) error {
	log.Print("正在配置Category")
	yamlData, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("failed to read config from \"%s\":%v", configPath, err)
		return err
	}
	rawMap := make(map[string]interface{})
	err = yaml.Unmarshal(yamlData, &rawMap)
	if err != nil {
		log.Printf("failed to parse config")
		return err
	}
	categoryMap, ok := rawMap["category"].(map[any]any)
	if !ok {
		log.Printf("category config is not a valid map")
		return fmt.Errorf("category config is not a valid map")
	}

	CategoryDecode = make(map[int]string)
	CategoryEncode = make(map[string]int)
	for k, v := range categoryMap {
		key, ok1 := k.(int)
		val, ok2 := v.(string)
		if !ok1 || !ok2 {
			log.Printf("invalid category key \"%s\": %v", k, err)
			return fmt.Errorf("invalid category key \"%s\": %v", k, err)
		}
		CategoryDecode[key] = val
		CategoryEncode[val] = key
	}
	// log.Print(CategoryDecode)

	return nil
}
