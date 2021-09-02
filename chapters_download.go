package mgdex

import (
	"log"
	"time"
)

const (
	PAGE_LIMIT = 200
)

func (chapters ChapterList) Download(dataSaver bool, prefix string) {
	commonBatchDownload(chapters, dataSaver, prefix, false, "")
}

func (chapters ChapterList) DownloadAsZip(dataSaver bool, prefix string, ext string) {
	commonBatchDownload(chapters, dataSaver, prefix, true, ext)
}

func commonBatchDownload(chapters ChapterList, dataSaver bool, prefix string, zip bool, ext string) {
	if len(chapters) == 0 {
		log.Println("Chapter list is empty")
	}
	page_cnt := 0
	c_cnt := 0
	c := make(chan error)
	delay := len(chapters) > 40
	for _, chapter := range chapters {
		var database []string
		if dataSaver {
			database = chapter.Data.Attributes.DataSaver
		} else {
			database = chapter.Data.Attributes.Data
		}

		if page_cnt+len(database) > PAGE_LIMIT {
			page_cnt = 0
			for c_cnt > 0 {
				err := <-c
				if err != nil {
					log.Println(err)
				}
				c_cnt--
			}
		}

		page_cnt = page_cnt + len(database)
		go func(chapter *ChapterData) {
			var err error
			if zip {
				if ext == "" {
					ext = "zip"
				}
				err = chapter.DownloadAsZip(dataSaver, prefix+chapter.Chapter()+"."+ext)
			} else {
				err = chapter.Download(dataSaver, prefix+chapter.Chapter())
			}
			if err == nil {
				println("Chapter " + chapter.Chapter() + " downloaded.")
			}
			c <- err
		}(chapter)
		if delay {
			time.Sleep(1500 * time.Millisecond)
		}
		c_cnt++
	}
	for c_cnt > 0 {
		err := <-c
		if err != nil {
			log.Println(err)
		}
		c_cnt--
	}
}
