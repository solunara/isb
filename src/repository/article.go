package repository

import (
	"context"

	"github.com/solunara/isb/src/repository/dao/article"
)

type ArticleRepository interface {
	Create(ctx context.Context, art article.Article) (int64, error)
	Update(ctx context.Context, art article.Article) error
}

type CachedArticleRepository struct {
	dao article.ArticleDAO
}

func NewArticleRepository(dao article.ArticleDAO) ArticleRepository {
	return &CachedArticleRepository{
		dao: dao,
	}
}

func (c *CachedArticleRepository) Create(ctx context.Context, art article.Article) (int64, error) {
	return c.dao.Insert(ctx, article.Article{
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.AuthorId,
	})
}

func (c *CachedArticleRepository) Update(ctx context.Context, art article.Article) error {
	return c.dao.UpdateById(ctx, article.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.AuthorId,
	})
}
