package web

import "github.com/solunara/isb/src/repository/dao/article"

// VO view object，是对前端的

type ArticleVO struct {
	Id    int64  `json:"id"`
	Title string `json:"title"`
	// 摘要
	Abstract string `json:"abstract"`
	// 内容
	Content string `json:"content"`
	// 注意一点，状态这个东西，可以是前端来处理，也可以是后端处理
	// 0 -> unknown -> 未知状态
	// 1 -> 未发表，手机 APP 这种涉及到发版的问题，那么后端来处理
	// 涉及到国际化，也是后端来处理
	Status uint8  `json:"status"`
	Author string `json:"author"`
	Ctime  string `json:"ctime"`
	Utime  string `json:"utime"`
}

type ListReq struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type ArticleReq struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (req ArticleReq) toDomain(uid int64) article.Article {
	return article.Article{
		Id:       req.Id,
		Title:    req.Title,
		Content:  req.Content,
		AuthorId: uid,
	}
}
