package storage

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

const stableReadAttempts = 3

var ErrUnstableRead = errors.New("file changed while reading stable snapshot")

type Snapshot struct {
	Data       []byte
	Revision   string
	Mode       os.FileMode
	Size       int64
	ModifiedAt time.Time
}

func ReadSnapshot(path string) (Snapshot, error) {
	for attempt := 0; attempt < stableReadAttempts; attempt++ {
		snapshot, stable, err := readSnapshotOnce(path)
		if err != nil {
			return Snapshot{}, err
		}
		if stable {
			return snapshot, nil
		}
	}
	return Snapshot{}, ErrUnstableRead
}

func readSnapshotOnce(path string) (Snapshot, bool, error) {
	entryBefore, err := os.Lstat(path)
	if err != nil {
		return Snapshot{}, false, err
	}
	if !entryBefore.Mode().IsRegular() {
		return Snapshot{}, false, fmt.Errorf("snapshot target is not a regular file")
	}

	file, err := os.Open(path)
	if err != nil {
		return Snapshot{}, false, err
	}
	openedBefore, err := file.Stat()
	if err != nil {
		file.Close()
		return Snapshot{}, false, err
	}
	data, readErr := io.ReadAll(file)
	openedAfter, statErr := file.Stat()
	closeErr := file.Close()
	if readErr != nil {
		return Snapshot{}, false, readErr
	}
	if statErr != nil {
		return Snapshot{}, false, statErr
	}
	if closeErr != nil {
		return Snapshot{}, false, closeErr
	}
	entryAfter, err := os.Lstat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Snapshot{}, false, nil
		}
		return Snapshot{}, false, err
	}

	stable := os.SameFile(entryBefore, openedBefore) &&
		os.SameFile(openedBefore, openedAfter) &&
		os.SameFile(openedAfter, entryAfter) &&
		openedBefore.Size() == openedAfter.Size() &&
		openedBefore.ModTime() == openedAfter.ModTime() &&
		int64(len(data)) == openedAfter.Size()
	return Snapshot{
		Data:       data,
		Revision:   Revision(data),
		Mode:       openedAfter.Mode().Perm(),
		Size:       openedAfter.Size(),
		ModifiedAt: openedAfter.ModTime().UTC(),
	}, stable, nil
}
