package mgdex

import (
	"log"
	"time"
)

const (
	// This constant specifies the limit number of image downloading at the same time. Since
	// too many downloads happen at a time can lead to timeout or network error.
	PAGE_LIMIT = 200
)

// Method Download downloads list of chapters. They will be named with format <prefix><chapter_number>.
// 'prefix' can have parent folder, it will be created if not exist.
func (chapters ChapterList) Download(dataSaver bool, prefix string) {
	commonBatchDownload(chapters, dataSaver, prefix, false, "")
}

// Method Download downloads list of chapters and zip them. They will be named with
// format <prefix><chapter_number>.<ext>. 'prefix' can have parent folder, it will be created if not exist.
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
		database := chapter.GetPageNames(dataSaver)

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
				err = chapter.DownloadAsZip(dataSaver, prefix+chapter.GetChapter()+"."+ext)
			} else {
				err = chapter.Download(dataSaver, prefix+chapter.GetChapter())
			}
			if err == nil {
				println("Chapter " + chapter.GetChapter() + " downloaded.")
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
