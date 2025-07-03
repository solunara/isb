package article

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

func NewGORMArticleDAO(db *gorm.DB) ArticleDAO {
	return &GORMArticleDAO{
		db: db,
	}
}

type GORMArticleDAO struct {
	db *gorm.DB
}

func (dao *GORMArticleDAO) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := dao.db.WithContext(ctx).Create(&art).Error
	return art.Id, err
}

func (dao *GORMArticleDAO) UpdateById(ctx context.Context,
	art Article) error {
	now := time.Now().UnixMilli()
	res := dao.db.Model(&Article{}).WithContext(ctx).
		Where("id=? AND author_id = ? ", art.Id, art.AuthorId).
		Updates(map[string]any{
			"title":   art.Title,
			"content": art.Content,
			"utime":   now,
		})
	err := res.Error
	if err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return errors.New("更新数据失败")
	}
	return nil
}

// Article 这是制作库的
type Article struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 长度 1024
	Title   string `gorm:"type=varchar(1024)"`
	Content string `gorm:"type=BLOB"`

	// 在 author_id 上创建索引
	AuthorId int64 `gorm:"index"`
	//AuthorId int64 `gorm:"index=aid_ctime"`
	//Ctime    int64 `gorm:"index=aid_ctime"`
	Ctime int64
	Utime int64
}
