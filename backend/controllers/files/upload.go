package files

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/Pauloo27/archvium/model"
	"github.com/Pauloo27/archvium/services/db"
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

	path := utils.Fmt("/%s/", c.Locals("user_name"))

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
			path += utils.Fmt("%s/", folder)
		}
	}

	if !utils.IsValidFileName(file.Filename) {
		return utils.AsError(c, http.StatusBadRequest, "Invalid file name "+file.Filename)
	}

	path += file.Filename

	sourceFile, err := file.Open()
	if err != nil {
		return utils.AsError(c, http.StatusInternalServerError, "Something went wrong while opening the source file")
	}

	dbFile := model.File{
		Path:    path,
		OwnerID: c.Locals("user_id").(int),
	}

	err = db.Connection.Save(&dbFile).Error
	if err != nil {
		if utils.IsNotUnique(err) {
			return utils.AsError(c, http.StatusConflict, "File alread exists (probably)")
		}
		return utils.AsError(c, http.StatusInternalServerError, "Something went wrong while storing the file in the db")
	}

	fmt.Println(utils.Fmt("%s/%d",
		c.Locals("ENV_STORAGE_ROOT").(string),
		dbFile.ID),
	)

	basePath := c.Locals("ENV_STORAGE_ROOT").(string)
	if !strings.HasSuffix(basePath, "/") {
		basePath += "/"
	}

	targetFile, err := os.Create(
		utils.Fmt("%s/%d",
			basePath,
			dbFile.ID),
	)
	if err != nil {
		return utils.AsError(c, http.StatusInternalServerError, "Something went wrong while opening the target file")
	}

	// TODO: check if written == size?
	_, err = io.Copy(targetFile, sourceFile)

	if err != nil {
		return utils.AsError(c, http.StatusInternalServerError, "Something went wronge while copying source to target file")
	}

	return utils.AsJSON(c, http.StatusCreated, fiber.Map{"id": dbFile.ID})
}
