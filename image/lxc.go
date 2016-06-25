package image

func LxcCreate(name, src string) (Image, error) {
	var i Image

	i.Name = name
	i.Type = TypeLXC
	i.Source = src

	// TODO: Check existance of LXC template

	return i, nil
}
