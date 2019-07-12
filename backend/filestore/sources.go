package filestore

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"

	cmp "github.com/snetsystems/cmp/backend"
)

// SrcExt is the the file extension searched for in the directory for source files
const SrcExt = ".src"

var _ cmp.SourcesStore = &Sources{}

// Sources are JSON sources stored in the filesystem
type Sources struct {
	Dir     string                                      // Dir is the directory containing the sources.
	Load    func(string, interface{}) error             // Load loads string name and dashbaord passed in as interface
	Create  func(string, interface{}) error             // Create will write source to file.
	ReadDir func(dirname string) ([]os.FileInfo, error) // ReadDir reads the directory named by dirname and returns a list of directory entries sorted by filename.
	Remove  func(name string) error                     // Remove file
	IDs     cmp.ID                                      // IDs generate unique ids for new sources
	Logger  cmp.Logger
}

// NewSources constructs a source store wrapping a file system directory
func NewSources(dir string, ids cmp.ID, logger cmp.Logger) cmp.SourcesStore {
	return &Sources{
		Dir:     dir,
		Load:    load,
		Create:  create,
		ReadDir: ioutil.ReadDir,
		Remove:  os.Remove,
		IDs:     ids,
		Logger:  logger,
	}
}

func sourceFile(dir string, source cmp.Source) string {
	base := fmt.Sprintf("%s%s", source.Name, SrcExt)
	return path.Join(dir, base)
}

// All returns all sources from the directory
func (d *Sources) All(ctx context.Context) ([]cmp.Source, error) {
	files, err := d.ReadDir(d.Dir)
	if err != nil {
		return nil, err
	}

	sources := []cmp.Source{}
	for _, file := range files {
		if path.Ext(file.Name()) != SrcExt {
			continue
		}
		var source cmp.Source
		if err := d.Load(path.Join(d.Dir, file.Name()), &source); err != nil {
			var fmtErr = fmt.Errorf("Error loading source configuration from %v:\n%v", path.Join(d.Dir, file.Name()), err)
			d.Logger.Error(fmtErr)
			continue // We want to load all files we can.
		} else {
			sources = append(sources, source)
		}
	}
	return sources, nil
}

// Add creates a new source within the directory
func (d *Sources) Add(ctx context.Context, source cmp.Source) (cmp.Source, error) {
	genID, err := d.IDs.Generate()
	if err != nil {
		d.Logger.
			WithField("component", "source").
			Error("Unable to generate ID")
		return cmp.Source{}, err
	}

	id, err := strconv.Atoi(genID)
	if err != nil {
		d.Logger.
			WithField("component", "source").
			Error("Unable to convert ID")
		return cmp.Source{}, err
	}

	source.ID = id

	file := sourceFile(d.Dir, source)
	if err = d.Create(file, source); err != nil {
		if err == cmp.ErrSourceInvalid {
			d.Logger.
				WithField("component", "source").
				WithField("name", file).
				Error("Invalid Source: ", err)
		} else {
			d.Logger.
				WithField("component", "source").
				WithField("name", file).
				Error("Unable to write source:", err)
		}
		return cmp.Source{}, err
	}
	return source, nil
}

// Delete removes a source file from the directory
func (d *Sources) Delete(ctx context.Context, source cmp.Source) error {
	_, file, err := d.idToFile(source.ID)
	if err != nil {
		return err
	}

	if err := d.Remove(file); err != nil {
		d.Logger.
			WithField("component", "source").
			WithField("name", file).
			Error("Unable to remove source:", err)
		return err
	}
	return nil
}

// Get returns a source file from the source directory
func (d *Sources) Get(ctx context.Context, id int) (cmp.Source, error) {
	board, file, err := d.idToFile(id)
	if err != nil {
		if err == cmp.ErrSourceNotFound {
			d.Logger.
				WithField("component", "source").
				WithField("name", file).
				Error("Unable to read file")
		} else if err == cmp.ErrSourceInvalid {
			d.Logger.
				WithField("component", "source").
				WithField("name", file).
				Error("File is not a source")
		}
		return cmp.Source{}, err
	}
	return board, nil
}

// Update replaces a source from the file system directory
func (d *Sources) Update(ctx context.Context, source cmp.Source) error {
	board, _, err := d.idToFile(source.ID)
	if err != nil {
		return err
	}

	if err := d.Delete(ctx, board); err != nil {
		return err
	}
	file := sourceFile(d.Dir, source)
	return d.Create(file, source)
}

// idToFile takes an id and finds the associated filename
func (d *Sources) idToFile(id int) (cmp.Source, string, error) {
	// Because the entire source information is not known at this point, we need
	// to try to find the name of the file through matching the ID in the source
	// content with the ID passed.
	files, err := d.ReadDir(d.Dir)
	if err != nil {
		return cmp.Source{}, "", err
	}

	for _, f := range files {
		if path.Ext(f.Name()) != SrcExt {
			continue
		}
		file := path.Join(d.Dir, f.Name())
		var source cmp.Source
		if err := d.Load(file, &source); err != nil {
			return cmp.Source{}, "", err
		}
		if source.ID == id {
			return source, file, nil
		}
	}

	return cmp.Source{}, "", cmp.ErrSourceNotFound
}
