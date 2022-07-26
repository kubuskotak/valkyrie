package valkyrie

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	ConfigOpts struct {
		Config    interface{}
		Paths     []string
		Filenames []string
	}
)

func Config(opts ConfigOpts) error {
	for _, p := range opts.Paths {
		filePath := filepath.Join(p, ".env")
		// load env from file
		if _, fileErr := os.Stat(filePath); fileErr == nil {
			filePath := fmt.Sprintf("%s/%s", p, ".env")
			// Set ENV for development
			_ = cleanenv.ReadConfig(filePath, opts.Config)
		}
	}
	var err error
	for _, f := range opts.Filenames {
		for _, p := range opts.Paths {
			filePath := filepath.Join(p, f)
			if _, fileErr := os.Stat(filePath); fileErr != nil {
				return fileErr
			}
			if err = cleanenv.ReadConfig(filePath, opts.Config); err != nil {
				return err
			}
		}
	}

	return err
}
