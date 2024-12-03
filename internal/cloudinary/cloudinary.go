package cloudinary

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/admin"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/ftrbnd/film-sync/internal/util"
)

var cld *cloudinary.Cloudinary
var ctx context.Context

func SetCredentials() error {
	c, err := cloudinary.New()
	if err != nil {
		return err
	}

	c.Config.URL.Secure = true
	cld = c
	ctx = context.Background()

	log.Default().Println("[Cloudinary] Successfully set credentials")
	return nil
}

func CreateFolder(name string) (*admin.CreateFolderResult, error) {
	result, err := cld.Admin.CreateFolder(ctx, admin.CreateFolderParams{
		Folder: name,
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func UploadImage(folder string, path string) (string, error) {
	resp, err := cld.Upload.Upload(ctx, path, uploader.UploadParams{
		PublicID:       filepath.Base(path),
		UniqueFilename: api.Bool(false),
		Overwrite:      api.Bool(true),
		AssetFolder:    folder,
	})
	if err != nil {
		return "", err
	}

	log.Default().Printf("[Cloudinary] Uploaded %s!\n", path)
	return resp.PublicID, nil
}

func SetFolderName(old string, new string) error {
	_, err := cld.Admin.RenameFolder(ctx, admin.RenameFolderParams{
		FromPath: old,
		ToPath:   new,
	})

	if err != nil {
		return err
	}

	return nil
}

func FolderLink(name string) (string, error) {
	id, err := util.LoadEnvVar("CLOUDINARY_ID")
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("https://console.cloudinary.com/pm/%s/media-explorer/%s", id, name), nil
}
