package web

import (
	"fmt"
	"math"
	"strconv"
)

const (
	defaultPageIndex int = 1
	defaultPerPage   int = 10
)

type Pagination struct {
	Url       string
	ItemCount int
	PageIndex int
	PerPage   int
	PageCount int
}

type Page struct {
	Index  int
	Active bool
}

func getPageCount(items, perPage int) int {
	return int((float64(items) + float64(perPage) - 1) / float64(perPage))
}

func NewPaginationWithStrings(url string, items int, page, perPage string) *Pagination {
	pageInt, err := strconv.Atoi(page)
	if err != nil {
		pageInt = defaultPageIndex
	}

	perPageInt, err := strconv.Atoi(perPage)
	if err != nil {
		perPageInt = defaultPerPage
	}

	return NewPagination(url, items, pageInt, perPageInt)
}

func NewPagination(url string, items, page, perPage int) *Pagination {
	pCount := getPageCount(items, perPage)

	if page < 1 || items == 0 {
		page = 1
	} else if page > pCount {
		page = pCount
	}

	return &Pagination{
		Url:       url,
		ItemCount: items,
		PageIndex: page,
		PerPage:   perPage,
		PageCount: pCount,
	}
}

func (p *Pagination) GetCurrentPages() []*Page {
	pages := []*Page{}

	for i := 1; i <= p.PageCount; i++ {
		newPage := &Page{Index: i, Active: false}
		if i == p.PageIndex {
			newPage.Active = true
			pages = append(pages, newPage)
		} else if i >= p.PageIndex-2 && i <= p.PageIndex+2 {
			pages = append(pages, newPage)
		}
	}

	return pages
}

// Function to get the 1st and last indexes using the current pagination
func (p *Pagination) GetSliceNumbers() (int, int) {
	return (p.PageIndex - 1) * p.PerPage, int(math.Min(float64(p.ItemCount), float64(p.PageIndex*p.PerPage)))
}

func (p *Pagination) GetPerPages() []int {
	return []int{10, 25, 50, 100}
}

func (p *Pagination) Href(pageIndex, perPage int) string {
	if pageIndex == 0 {
		pageIndex = p.PageIndex
	}
	if perPage == 0 {
		perPage = p.PerPage
	}
	return fmt.Sprintf("%s?page=%d&per_page=%d", p.Url, pageIndex, perPage)
}
