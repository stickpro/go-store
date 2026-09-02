package console

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/stickpro/go-store/internal/config"
	"github.com/urfave/cli/v3"
)

func prepareMediaCommands(_, _ string) []*cli.Command {
	return []*cli.Command{
		{
			Name: "cache-clear",
			Description: "delete generated image resize variants (storage/public/images/*_<size>.{webp,jpg}); " +
				"they regenerate on the next request",
			Flags: []cli.Flag{cfgPathsFlag()},
			Action: func(_ context.Context, cl *cli.Command) error {
				conf, err := loadConfig(cl.Args().Slice(), cl.StringSlice("configs"))
				if err != nil {
					return fmt.Errorf("failed to load config: %w", err)
				}

				dir := filepath.Join(conf.FileStorage.Path, "public", "images")
				removed := 0
				for _, size := range presetSizes(conf) {
					for _, ext := range []string{"webp", "jpg"} {
						matches, _ := filepath.Glob(filepath.Join(dir, fmt.Sprintf("*_%d.%s", size, ext)))
						for _, m := range matches {
							if err := os.Remove(m); err == nil {
								removed++
							}
						}
					}
				}
				fmt.Printf("removed %d variant files from %s\n", removed, dir)
				return nil
			},
		},
	}
}

func presetSizes(conf *config.Config) []int {
	list := conf.Images.Presets
	if len(list) == 0 {
		list = config.DefaultImagePresets()
	}
	sizes := make([]int, 0, len(list))
	for _, p := range list {
		if p.Size > 0 {
			sizes = append(sizes, p.Size)
		}
	}
	return sizes
}
