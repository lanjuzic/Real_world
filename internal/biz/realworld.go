package biz

import (
	"context"
	//"fmt"

	v1 "kratos-realworld/api/realworld/v1"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrUserNotFound is user not found.
	ErrUserNotFound = errors.NotFound(v1.ErrorReason_USER_NOT_FOUND.String(), "user not found")
)

// RealWorld is a RealWorld model.
// RealWorld 用户模型
type RealWorld struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Email     string    `gorm:"uniqueIndex;size:120;not null" json:"email"`
	Password  string    `gorm:"column:password_hash;size:255;not null"`
	UserName  string    `gorm:"column:username;size:50;not null" json:"username"` // ✅ 修复关键点
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	Bio       string    `gorm:"column:bio;" json:"bio"`
	Image     string    `gorm:"column:image;" json:"image"`
}

type Article struct{
	ID int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Slug string `gorm:"size:255;unqueIndex;not null" json:"slug"`
	Title string `gorm:"size:255;not null" json:"title"`
 	Description string `gorm:"type:text" json:"description"`
 	Body  string  `gorm:"type:text;not null" json:"body"`
	AuthorID int64 `gorm:"not null" json:"author_id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdateAt time.Time  `gorm:"autoUpdateTIme" json:"update_at"`
}

type Tags struct{
	ID int64 `gorm:"primaryKey;autoIncrement" json:"id"`
	Name string `gorm:"not null" json:"name"`
}

func(Article) TableName()string{
	return "articles"
}

func (RealWorld) TableName() string {
	return "users"
}

// RealWorldRepo is a Greater repo.
type RealWorldRepo interface {
	//Save(context.Context, *RealWorld) (*RealWorld, error)
	//Update(context.Context, *RealWorld) (*RealWorld, error)
	//FindByID(context.Context, int64) (*RealWorld, error)
	FindByEmail(context.Context, string) (*RealWorld, error)
	CreateUser(context.Context, *RealWorld) (*RealWorld, error)
	SetUserOnline(context.Context, int64) error
	FindByID(context.Context, int64) (*RealWorld, error)
	FindByUserName(context.Context, string) (*RealWorld, error)
	UpdateUser(context.Context, *RealWorld) (*RealWorld, error)
	FindAFollowB(context.Context, int64, int64) (bool, error)
	AFollowB(context.Context, int64, int64) error
	AUnFollowB(context.Context, int64, int64) error
	//ListByHello(context.Context, string) ([]*RealWorld, error)
	//ListAll(context.Context) ([]*RealWorld, error)
}

// RealWorldUsecase is a RealWorld usecase.
type RealWorldUsecase struct {
	repo RealWorldRepo
	log  *log.Helper
}

// NewRealWorldUsecase new a RealWorld usecase.
func NewRealWorldUsecase(repo RealWorldRepo, logger log.Logger) *RealWorldUsecase {
	return &RealWorldUsecase{repo: repo, log: log.NewHelper(logger)}
}

// CreateRealWorld creates a RealWorld, and returns the new RealWorld.
func (uc *RealWorldUsecase) CreateRealWorld(ctx context.Context, g *RealWorld) (*RealWorld, error) {
	uc.log.WithContext(ctx).Infof("CreateRealWorld: %v", g.UserName)
	return nil, nil
	//return uc.repo.Save(ctx, g)
}

func (uc *RealWorldUsecase) Login(ctx context.Context, g *RealWorld) (*RealWorld, error) {
	//查询用户是否存在
	if user, err := uc.repo.FindByEmail(ctx, g.Email); err != nil {
		return nil, err
	} else {
		//检验密码是否正确
		if CheckPasswordHash(g.Password, user.Password) {
			//密码正确
			//此时用户应该事在线状态了 调repo层的接口往redis里面记录数据
			if err := uc.repo.SetUserOnline(ctx, user.ID); err != nil {
				return nil, err
			}
			return user, nil
		} else {
			//密码错误
			return nil, errors.Unauthorized("password is wrong", "")
		}
	}
}

func (uc *RealWorldUsecase) Register(ctx context.Context, g *RealWorld) (*RealWorld, error) {
	//查找用户是否已经存在 repo层
	if user, err := uc.repo.FindByEmail(ctx, g.Email); err != nil {
		return nil, err
	} else if user != nil {
		//用户存在返回错误
		return nil, errors.Conflict("user already exist", "")
	} else {
		//用户不存在，且没有错误
		//新建一个用户
		//密码加密
		if password, err := HashPassword(g.Password); err != nil {
			return nil, err
		} else {
			g.Password = password //转换成哈希值
			if user, err := uc.repo.CreateUser(ctx, g); err != nil {
				return nil, err
			} else {
				return user, nil //新建用户成功
			}
		}
	}
}

func (uc *RealWorldUsecase) GetCurrentUser(ctx context.Context, g *RealWorld) (*RealWorld, error) {
	user, err := uc.repo.FindByID(ctx, g.ID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (uc *RealWorldUsecase) UpdateUser(ctx context.Context, g *RealWorld) (*RealWorld, error) {
	user, err := uc.repo.FindByID(ctx, g.ID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.Conflict("user is not extit", "")
	}
	user_now, err := uc.repo.UpdateUser(ctx, g)
	if err != nil {
		return nil, err
	}
	return user_now, nil
}

func (uc *RealWorldUsecase) GetProfileByUserName(ctx context.Context, myid int64, username string) (*RealWorld, *bool, error) {
	//当前用户的id，被查看者名字
	user_be, err := uc.repo.FindByUserName(ctx, username)
	if err != nil {
		return nil, nil, err
	}
	if user_be == nil {
		return nil, nil, errors.Conflict("username is not found", "")
	}
	//查找myid是否关注user_beid
	isfollow, err := uc.repo.FindAFollowB(ctx, myid, user_be.ID)
	if err != nil {
		return nil, nil, err
	}

	return user_be, &isfollow, nil

}

func (uc *RealWorldUsecase) FollowUser(ctx context.Context, myid int64, username string) (*RealWorld, error) {
	//当前用户的id，被查看者名字
	user_be, err := uc.repo.FindByUserName(ctx, username)
	if err != nil {
		return nil, err
	}
	if user_be == nil {
		return nil, errors.Conflict("username is not found", "")
	}
	//查找myid是否关注user_beid
	isfollow, err := uc.repo.FindAFollowB(ctx, myid, user_be.ID)
	if err != nil {
		return nil, err
	}
	if isfollow { //已经关注了
		return user_be, nil
	} else { //还没有关注
		//提供二者id进行关注
		if err := uc.repo.AFollowB(ctx, myid, user_be.ID); err != nil {
			return nil, err
		} else {
			return user_be, nil
		}

	}
}

func (uc *RealWorldUsecase) UnFollowUser(ctx context.Context, myid int64, username string) (*RealWorld, error) {
	//当前用户的id，被查看者名字
	user_be, err := uc.repo.FindByUserName(ctx, username)
	if err != nil {
		return nil, err
	}
	if user_be == nil {
		return nil, errors.Conflict("username is not found", "")
	}
	//查找myid是否关注user_beid
	isfollow, err := uc.repo.FindAFollowB(ctx, myid, user_be.ID)
	if err != nil {
		return nil, err
	}
	if !isfollow { //没有关注
		return user_be, nil
	} else { //还没有关注
		//提供二者id进行关注
		if err := uc.repo.AUnFollowB(ctx, myid, user_be.ID); err != nil {
			return nil, err
		} else {
			return user_be, nil
		}

	}
}

func (uc *RealWorldUsecase) CreateArticle(ctx context.Context, art *Article, tags *[]string)(*Article,*[]string ,error){

	return nil,nil,nil
}

// 查找用户是否已经存在 repo层
// 验证密码
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// HashPassword 加密明文密码
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
