package pond

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"sort"

	"pond/pond/deployer"
	"pond/utils"

	"github.com/rs/zerolog"
)

type Registry struct {
	logger zerolog.Logger
	path   string
	Data   map[string]deployer.Code
}

func NewRegistry(logger zerolog.Logger, path string) (Registry, error) {
	registry := Registry{
		logger: logger,
		path:   path,
	}

	err := registry.Load(path)
	if err != nil {
		logger.Err(err)
		return registry, err
	}

	return registry, nil
}

func (r *Registry) error(err error) error {
	r.logger.Err(err).Msg("")
	return err
}

func (r *Registry) Load(filename string) error {
	r.logger.Debug().Str("file", filename).Msg("load registry")

	data, err := os.ReadFile(filename)
	if err != nil {
		return r.error(err)
	}

	var items map[string]deployer.Code
	err = json.Unmarshal(data, &items)
	if err != nil {
		return r.error(err)
	}

	r.Data = items

	return nil
}

func (r *Registry) Import(filename string) error {
	r.Load(filename)
	return r.Save()
}

func (r *Registry) Save() error {
	data, err := json.Marshal(r.Data)
	if err != nil {
		return r.error(err)
	}

	err = os.WriteFile(r.path, data, 0o644)
	if err != nil {
		return r.error(err)
	}

	return nil
}

func (r *Registry) Export(filename string) error {
	data, err := json.Marshal(r.Data)
	if err != nil {
		return r.error(err)
	}

	err = os.WriteFile(filename, data, 0o644)
	if err != nil {
		return r.error(err)
	}

	return nil
}

func (r *Registry) List() error {
	padding := 0
	keys := []string{}

	for key := range r.Data {
		keys = append(keys, key)
		length := len(key)
		if length > padding {
			padding = length
		}
	}

	sort.Strings(keys)

	for _, key := range keys {
		code := r.Data[key]
		checksum := code.Checksum[:8] + "…" + code.Checksum[56:]
		fmt.Printf("%s %-*s %s\n", checksum, padding, key, code.Source)
	}

	return nil
}

func (r *Registry) Update(name string, args map[string]string) error {
	r.logger.Debug().Msg("update registry")

	item, found := r.Data[name]
	if !found {
		err := fmt.Errorf("code not registered")
		r.logger.Err(err).Str("name", name).Msg("")
		return err
	}

	newName, found := args["name"]
	if found {
		delete(r.Data, name)
		name = newName
	}

	newSource, found := args["source"]
	if found {
		parts, err := url.Parse(newSource)
		if err != nil {
			return r.error(err)
		}

		switch parts.Scheme {
		case "file":
			data, err := os.ReadFile(parts.Path)
			if err != nil {
				return r.error(err)
			}

			item.Checksum = utils.Sha256(data)
		case "kaiyo-1":
			break
		default:
			err := fmt.Errorf("scheme not supported")
			r.logger.Err(err).Str("scheme", parts.Scheme).Msg("")
			return err
		}

		item.Source = newSource
	}

	r.Data[name] = item

	return r.Save()
}

func (p *Pond) ListRegistry() error {
	return p.registry.List()
}

func (p *Pond) UpdateRegistry(name string, args map[string]string) error {
	return p.registry.Update(name, args)
}

func (p *Pond) ExportRegistry(filename string) error {
	return p.registry.Export(filename)
}

func (p *Pond) ImportRegistry(filename string) error {
	return p.registry.Import(filename)
}
