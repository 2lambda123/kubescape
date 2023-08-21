package opaprocessor

import (
	"github.com/docker/distribution/reference"
)

func normalize_image_name(img string) (string, error) {
	name, err := reference.ParseNormalizedNamed(img)
	if err != nil {
		return "", err
	}
	return name.String(), nil
}
