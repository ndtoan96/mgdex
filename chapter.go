// Package mgdex provides interfaces to get information as well as download chapters and manga from mangadex.
package mgdex

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/ndtoan96/imgdl"
)

// A ChapterData represents data of chapter gotten from mangadex api. It does not
// include all possible information, only the ones commonly used.
type ChapterData struct {
	Data struct {
		Id         string
		Attributes struct {
			Volume             string
			Chapter            string
			Title              string
			Hash               string
			Data               []string
			DataSaver          []string
			TranslatedLanguage string
		}
		Relationships []map[string]interface{}
	}
}

type serverData struct {
	BaseUrl string
}

// GetChapter send request to mangadex api and get back chapter data.
func GetChapter(id string) (*ChapterData, error) {
	// Request chapter via api
	url := fmt.Sprintf("https://api.mangadex.org/chapter/%v", id)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("error getting %v, %v", url, resp.Status))
	}

	// Deserialize chapter json response to struct
	var chapter ChapterData
	err = json.NewDecoder(resp.Body).Decode(&chapter)
	if err != nil {
		return nil, err
	}
	return &chapter, nil
}

// Volume returns volume number of chapter, default is empty string.
func (chapter ChapterData) Volume() string {
	return chapter.Data.Attributes.Volume
}

// Chapter returns chapter number of chapter, default is empty string.
func (chapter ChapterData) Chapter() string {
	return chapter.Data.Attributes.Chapter
}

// Title returns title of chapter, default is empty string.
func (chapter ChapterData) Title() string {
	return chapter.Data.Attributes.Title
}

// ScanlationGroup returns scanlation group of chapter. Note that this requires
// an additional parameter in chapter request and the function GetChapter does not
// implements it. So this function is only useful with ChapterData gotten from
// manga query where includeGroup is enabled.
func (chapter ChapterData) ScanlationGroup() string {
	for _, rel := range chapter.Data.Relationships {
		if rel["type"].(string) == "scanlation_group" {
			return rel["attributes"].(map[string]interface{})["name"].(string)
		}
	}
	return ""
}

// GetPageUrls returns urls for all pages in the chapter.
func (chapter ChapterData) GetPageUrls(dataSaver bool) ([]string, error) {
	// Get base url
	serverUrl := fmt.Sprintf("https://api.mangadex.org/at-home/server/%v", chapter.Data.Id)
	resp, err := http.Get(serverUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("error getting %v: %v", serverUrl, resp.Status))
	}
	var server serverData
	err = json.NewDecoder(resp.Body).Decode(&server)
	if err != nil {
		return nil, err
	}

	// Construct images urls
	var quality string
	var database []string
	if dataSaver {
		quality = "data-saver"
		database = chapter.Data.Attributes.DataSaver
	} else {
		quality = "data"
		database = chapter.Data.Attributes.Data
	}
	var urls []string
	for _, img := range database {
		urls = append(urls, fmt.Sprintf("%v/%v/%v/%v", server.BaseUrl, quality, chapter.Data.Attributes.Hash, img))
	}
	return urls, nil
}

// Download downloads the chapter and save to folder specified by 'path'.
// If 'path' is empty, current folder will be used.
func (chapter ChapterData) Download(dataSaver bool, path string) error {
	// Get urls
	urls, err := chapter.GetPageUrls(dataSaver)
	if err != nil {
		return err
	}

	// Download images
	err = imgdl.DownloadImgages(urls, filepath.Join(path, "page_"))
	if err != nil {
		return err
	}
	return nil
}

// DownloadAsZip downloads the chapter and save to zip file specified by 'path'.
func (chapter ChapterData) DownloadAsZip(dataSaver bool, path string) error {
	// Get page urls
	urls, err := chapter.GetPageUrls(dataSaver)
	if err != nil {
		return err
	}

	// Download images to zip
	err = imgdl.DownloadImagesZip(urls, path, "page_")
	if err != nil {
		return err
	}

	return nil
}
