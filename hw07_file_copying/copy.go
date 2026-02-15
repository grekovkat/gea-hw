package main

import (
	"errors"
	"fmt"
	"io"
	"os"
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
)

func checkPath(path string) (size int64, isDir bool, err error) {
	data, err := os.Stat(path)
	if err != nil {
		return 0, false, err
	}

	return data.Size(), data.IsDir(), nil
}

func Validation(fromPath string, offset int64) (fromSize int64, err error) {
	fromSize, fromIsDir, err := checkPath(fromPath)
	if err != nil {
		return 0, fmt.Errorf("проверка пути %s: %w", fromPath, err)
	}

	if fromIsDir {
		return 0, fmt.Errorf("проверка файла %v:%w", fromIsDir, ErrUnsupportedFile)
	}

	if fromSize < offset {
		return 0, fmt.Errorf("проверка отступа:%w", ErrOffsetExceedsFileSize)
	}

	return fromSize, nil
}

type ProgressReader struct {
	reader  io.Reader
	current int64
	total   int64
}

func (pr *ProgressReader) Read(p []byte) (n int, err error) {
	n, err = pr.reader.Read(p)
	pr.current += int64(n)

	percent := float64(pr.current) * 100 / float64(pr.total)
	fmt.Println("Прогресс: ", percent, "%")

	return n, err
}

func Copy(fromPath, toPath string, offset, limit int64) error {
	// Валидация аргументов.
	fromSize, err := Validation(fromPath, offset)
	if err != nil {
		return fmt.Errorf("валидация аргументов: %w", err)
	}

	// Открытие файла.
	fromFile, err := os.OpenFile(fromPath, os.O_RDONLY, 0)
	if err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("доступ к файлу: %w", err)
		}
		return fmt.Errorf("попытка открыть файл: %w", err)
	}
	defer fromFile.Close()

	// временный файл.
	tempFile, err := os.CreateTemp("", "tempToFile-*")
	if err != nil {
		return fmt.Errorf("создание временного файла: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Установка сдвига.
	if offset > 0 {
		_, err = fromFile.Seek(offset, io.SeekStart)
		if err != nil {
			return fmt.Errorf("установка сдвига: %w", err)
		}
	}

	// копирование.
	if limit == 0 {
		limit = fromSize
	}
	copySize := min(fromSize-offset, limit)

	prReader := &ProgressReader{
		reader: fromFile,
		total:  copySize,
	}
	written, err := io.CopyN(tempFile, prReader /*fromFile*/, copySize)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("копирование данных: %w", err)
	}

	fmt.Printf("\nСкопировано %d байт\n", written)

	err = tempFile.Sync()
	if err != nil {
		return fmt.Errorf("синхронизация: %w", err)
	}

	err = os.Rename(tempFile.Name(), toPath)
	if err != nil {
		return fmt.Errorf("перемещение файла: %w", err)
	}

	return nil
}
