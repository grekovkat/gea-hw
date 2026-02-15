package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidation(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		offset      int64
		expectedErr error
	}{
		{name: "path not exist", path: "testdata/", offset: 0, expectedErr: ErrUnsupportedFile},
		{name: "offset to large", path: "testdata/input.txt", offset: 1000_000_000, expectedErr: ErrOffsetExceedsFileSize},
		{name: "offset to large", path: "testdata/input_1.txt", offset: 0, expectedErr: os.ErrNotExist},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Validation(test.path, test.offset)

			require.ErrorIs(t, err, test.expectedErr)
		})
	}
}

func TestTempFiles(t *testing.T) {
	tmpPattern := "/tmp/tempToFile-*"
	fileFrom := "testdata/input.txt"
	fileTo := "input_copy.txt"

	t.Run("no temp file", func(t *testing.T) {
		// подсчет временных файлов до копирования.
		filesBefore, _ := filepath.Glob(tmpPattern)

		err := Copy(fileFrom, fileTo, 0, 0)

		filesAfter, _ := filepath.Glob(tmpPattern)

		require.NoError(t, err)
		require.Equal(t, len(filesBefore), len(filesAfter))

		os.Remove(fileTo)
	})
}
