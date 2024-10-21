package transformer_test

import (
	"os"
	"path"
	"testing"

	"github.com/spf13/afero"
	"go.nhat.io/aferocopy/v2"
)

func SnapshotFs(t *testing.T, id string, fs afero.Fs) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	err = aferocopy.Copy("", path.Join("snapshot", id), aferocopy.Options{
		SrcFs:             fs,
		DestFs:            afero.NewBasePathFs(afero.NewOsFs(), wd),
		PermissionControl: aferocopy.AddPermission(0o777),
	})
	if err != nil {
		t.Fatal(err)
	}
}
