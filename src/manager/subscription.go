package manager

import (
	"encoding/json"
	"fmt"
	"os"
)

const subscriptionsFile = "subscriptions.json"

type Subscriptions struct {
	Categories map[string][]string `json:"categories"`
}

func LoadSubscriptions() (*Subscriptions, error) {
	file, err := os.Open(subscriptionsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &Subscriptions{Categories: make(map[string][]string)}, nil
		}
		return nil, err
	}
	defer file.Close()
	var subs Subscriptions
	if err := json.NewDecoder(file).Decode(&subs); err != nil {
		return nil, err
	}
	if subs.Categories == nil {
		subs.Categories = make(map[string][]string)
	}
	return &subs, nil
}

func SaveSubscriptions(subs *Subscriptions) error {
	file, err := os.Create(subscriptionsFile)
	if err != nil {
		return err
	}
	defer file.Close()
	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	return enc.Encode(subs)
}

func CreateCategory(name string) error {
	subs, err := LoadSubscriptions()
	if err != nil {
		return err
	}
	if _, exists := subs.Categories[name]; exists {
		return fmt.Errorf("category '%s' already exists", name)
	}
	subs.Categories[name] = []string{}
	return SaveSubscriptions(subs)
}

func DeleteCategory(name string) error {
	subs, err := LoadSubscriptions()
	if err != nil {
		return err
	}
	if _, exists := subs.Categories[name]; !exists {
		return fmt.Errorf("category '%s' does not exist", name)
	}
	delete(subs.Categories, name)
	return SaveSubscriptions(subs)
}

func AddChannelToCategory(category, channel string) error {
	subs, err := LoadSubscriptions()
	if err != nil {
		return err
	}
	channels, exists := subs.Categories[category]
	if !exists {
		return fmt.Errorf("category '%s' does not exist", category)
	}
	for _, ch := range channels {
		if ch == channel {
			return fmt.Errorf("channel '%s' is already in category '%s'", channel, category)
		}
	}
	subs.Categories[category] = append(channels, channel)
	return SaveSubscriptions(subs)
}

func RemoveChannelFromCategory(category, channel string) error {
	subs, err := LoadSubscriptions()
	if err != nil {
		return err
	}
	channels, exists := subs.Categories[category]
	if !exists {
		return fmt.Errorf("category '%s' does not exist", category)
	}
	found := false
	newChannels := []string{}
	for _, ch := range channels {
		if ch == channel {
			found = true
		} else {
			newChannels = append(newChannels, ch)
		}
	}
	if !found {
		return fmt.Errorf("channel '%s' not found in category '%s'", channel, category)
	}
	subs.Categories[category] = newChannels
	return SaveSubscriptions(subs)
}

func GetCategories() ([]string, error) {
	subs, err := LoadSubscriptions()
	if err != nil {
		return nil, err
	}
	var cats []string
	for k := range subs.Categories {
		cats = append(cats, k)
	}
	return cats, nil
}

func GetChannels(category string) ([]string, error) {
	subs, err := LoadSubscriptions()
	if err != nil {
		return nil, err
	}
	channels, exists := subs.Categories[category]
	if !exists {
		return nil, fmt.Errorf("category '%s' does not exist", category)
	}
	return channels, nil
}
