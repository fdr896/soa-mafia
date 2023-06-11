package filesystem

import (
	"errors"
	"os"
	"path"
)

const (
	PDF_FOLDER = "pdf"
)

type PdfStorage struct {
	folder string
}

func CreatePdfStorage() (*PdfStorage, error) {
	ps := PdfStorage{
		folder: path.Join(BASE_FOLDER, PDF_FOLDER),
	}

	if err := os.MkdirAll(ps.folder, 0744); err != nil {
		return nil, err
	}

	return &ps, nil
}

func (ps *PdfStorage) Exists(username string) bool {
	path := ps.UserPdfPath(username)

	 _, err := os.Stat(path)

	 return !errors.Is(err, os.ErrNotExist)
}

func (ps *PdfStorage) CreateUserFolder(username string) error {
	userFolder := path.Join(ps.folder, username)

	if err := os.RemoveAll(userFolder); err != nil {
		return err
	}
	return os.MkdirAll(userFolder, 0744)
}

func (ps *PdfStorage) UserPdfPath(username string) string {
	return path.Join(ps.folder, username, "pdf.pdf")
}
