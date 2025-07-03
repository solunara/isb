package model

import "time"

type ArticleStatus uint8

const (
	// ArticleStatusUnknown 为了避免零值之类的问题
	ArticleStatusUnknown ArticleStatus = iota
	ArticleStatusUnpublished
	ArticleStatusPublished
	ArticleStatusPrivate
)

// Article 可以同时表达线上库和制作库的概念吗？
// 可以同时表达，作者眼中的 Article 和读者眼中的 Article 吗？
type Article struct {
	Id      int64
	Title   string
	Content string
	// Author 要从用户来
	Author Author
	Status ArticleStatus
	Ctime  time.Time
	Utime  time.Time
}

// Author 在帖子这个领域内，是一个值对象
type Author struct {
	Id   string
	Name string
}
