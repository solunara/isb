package service

import (
	"context"

	"github.com/solunara/isb/src/repository"
	"github.com/solunara/isb/src/repository/dao/article"
)

type ArticleService interface {
	Save(ctx context.Context, art article.Article) (int64, error)
}

type articleService struct {
	repo repository.ArticleRepository
}

func NewArticleService(repo repository.ArticleRepository) ArticleService {
	return &articleService{
		repo: repo,
	}
}

func (a *articleService) Save(ctx context.Context, art article.Article) (int64, error) {
	if art.Id > 0 {
		err := a.repo.Update(ctx, art)
		return art.Id, err
	}
	return a.repo.Create(ctx, art)
}
