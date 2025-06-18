package conf

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	gorseCli "github.com/gorse-io/gorse-go"
	"gopkg.in/yaml.v3"
)

var RecommendationCfg *RecommendationConfig
var GorseClient *gorseCli.GorseClient

const sleepTime = 10 * time.Second

type RecommendationConfig struct {
	Apikey   string
	Protocal string
	Host     string
	Port     int
}

func (recommendCfg *RecommendationConfig) GetRecommendationEntryPoint() string {
	return fmt.Sprintf("%s://%s:%d", recommendCfg.Protocal, recommendCfg.Host, recommendCfg.Port)
}

func GetRecommandationConfig(configPath string) (*RecommendationConfig, error) {
	yamlData, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("failed to read config from \"%s\":%v", configPath, err)
		return nil, err
	}
	rawMap := make(map[string]any)
	err = yaml.Unmarshal(yamlData, &rawMap)
	if err != nil {
		log.Printf("failed to parse config")
		return nil, err
	}
	recommandationMap := rawMap["recommendation"].(map[string]any)
	cfg := &RecommendationConfig{
		Apikey:   recommandationMap["apikey"].(string),
		Protocal: recommandationMap["protocal"].(string),
		Host:     recommandationMap["host"].(string),
		Port:     recommandationMap["port"].(int),
	}

	return cfg, nil
}

func InitRecommendationConfig(configPath string) error {
	cfg, err := GetRecommandationConfig(configPath)
	if err != nil {
		log.Printf("failed to get recommendation config: %v", err)
		return err
	}
	RecommendationCfg = cfg
	log.Print("推荐系统配置完成")
	GorseClient = gorseCli.NewGorseClient(cfg.GetRecommendationEntryPoint(), cfg.Apikey)
	return nil
}

func SetItemHidden(itemId string, hidden bool) {
	go setItemHidden(itemId, hidden)
}

func setItemHidden(itemId string, hidden bool) {
	for {
		recommendationUpdate, err := GorseClient.GetItem(context.Background(), itemId)
		if err != nil {
			time.Sleep(sleepTime)
			continue
		}
		if recommendationUpdate.IsHidden == hidden {
			return
		}
		currentTime := time.Now()
		var labels []string
		for _, label := range recommendationUpdate.Labels.([]any) {
			labels = append(labels, label.(string))
		}
		if err != nil {
			log.Print("Error unmarshalling labels:", err)
			time.Sleep(sleepTime)
			continue
		} else {
			_, err := GorseClient.UpdateItem(context.Background(), recommendationUpdate.ItemId, gorseCli.ItemPatch{
				IsHidden:   &hidden,
				Categories: recommendationUpdate.Categories,
				Timestamp:  &currentTime,
				Labels:     labels,
				Comment:    &recommendationUpdate.Comment,
			})
			if err != nil {
				time.Sleep(sleepTime)
				continue
			}
		}
	}
}
