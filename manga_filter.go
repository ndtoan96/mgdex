package mgdex

import (
	"sort"
	"strconv"
	"strings"
)

type ChapterList []*ChapterData

// Filter criterias of chapters in a manga. These criterias act in an AND manner.
type mangaFilter struct {
	manga        *MangaData
	volumes      map[string]struct{}
	chapters     map[string]struct{}
	volumeRange  *[2]float64
	chapterRange *[2]float64
	preferGroups map[string]int
}

// Filter returns a pointer to mangaFilter with default values. It works
// in a builder manner.
func (manga MangaData) Filter() *mangaFilter {
	return &mangaFilter{
		manga:        &manga,
		volumes:      nil,
		chapters:     nil,
		volumeRange:  nil,
		chapterRange: nil,
		preferGroups: nil,
	}
}

// Volumes specifies list of volumes.
func (filter mangaFilter) Volumes(vols []string) *mangaFilter {
	filter.volumes = make(map[string]struct{})
	for _, vol := range vols {
		filter.volumes[vol] = struct{}{}
	}
	return &filter
}

// Chapters specifies list of chapters.
func (filter mangaFilter) Chapters(chaps []string) *mangaFilter {
	filter.chapters = make(map[string]struct{})
	for _, chap := range chaps {
		filter.chapters[chap] = struct{}{}
	}
	return &filter
}

// VolumeRange specifies an inclusive range of volume
func (filter mangaFilter) VolumeRange(rng [2]float64) *mangaFilter {
	filter.volumeRange = &rng
	return &filter
}

// ChapterRange specifies an inclusive range of chapter
func (filter mangaFilter) ChapterRange(rng [2]float64) *mangaFilter {
	filter.chapterRange = &rng
	return &filter
}

// PreferGroups specifies the priority of each group in the order they are presented.
// Only takes effect if mangaQuery.IncludeScanlationGroup is enabled.
//
// Note that this does not filter only the chapter translated by these groups. In case
// there are several version of a chapter then the groups specifed here will take precedence
// when filtered.
func (filter mangaFilter) PreferGroups(groups []string) *mangaFilter {
	filter.preferGroups = make(map[string]int)
	for i, group := range groups {
		filter.preferGroups[strings.ToLower(group)] = len(groups) - i
	}
	return &filter
}

// GetChapters returns list of chapter sastified the criterias.
func (filter mangaFilter) GetChapters() (chapters ChapterList) {
	chapterMap := make(map[string]*ChapterData)
	for i, chapter := range filter.manga.Results {
		old_chapter, exist := chapterMap[chapter.Chapter()]
		if exist {
			if filter.preferGroups != nil {
				old_group := strings.ToLower(old_chapter.ScanlationGroup())
				new_group := strings.ToLower(chapter.ScanlationGroup())
				if filter.preferGroups[old_group] < filter.preferGroups[new_group] {
					chapterMap[chapter.Chapter()] = &filter.manga.Results[i]
				}
			}
			continue
		}
		isGood := true
		if filter.volumes != nil {
			_, exist := filter.volumes[chapter.Volume()]
			isGood = isGood && exist
		}
		if filter.chapters != nil {
			_, exist := filter.chapters[chapter.Chapter()]
			isGood = isGood && exist
		}
		if filter.volumeRange != nil {
			val, err := strconv.ParseFloat(chapter.Data.Attributes.Volume, 64)
			isGood = isGood && err == nil && val >= filter.volumeRange[0] && val <= filter.volumeRange[1]
		}
		if filter.chapterRange != nil {
			val, err := strconv.ParseFloat(chapter.Chapter(), 64)
			isGood = isGood && err == nil && val >= filter.chapterRange[0] && val <= filter.chapterRange[1]
		}
		if isGood {
			chapterMap[chapter.Chapter()] = &filter.manga.Results[i]
		}
	}
	for _, value := range chapterMap {
		chapters = append(chapters, value)
	}
	sort.Slice(chapters, func(i, j int) bool {
		chapter_i, _ := strconv.ParseFloat(chapters[i].Chapter(), 64)
		chapter_j, _ := strconv.ParseFloat(chapters[j].Chapter(), 64)
		return chapter_i < chapter_j
	})
	return
}
