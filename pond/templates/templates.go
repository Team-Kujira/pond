package templates

import (
	"embed"
	"strings"
)

//go:embed config plan
var Templates embed.FS

func GetPlans() ([]string, error) {
	plans := []string{}

	entries, err := Templates.ReadDir("plan")
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		filename := entry.Name()
		if strings.HasSuffix(filename, ".json") {
			plans = append(plans, strings.Replace(filename, ".json", "", 1))
		}
	}

	return plans, nil
}

func GetChains() ([]string, error) {
	chains := []string{}

	entries, err := Templates.ReadDir("config")
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.Name() == "proxy.conf" {
			continue
		}
		chains = append(chains, entry.Name())
	}

	return chains, nil
}
