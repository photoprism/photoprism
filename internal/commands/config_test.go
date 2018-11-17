package commands

import (
	"testing"

	"github.com/photoprism/photoprism/internal/test"
	"github.com/stretchr/testify/assert"
)

func TestConfigCommand(t *testing.T) {
	var err error

	ctx := test.CliContext()

	output := test.Capture(func() {
		err = ConfigCommand.Run(ctx)
	})

	expected := `NAME                  VALUE
debug                 false
config-file           /go/src/github.com/photoprism/photoprism/configs/photoprism.yml
darktable-cli         /usr/bin/darktable-cli
originals-path        /go/src/github.com/photoprism/photoprism/assets/testdata/originals
import-path           /srv/photoprism/photos/import
export-path           /srv/photoprism/photos/export
cache-path            /srv/photoprism/cache
assets-path           /go/src/github.com/photoprism/photoprism/assets
database-driver       mysql
database-dsn          photoprism:photoprism@tcp(database:3306)/photoprism?parseTime=true
`

	assert.Equal(t, expected, output)
	assert.Nil(t, err)
}
