package library

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/boreq/eggplant/logging"
	"github.com/boreq/eggplant/store"
	"github.com/pkg/errors"
)

type Id string

func (id Id) String() string {
	return string(id)
}

type Track struct {
	Id       string  `json:"id,omitempty"`
	Title    string  `json:"title,omitempty"`
	FileHash string  `json:"fileHash,omitempty"`
	Duration float64 `json:"duration,omitempty"`
}

type Album struct {
	Id    string `json:"id,omitempty"`
	Title string `json:"title,omitempty"`

	Parents []Album `json:"parents,omitempty"`
	Albums  []Album `json:"albums,omitempty"`
	Tracks  []Track `json:"tracks,omitempty"`
}

const rootDirectoryName = "Eggplant"

type track struct {
	title    string
	fileHash string
	path     string
}

func newTrack(path string) (*track, error) {
	_, title := filepath.Split(path)
	i := strings.LastIndex(title, ".")
	if i > 0 {
		title = title[:i]
	}

	h, err := getFileHash(path)
	if err != nil {
		return nil, err
	}
	t := &track{
		title:    title,
		fileHash: h,
		path:     path,
	}
	return t, nil
}

type directory struct {
	title       string
	directories map[Id]*directory
	tracks      map[Id]*track
}

func newDirectory(title string) *directory {
	return &directory{
		title:       title,
		directories: make(map[Id]*directory),
		tracks:      make(map[Id]*track),
	}
}

type Library struct {
	directory string
	root      *directory
	store     *store.Store
	log       logging.Logger
}

func Open(directory string, s *store.Store) (*Library, error) {
	l := &Library{
		log:       logging.New("library"),
		root:      newDirectory(rootDirectoryName),
		directory: directory,
		store:     s,
	}

	if err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			l.log.Debug("file", "name", info.Name(), "path", path)
			if err := l.addFile(path); err != nil {
				return errors.Wrap(err, "could not add a file")
			}
		}
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "walk failed")
	}
	l.store.SetTracks(l.List())
	return l, nil

}

func (l *Library) addFile(path string) error {
	relativePath, err := filepath.Rel(l.directory, path)
	if err != nil {
		return errors.Wrap(err, "could not get relative filepath")
	}

	dir, file := filepath.Split(relativePath)
	dirs := strings.Split(strings.Trim(dir, string(os.PathSeparator)), string(os.PathSeparator))

	directory, err := l.getOrCreateDirectory(dirs)
	if err != nil {
		return errors.Wrap(err, "could not get directory")
	}
	track, err := newTrack(path)
	if err != nil {
		return err
	}
	trackId, err := getHash(file)
	if err != nil {
		return err
	}
	directory.tracks[trackId] = track
	return nil
}

func (l *Library) getOrCreateDirectory(names []string) (*directory, error) {
	var dir *directory = l.root
	for _, name := range names {
		id, err := getHash(name)
		if err != nil {
			return nil, err
		}

		subdir, ok := dir.directories[id]
		if !ok {
			subdir = newDirectory(name)
			dir.directories[id] = subdir
		}
		dir = subdir
	}
	return dir, nil
}

func (l *Library) getDirectory(ids []Id) (*directory, error) {
	var dir *directory = l.root
	for _, id := range ids {
		subdir, ok := dir.directories[id]
		if !ok {
			return nil, errors.Errorf("subdirectory '%s' not found", id)
		}
		dir = subdir
	}
	return dir, nil
}

func (l *Library) Browse(ids []Id) (Album, error) {
	listed := Album{}

	//if len(parts) > 0 {
	//	listed.Name = parts[len(parts)-1]
	//}

	for i := 0; i < len(ids); i++ {
		parentIds := ids[:i+1]
		dir, err := l.getDirectory(parentIds)
		if err != nil {
			return Album{}, errors.Wrap(err, "failed to get directory")
		}
		parent := Album{
			Id:    parentIds[len(parentIds)-1].String(),
			Title: dir.title,
		}
		listed.Parents = append(listed.Parents, parent)
	}

	dir, err := l.getDirectory(ids)
	if err != nil {
		return Album{}, errors.Wrap(err, "failed to get directory")
	}

	listed.Title = dir.title

	for id, directory := range dir.directories {
		d := Album{
			Id:    id.String(),
			Title: directory.title,
		}
		listed.Albums = append(listed.Albums, d)
	}
	sort.Slice(listed.Albums, func(i, j int) bool { return listed.Albums[i].Title < listed.Albums[j].Title })

	for id, track := range dir.tracks {
		t := Track{
			Id:       id.String(),
			Title:    track.title,
			FileHash: track.fileHash,
			Duration: l.store.GetDuration(track.fileHash).Seconds(),
		}
		listed.Tracks = append(listed.Tracks, t)
	}
	sort.Slice(listed.Tracks, func(i, j int) bool { return listed.Tracks[i].Title < listed.Tracks[j].Title })

	return listed, nil
}

func (l *Library) List() []store.Track {
	m := make(map[string]store.Track)
	l.list(m, l.root)
	var tracks []store.Track
	for _, track := range m {
		tracks = append(tracks, track)
	}
	return tracks
}

func (l *Library) list(tracks map[string]store.Track, dir *directory) {
	for _, track := range dir.tracks {
		tracks[track.fileHash] = store.Track{
			Path: track.path,
			Id:   track.fileHash,
		}
	}
	for _, subdirectory := range dir.directories {
		l.list(tracks, subdirectory)
	}
}

func getHash(s string) (Id, error) {
	buf := bytes.NewBuffer([]byte(s))
	hasher := sha256.New()
	if _, err := io.Copy(hasher, buf); err != nil {
		return "", err
	}
	var sum []byte
	sum = hasher.Sum(sum)
	return hashToId(sum), nil
}

func getFileHash(p string) (string, error) {
	f, err := os.Open(p)
	if err != nil {
		return "", err
	}
	defer f.Close()
	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return "", err
	}
	var sum []byte
	sum = hasher.Sum(sum)
	return hex.EncodeToString(sum), nil
}

const idLength = 20

func hashToId(sum []byte) Id {
	return Id(hex.EncodeToString(sum)[:idLength])
}