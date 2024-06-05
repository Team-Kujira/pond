package pond

func (p *Pond) ListRegistry() error {
	return p.registry.List()
}

func (p *Pond) UpdateRegistry(name string, args map[string]string) error {
	err := p.registry.Update(name, args)
	if err != nil {
		return err
	}

	return p.UpdateCodes()
}

func (p *Pond) ExportRegistry(filename string) error {
	return p.registry.Export(filename)
}

func (p *Pond) ImportRegistry(filename string) error {
	err := p.registry.Import(filename)
	if err != nil {
		return err
	}

	return p.UpdateCodes()
}
