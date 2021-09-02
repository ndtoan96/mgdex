package mgdex

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

type MangaData struct{ Results []ChapterData }

type mangaQuery struct {
	id           string
	lang         string
	limit        int
	offset       int
	order        string
	includeGroup bool
}

func MangaQuery(id string) *mangaQuery {
	return &mangaQuery{
		id:           id,
		lang:         "en",
		limit:        100,
		offset:       0,
		order:        "asc",
		includeGroup: false,
	}
}

func (q mangaQuery) Language(lang string) *mangaQuery {
	q.lang = lang
	return &q
}

func (q mangaQuery) Limit(limit int) *mangaQuery {
	q.limit = limit
	return &q
}

func (q mangaQuery) Offset(offset int) *mangaQuery {
	q.offset = offset
	return &q
}

func (q mangaQuery) Order(order string) *mangaQuery {
	q.order = order
	return &q
}

func (q mangaQuery) IncludeScanlationGroup() *mangaQuery {
	q.includeGroup = true
	return &q
}

func (q mangaQuery) verify() error {
	if q.lang == "" {
		return errors.New("language is empty")
	}
	if q.limit < 1 || q.limit > 500 {
		return errors.New("limit is not in range [1..500]")
	}
	if q.offset < 0 {
		return errors.New("offset is negative")
	}
	if q.order != "asc" && q.order != "desc" {
		return errors.New(`expect order to be "asc" or "desc", found "` + q.order + `"`)
	}
	return nil
}

func (q mangaQuery) GetManga() (*MangaData, error) {
	err := q.verify()
	if err != nil {
		return nil, err
	}

	base, err := url.Parse(fmt.Sprintf("https://api.mangadex.org/manga/%v/feed", q.id))
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("translatedLanguage[]", q.lang)
	params.Add("limit", fmt.Sprint(q.limit))
	params.Add("offset", fmt.Sprint(q.offset))
	params.Add("order[chapter]", q.order)
	if q.includeGroup {
		params.Add("includes[]", "scanlation_group")
	}
	base.RawQuery = params.Encode()

	resp, err := http.Get(base.String())
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error getting %v, %v", base, err))
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("error getting %v, %v", base, resp.Status))
	}

	var manga MangaData
	err = json.NewDecoder(resp.Body).Decode(&manga)
	if err != nil {
		return nil, err
	}

	return &manga, nil
}
