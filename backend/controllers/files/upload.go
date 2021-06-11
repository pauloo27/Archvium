package files

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/Pauloo27/archvium/utils"
	"github.com/gofiber/fiber/v2"
)

func Upload(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil || file == nil {
		return utils.AsError(c, http.StatusBadRequest, "Missing upload file")
	}

	maxFileSize := c.Locals("ENV_MAX_FILE_SIZE").(int64)

	if file.Size > maxFileSize {
		return utils.AsError(c, http.StatusBadRequest, utils.Fmt("Max file size is %d bytes", maxFileSize))
	}

	fullPath := utils.Fmt("/%s/", c.Locals("user_name"))

	folderTree := c.FormValue("target_folder")

	if folderTree != "" {
		var folderList []string
		err = json.Unmarshal([]byte(folderTree), &folderList)
		if err != nil {
			return utils.AsError(c, http.StatusBadRequest, "target_folder needes to be a json string array")
		}
		for _, folder := range folderList {
			if !utils.IsWord(folder) {
				return utils.AsError(c, http.StatusBadRequest, "Invalid folder name "+folder)
			}
			fullPath += utils.Fmt("%s/", folder)
		}
	}

	if !utils.IsValidFileName(file.Filename) {
		return utils.AsError(c, http.StatusBadRequest, "Invalid file name "+file.Filename)
	}

	foldersOnlyPath := fullPath
	fullPath += file.Filename

	sourceFile, err := file.Open()
	if err != nil {
		return utils.AsError(c, http.StatusInternalServerError, "Something went wrong while opening the source file")
	}

	basePath := utils.WithoutSlashSuffix(c.Locals("ENV_STORAGE_ROOT").(string))

	err = os.MkdirAll(basePath+foldersOnlyPath, 0700)

	if err != nil {
		return utils.AsError(c, http.StatusInternalServerError, "Something went wrong while creating folders")
	}

	realPath := utils.Fmt("%s/%s", basePath, fullPath)
	targetFile, err := os.Create(realPath)
	if err != nil {
		return utils.AsError(c, http.StatusInternalServerError, "Something went wrong while opening the target file")
	}

	// TODO: check if written == size?
	_, err = io.Copy(targetFile, sourceFile)

	if err != nil {
		return utils.AsError(c, http.StatusInternalServerError, "Something went wronge while copying source to target file")
	}

	info, err := utils.GetFileInfo(realPath)
	if err != nil {
		return utils.AsError(c, http.StatusInternalServerError, "Something went wrong while getting file info")
	}
	return utils.AsJSON(c, http.StatusCreated, *info)
}
