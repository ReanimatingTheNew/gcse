package main

import (
	"fmt"
	"github.com/daviddengcn/gcse"
	"github.com/daviddengcn/go-villa"
	"strings"
)

type StatItem struct {
	Name    string
	Package string
	Link    string // no package, specify a link
	Info    string
}
type StatList struct {
	Name  string
	Info  string
	Items []StatItem
}

type TopN struct {
	cmp villa.CmpFunc
	pq  *villa.PriorityQueue
	n   int
}

func NewTopN(cmp villa.CmpFunc, n int) *TopN {
	return &TopN{
		cmp: cmp,
		pq:  villa.NewPriorityQueue(cmp),
		n:   n,
	}
}

func (t *TopN) Append(item interface{}) {
	if t.pq.Len() < t.n {
		t.pq.Push(item)
	} else if t.cmp(t.pq.Peek(), item) < 0 {
		t.pq.Pop()
		t.pq.Push(item)
	}
}

func (t *TopN) PopAll() []interface{} {
	lst := make([]interface{}, t.pq.Len())
	for i := range lst {
		lst[len(lst)-i-1] = t.pq.Pop()
	}

	return lst
}

func (t *TopN) Len() int {
	return t.pq.Len()
}

func statTops(N int) []StatList {
	if indexDB == nil {
		return nil
	}

	topStaticScores := NewTopN(func(a, b interface{}) int {
		return villa.FloatValueCompare(a.(gcse.HitInfo).StaticScore,
			b.(gcse.HitInfo).StaticScore)
	}, N)

	topImported := NewTopN(func(a, b interface{}) int {
		return villa.IntValueCompare(len(a.(gcse.HitInfo).Imported),
			len(b.(gcse.HitInfo).Imported))
	}, N)

	sites := make(map[string]int)

	indexDB.Search(nil, func(docID int32, data interface{}) error {
		hit := data.(gcse.HitInfo)
		hit.Name = packageShowName(hit.Name, hit.Package)

		topStaticScores.Append(hit)
		topImported.Append(hit)

		host := strings.ToLower(gcse.HostOfPackage(hit.Package))
		if host != "" {
			sites[host] = sites[host] + 1
		}

		return nil
	})

	tlStaticScore := StatList{
		Name:  "Hot",
		Info:  "refs stars",
		Items: make([]StatItem, 0, topStaticScores.Len()),
	}
	for _, item := range topStaticScores.PopAll() {
		hit := item.(gcse.HitInfo)
		tlStaticScore.Items = append(tlStaticScore.Items, StatItem{
			Name:    hit.Name,
			Package: hit.Package,
			Info:    fmt.Sprintf("%d %d", len(hit.Imported), hit.StarCount),
		})
	}

	tlImported := StatList{
		Name:  "Most Imported",
		Info:  "refs",
		Items: make([]StatItem, 0, topImported.Len()),
	}
	for _, item := range topImported.PopAll() {
		hit := item.(gcse.HitInfo)
		tlImported.Items = append(tlImported.Items, StatItem{
			Name:    hit.Name,
			Package: hit.Package,
			Info:    fmt.Sprintf("%d", len(hit.Imported)),
		})
	}

	topSites := NewTopN(func(a, b interface{}) int {
		return villa.IntValueCompare(sites[a.(string)], sites[b.(string)])
	}, N)
	for site := range sites {
		topSites.Append(site)
	}
	tlSites := StatList{
		Name:  "Sites",
		Info:  "packages",
		Items: make([]StatItem, 0, topSites.Len()),
	}
	for _, st := range topSites.PopAll() {
		site := st.(string)
		cnt := sites[site]
		tlSites.Items = append(tlSites.Items, StatItem{
			Name: site,
			Link: "http://" + site,
			Info: fmt.Sprintf("%d", cnt),
		})
	}

	return []StatList{
		tlStaticScore, tlImported, tlSites,
	}
}
