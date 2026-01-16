package data

import (
	"context"
	"errors"
	"fmt"
	"time"

	"kratos-realworld/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type RealWorldRepo struct {
	data *Data
	log  *log.Helper
}

// NewRealWorldRepo .
func NewRealWorldRepo(data *Data, logger log.Logger) biz.RealWorldRepo {
	return &RealWorldRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *RealWorldRepo) Save(ctx context.Context, g *biz.RealWorld) (*biz.RealWorld, error) {
	return g, nil
}

func (r *RealWorldRepo) Update(ctx context.Context, g *biz.RealWorld) (*biz.RealWorld, error) {
	return g, nil
}

func (r *RealWorldRepo) FindByID(ctx context.Context, id int64) (*biz.RealWorld, error) {
	var user biz.RealWorld

	// 使用 GORM 的 WithContext，防止阻塞和支持 trace
	res := r.data.DB.WithContext(ctx).First(&user, id)

	// 没找到记录
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		r.log.Infof("user not found, id=%d", id)
		return nil, nil
	}

	// 发生其他数据库错误
	if res.Error != nil {
		r.log.Errorf("FindByID error: %v", res.Error)
		return nil, res.Error
	}

	// 返回查询到的用户
	return &user, nil
}

func (r *RealWorldRepo) UpdateUser(ctx context.Context, user *biz.RealWorld) (*biz.RealWorld, error) {
	updateData := map[string]interface{}{}
	if user.UserName != "" {
		updateData["username"] = user.UserName
	}
	if user.Bio != "" {
		updateData["bio"] = user.Bio
	}
	if user.Image != "" {
		updateData["image"] = user.Image
	}

	// 没有要更新的字段直接返回
	if len(updateData) == 0 {
		return nil, fmt.Errorf("no data need be changed")
	}

	// GORM 更新（自动 WHERE id = ?）
	res := r.data.DB.WithContext(ctx).
		Model(&biz.RealWorld{}).
		Where("id = ?", user.ID).
		Updates(updateData)

	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, errors.New("no user updated")
	}
	return user, nil
}

func (r *RealWorldRepo) FindByUserName(ctx context.Context, username string) (*biz.RealWorld, error) {
	var user biz.RealWorld

	res := r.data.DB.WithContext(ctx).
		Where("username = ?", username).
		First(&user)

	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		r.log.Infof("user not found, username=%s", username)
		return nil, nil
	}

	if res.Error != nil {
		r.log.Errorf("FindByUserName error: %v", res.Error)
		return nil, res.Error
	}

	return &user, nil
}

func (r *RealWorldRepo) FindAFollowB(ctx context.Context, myid int64, otherid int64) (bool, error) {
	var count int64

	err := r.data.DB.WithContext(ctx).
		Table("follows").
		Where("follower_id = ? AND followee_id = ?", myid, otherid).
		Count(&count).Error

	if err != nil {
		r.log.Errorf("FindAFollowB query failed: %v", err)
		return false, err
	}

	return count > 0, nil
}

func (r *RealWorldRepo) AFollowB(ctx context.Context, myid int64, otherid int64) error {
	// 如果 A 关注自己，直接返回错误
	if myid == otherid {
		return fmt.Errorf("cannot follow yourself")
	}

	// 先检查是否已经关注过
	var count int64
	if err := r.data.DB.WithContext(ctx).
		Table("follows").
		Where("follower_id = ? AND followee_id = ?", myid, otherid).
		Count(&count).Error; err != nil {
		r.log.Errorf("AFollowB check follow error: %v", err)
		return err
	}

	if count > 0 {
		// 已经关注了，不重复插入
		r.log.Infof("user %d already follows user %d", myid, otherid)
		return nil
	}

	// 插入新的关注记录
	follow := map[string]interface{}{
		"follower_id": myid,
		"followee_id": otherid,
		"created_at":  time.Now(),
	}

	if err := r.data.DB.WithContext(ctx).
		Table("follows").
		Create(&follow).Error; err != nil {
		r.log.Errorf("AFollowB insert error: %v", err)
		return err
	}

	r.log.Infof("user %d followed user %d successfully", myid, otherid)
	return nil
}

func (r *RealWorldRepo) AUnFollowB(ctx context.Context, myid int64, otherid int64) error {
	// 不允许取关自己
	if myid == otherid {
		return fmt.Errorf("cannot unfollow yourself")
	}

	// 检查是否存在关注记录
	var count int64
	if err := r.data.DB.WithContext(ctx).
		Table("follows").
		Where("follower_id = ? AND followee_id = ?", myid, otherid).
		Count(&count).Error; err != nil {
		r.log.Errorf("AUnFollowB check follow error: %v", err)
		return err
	}

	if count == 0 {
		// 没有关注记录，无需取关
		r.log.Infof("user %d has not followed user %d", myid, otherid)
		return nil
	}

	// 删除关注记录（执行取关操作）
	if err := r.data.DB.WithContext(ctx).
		Table("follows").
		Where("follower_id = ? AND followee_id = ?", myid, otherid).
		Delete(nil).Error; err != nil {
		r.log.Errorf("AUnFollowB delete error: %v", err)
		return err
	}

	r.log.Infof("user %d unfollowed user %d successfully", myid, otherid)
	return nil
}

func (r *RealWorldRepo) ListByHello(context.Context, string) ([]*biz.RealWorld, error) {
	return nil, nil
}

func (r *RealWorldRepo) ListAll(context.Context) ([]*biz.RealWorld, error) {
	return nil, nil
}

// 根据邮箱查找用户
func (r *RealWorldRepo) FindByEmail(ctx context.Context, email string) (*biz.RealWorld, error) {
	var user biz.RealWorld
	res := r.data.DB.WithContext(ctx).Where("email = ?", email).First(&user)

	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if res.Error != nil {
		r.log.Errorf("FindByEmail error: %v", res.Error)
		return nil, res.Error
	}

	return &user, nil
}

// 创建用户
func (r *RealWorldRepo) CreateUser(ctx context.Context, g *biz.RealWorld) (*biz.RealWorld, error) {
	res := r.data.DB.WithContext(ctx).Create(g)
	if res.Error != nil {
		r.log.Errorf("CreateUser error: %v", res.Error)
		return nil, res.Error
	}
	return g, nil
}

func (r *RealWorldRepo) SetUserOnline(ctx context.Context, id int64) error {
	key := fmt.Sprintf("user:online:%d", id)
	if err := r.data.RDB.Set(ctx, key, "true", 30*time.Minute).Err(); err != nil {
		return err
	}
	return nil
}

func (r *RealWorldRepo) CreateArticle(ctx context.Context, art *biz.Article) (*biz.Article, error) {
	res := r.data.DB.WithContext(ctx).Create(art)
	if res.Error != nil {
		r.log.Errorf("CreateTag error: %v", res.Error)
		return nil, res.Error
	}
	return art, nil //data层还未实现
}
func (r *RealWorldRepo) CreateTag(ctx context.Context, tag *biz.Tags) error {
	// 使用 FirstOrCreate 检查标签是否已经存在
	res := r.data.DB.WithContext(ctx).FirstOrCreate(tag, biz.Tags{Name: tag.Name})

	// 如果发生其他错误，则返回错误
	if res.Error != nil {
		r.log.Errorf("CreateTag error: %v", res.Error)
		return res.Error
	}

	// 如果没有错误，则说明标签已经存在或已经成功创建
	return nil
}

func (r *RealWorldRepo) CreateTags(ctx context.Context, tags *[]biz.Tags) error {
	//变量tags，查看是否已经出现
	var end_err error
	for i := 0; i < len(*tags); i++ {
		if err := r.CreateTag(ctx, &(*tags)[i]); err != nil {
			end_err = err
		}
	}

	return end_err //data层还未实现
}

func (r *RealWorldRepo) GetArticleBySlug(ctx context.Context, slug string) (*biz.Article, error) {
	var art biz.Article
	res := r.data.DB.WithContext(ctx).Where("slug = ?", slug).First(&art)
	if res.Error != nil {
		return nil, res.Error
	}
	return &art, nil
}
func (r *RealWorldRepo) UpdateArticle(ctx context.Context, up *biz.Article) (*biz.Article, error) {
	upData := map[string]interface{}{}
	if up.ID == 0 {
		return nil, fmt.Errorf("cant found the atrticle id")
	}
	if up.Title != "" {
		upData["title"] = up.Title
	}
	if up.Body != "" {
		upData["body"] = up.Body
	}
	if up.Description != "" {
		upData["description"] = up.Description
	}

	if len(upData) == 0 {
		return nil, fmt.Errorf("no data need update")
	}

	//return nil, fmt.Errorf("ceshi")
	res := r.data.DB.WithContext(ctx).Model(&biz.Article{}).Where("id = ?", up.ID).Updates(upData)

	if res.Error != nil {
		return nil, res.Error
	}

	if res.RowsAffected == 0 {
		return nil, errors.New("no article updated")
	}
	// 重新查询更新后的文章
	var updatedArticle biz.Article
	if err := r.data.DB.WithContext(ctx).First(&updatedArticle, up.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch updated article: %v", err)
	}

	return &updatedArticle, nil
}
