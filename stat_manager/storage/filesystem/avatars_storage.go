package filesystem

import (
	"io"
	"os"
	"path"
)

const (
	AVATARS_FOLDER = "avatars"

	AVATAR_DEFAULT = "default_avatar.png"
	AVATAR_DEFAULT_PATH = "stat_manager/storage/filesystem/data/default_avatar.png"
)

type AvatarsStorage struct {
	folder string

	avatars map[string]interface{}
}

func CreateAvatarsStorage() (*AvatarsStorage, error) {
	as := AvatarsStorage{
		folder: path.Join(BASE_FOLDER, AVATARS_FOLDER),
		avatars: make(map[string]interface{}),
	}

	if err := os.MkdirAll(as.folder, 0744); err != nil {
		return nil, err
	}

	if err := as.copyDefaultAvatarToFolder(); err != nil {
		return nil, err
	}

	return &as, nil
}

func (as *AvatarsStorage) WriteUserAvatar(username string, contentType string, avatarFile io.Reader) (string, error) {
	if err := as.createUserFolder(username); err != nil {
		return "", err
	}

	imageExt, err := parseContentType(contentType)
	if err != nil {
		return "", err
	}

	avatarPath := path.Join(username, "avatar" + imageExt)

	return avatarPath, as.copyFileToFolder(avatarPath, avatarFile)
}

func GetDefaultAvatar() string {
	return AVATAR_DEFAULT
}

func (as *AvatarsStorage) GetAvatarPath(avatarPath string) string {
	return as.filePath(avatarPath)
}

func parseContentType(contentType string) (string, error) {
	switch contentType {
	case "image/jpeg":
		return ".jpeg", nil
	case "image/png":
		return ".png", nil
	default:
		return "", ErrUnsupportedContentType
	}
}

func (as *AvatarsStorage) filePath(filename string) string {
	return path.Join(as.folder, filename)
}

func (as *AvatarsStorage) createUserFolder(username string) error {
	userFolder := path.Join(as.folder, username)

	if err := os.RemoveAll(userFolder); err != nil {
		return err
	}
	return os.MkdirAll(userFolder, 0744)
}

func (as *AvatarsStorage) copyFileToFolder(filename string, file io.Reader) error {
	dest, err := os.Create(as.filePath(filename))
	if err != nil {
		return err
	}

	_, err = io.Copy(dest, file)
	if err != nil {
		return err
	}
	as.avatars[filename] = struct{}{}

	return err
}

func (as *AvatarsStorage) copyDefaultAvatarToFolder() error {
	defaultAvatar, err := os.Open(AVATAR_DEFAULT_PATH)
	if err != nil {
		return err
	}

	return as.copyFileToFolder(AVATAR_DEFAULT, defaultAvatar)
}
