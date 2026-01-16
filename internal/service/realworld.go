package service

import (
	"context"
	"fmt"
	pb "kratos-realworld/api/realworld/v1"
	"kratos-realworld/internal/biz"
	"kratos-realworld/internal/pkg/jwt"
	"strings"
	"time"

	kjwt "github.com/go-kratos/kratos/v2/middleware/auth/jwt"

	//"golang.org/x/crypto/bcrypt"
	"github.com/go-kratos/kratos/v2/errors"
	"google.golang.org/protobuf/types/known/emptypb"
)

type RealWorldService struct {
	uc  *biz.RealWorldUsecase
	jwt *jwt.JWTService
	pb.UnimplementedRealWorldServer
}

func NewRealWorldService(uc *biz.RealWorldUsecase, jwt *jwt.JWTService) *RealWorldService {
	return &RealWorldService{
		uc:  uc,
		jwt: jwt,
	}
}

func (s *RealWorldService) Login(ctx context.Context, req *pb.AuthRequest) (*pb.UserReply, error) {
	if req.User.Email == "" || req.User.Password == "" {
		return nil, errors.BadRequest("email or password didn't exist", "")
	}
	user, err := s.uc.Login(ctx, &biz.RealWorld{
		Email:    req.User.Email,
		Password: req.User.Password,
	})
	if err != nil {
		return nil, err
	}

	//这是登录成功后才进行token签发
	token, err := s.jwt.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}
	return &pb.UserReply{
		User: &pb.UserReply_User{
			Email: user.Email,
			Token: token,
		},
	}, nil
}
func (s *RealWorldService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.UserReply, error) {
	if req.User.Email == "" || req.User.Password == "" || req.User.Username == "" {
		return nil, errors.BadRequest("email, password, and username are required", "")
	} //数据完备性检测

	user, err := s.uc.Register(ctx, &biz.RealWorld{
		Email:    req.User.Email,
		Password: req.User.Password,
		UserName: req.User.Username,
	})
	if err != nil {
		return nil, err
	}
	return &pb.UserReply{
		User: &pb.UserReply_User{
			Email:    user.Email,
			Username: user.UserName,
		},
	}, nil
}
func (s *RealWorldService) GetCurrentUser(ctx context.Context, req *emptypb.Empty) (*pb.UserReply, error) {
	//获取当前用户的信息
	//中间级已经自动鉴权了，尝试从ctx拿到用户信息
	// 从 ctx 中获取 JWT claims
	claims, ok := kjwt.FromContext(ctx)
	if !ok {
		return nil, errors.Unauthorized("UNAUTHORIZED", "no jwt claims in context")
	}

	// claims 是 interface{} 类型，需要断言成你的自定义结构体或 MapClaims
	mapClaims := claims.(*jwt.CustomClaims)

	userID := mapClaims.UserID
	email := mapClaims.Email

	if userID <= 0 || email == "" {
		return nil, errors.BadRequest("info is not valied", "")
	}

	user, err := s.uc.GetCurrentUser(ctx, &biz.RealWorld{
		ID:    userID,
		Email: email,
	})
	if err != nil {
		return nil, err
	} else {
		//签发新token
		newtoken, err := s.jwt.GenerateToken(user.ID, user.Email)
		if err != nil {
			return nil, err
		}
		return &pb.UserReply{
			User: &pb.UserReply_User{
				Email:    user.Email,
				Token:    newtoken,
				Username: user.UserName,
				Bio:      user.Bio,
				Image:    user.Image,
			},
		}, nil
	}

	//s.logger.Infof("当前用户ID: %d, 邮箱: %s", userID, email)

}
func (s *RealWorldService) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UserReply, error) {
	//从context中拿到断言
	claims, ok := kjwt.FromContext(ctx)
	if !ok {
		return nil, errors.Unauthorized("UNAUTHORIZED", "no jwt claims in context")
	}
	mapClaims := claims.(*jwt.CustomClaims)

	userID := mapClaims.UserID
	email := mapClaims.Email
	if userID <= 0 || email == "" {
		return nil, errors.BadRequest("jwt no valied data", "")
	}
	user := &biz.RealWorld{
		ID:    userID,
		Email: email,
	}
	if req.User.Username != "" {
		user.UserName = req.User.Username
	}
	if req.User.Bio != "" {
		user.Bio = req.User.Bio
	}
	if req.User.Image != "" {
		user.Image = req.User.Image
	}
	//对数据进行判定是否存在

	//调用uc层的更新方法
	user, err := s.uc.UpdateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	return &pb.UserReply{
		User: &pb.UserReply_User{
			Username: user.UserName,
		},
	}, nil
}
func (s *RealWorldService) GetProfile(ctx context.Context, req *pb.GetProfileRequest) (*pb.ProfileReply, error) {
	//先鉴权拿请求方的id和email信息
	claims, ok := kjwt.FromContext(ctx)
	if !ok {
		return nil, errors.Unauthorized("UNAUTHORIZED", "no jwt claims in context")
	}
	mapClaims := claims.(*jwt.CustomClaims)

	userID := mapClaims.UserID
	email := mapClaims.Email
	if userID <= 0 || email == "" {
		return nil, errors.BadRequest("jwt no valied data", "")
	}
	if user, follow, err := s.uc.GetProfileByUserName(ctx, userID, req.Username); err != nil {
		return nil, err
	} else {

		return &pb.ProfileReply{
			Profile: &pb.ProfileReply_Profile{
				Username:  user.UserName,
				Bio:       user.Bio,
				Image:     user.Image,
				Following: *follow, //是否关注
			},
		}, nil
	}

}
func (s *RealWorldService) FollowUser(ctx context.Context, req *pb.FollowUserRequest) (*pb.ProfileReply, error) {
	//先鉴权拿请求方的id和email信息
	claims, ok := kjwt.FromContext(ctx)
	if !ok {
		return nil, errors.Unauthorized("UNAUTHORIZED", "no jwt claims in context")
	}
	mapClaims := claims.(*jwt.CustomClaims)

	userID := mapClaims.UserID
	email := mapClaims.Email
	if userID <= 0 || email == "" {
		return nil, errors.BadRequest("jwt no valied data", "")
	}
	user, err := s.uc.FollowUser(ctx, userID, req.Username)
	if err != nil {
		return nil, err
	}
	return &pb.ProfileReply{
		Profile: &pb.ProfileReply_Profile{
			Username:  user.UserName,
			Bio:       user.Bio,
			Image:     user.Image,
			Following: true, //是否关注
		},
	}, nil
}
func (s *RealWorldService) UnFollowUser(ctx context.Context, req *pb.FollowUserRequest) (*pb.ProfileReply, error) {
	//先鉴权拿请求方的id和email信息
	claims, ok := kjwt.FromContext(ctx)
	if !ok {
		return nil, errors.Unauthorized("UNAUTHORIZED", "no jwt claims in context")
	}
	mapClaims := claims.(*jwt.CustomClaims)

	userID := mapClaims.UserID
	email := mapClaims.Email
	if userID <= 0 || email == "" {
		return nil, errors.BadRequest("jwt no valied data", "")
	}
	user, err := s.uc.UnFollowUser(ctx, userID, req.Username)
	if err != nil {
		return nil, err
	}
	return &pb.ProfileReply{
		Profile: &pb.ProfileReply_Profile{
			Username:  user.UserName,
			Bio:       user.Bio,
			Image:     user.Image,
			Following: false, //是否关注
		},
	}, nil
}
func (s *RealWorldService) ListArticles(ctx context.Context, req *pb.ListArticlesRequest) (*pb.MultipleArticleReply, error) {
	return &pb.MultipleArticleReply{}, nil
}
func (s *RealWorldService) FeedArticles(ctx context.Context, req *pb.FeedArticlesRequest) (*pb.MultipleArticleReply, error) {
	return &pb.MultipleArticleReply{}, nil
}
func (s *RealWorldService) GetArticle(ctx context.Context, req *pb.GetArticleRequest) (*pb.SingleArticleReply, error) {
	return &pb.SingleArticleReply{}, nil
}
func (s *RealWorldService) CreateArticle(ctx context.Context, req *pb.CreateArticleRequest) (*pb.SingleArticleReply, error) {
	//从ctx中获取当前用户的id
	claims, ok := kjwt.FromContext(ctx)
	if !ok {
		return nil, errors.Unauthorized("UNAUTHORIZED", "no jwt claims in context")
	}
	mapClaims := claims.(*jwt.CustomClaims)

	userID := mapClaims.UserID
	email := mapClaims.Email
	if userID <= 0 || email == "" {
		return nil, errors.BadRequest("jwt no valied data", "")
	}
	//解析请求体中是否包含有效信息
	if req.Article.Body == "" || req.Article.Title == "" {
		return nil, errors.BadRequest("article not include valied data", "")
	}

	art, err := s.uc.CreateArticle(ctx, &biz.Article{
		AuthorID:    userID,
		Title:       req.Article.Title,
		Description: req.Article.Description,
		Body:        req.Article.Body,
		Slug:        GenerateSlug(req.Article.Title),
	}, &req.Article.TagList)

	if err != nil {
		return nil, err
	} else {
		return &pb.SingleArticleReply{
			Article: &pb.SingleArticleReply_Article{
				Slug:        art.Slug,
				Title:       art.Title,
				Description: art.Description,
				Body:        art.Body,
				TagList:     nil,
				//	CreatedAt: string(art.CreatedAt),
				//	UpdatedAt: art.UpdateAt,
				Favorited:      false,
				FavoritesCount: 0,
				//	Author: userID,
			},
		}, nil
	}
}
func (s *RealWorldService) UpdateArticle(ctx context.Context, req *pb.UpdateArticleRequest) (*pb.SingleArticleReply, error) {
	//先从断言中拿到ID
	claims, ok := kjwt.FromContext(ctx)
	if !ok {
		return nil, errors.Unauthorized("UNAUTHORIZED", "no jwt claims in context")
	}
	mapClaims := claims.(*jwt.CustomClaims)

	userID := mapClaims.UserID
	email := mapClaims.Email
	if userID <= 0 || email == "" {
		return nil, errors.BadRequest("jwt no valied data", "")
	}

  	art ,err := s.uc.UpdateArticle(ctx,&biz.Article{
		AuthorID:    userID,
		Title:       req.Article.Title,
		Description: req.Article.Description,
		Body:        req.Article.Body,
		Slug:        req.Slug,
	})
	if err != nil{
		return nil,err
	}
	return &pb.SingleArticleReply{
		Article: &pb.SingleArticleReply_Article{
			Slug:        art.Slug,
			Title:       art.Title,
			Description: art.Description,
			Body:        art.Body,
			TagList:     nil,
			//	CreatedAt: string(art.CreatedAt),
			//	UpdatedAt: art.UpdateAt,
			Favorited:      false,
			FavoritesCount: 0,
			//	Author: userID,
		},
	}, nil
}
func (s *RealWorldService) DeleteArticle(ctx context.Context, req *pb.DeleteArticleRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
func (s *RealWorldService) AddComments(ctx context.Context, req *pb.AddCommentsRequest) (*pb.SingleCommentReply, error) {
	return &pb.SingleCommentReply{}, nil
}
func (s *RealWorldService) GetComments(ctx context.Context, req *pb.GetCommentsRequest) (*pb.MultipleCommentReply, error) {
	return &pb.MultipleCommentReply{}, nil
}
func (s *RealWorldService) DeleteComment(ctx context.Context, req *pb.DeleteCommentRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
func (s *RealWorldService) FavoriteArticle(ctx context.Context, req *pb.FavoriteArticleRequest) (*pb.SingleArticleReply, error) {
	return &pb.SingleArticleReply{}, nil
}
func (s *RealWorldService) UnFavoriteArticle(ctx context.Context, req *pb.FavoriteArticleRequest) (*pb.SingleArticleReply, error) {
	return &pb.SingleArticleReply{}, nil
}
func (s *RealWorldService) GetTags(ctx context.Context, req *emptypb.Empty) (*pb.ListTagsReply, error) {
	return &pb.ListTagsReply{}, nil
}

func GenerateSlug(title string) string {
	base := strings.ToLower(strings.ReplaceAll(title, " ", "-"))
	unique := fmt.Sprintf("%s-%d", base, time.Now().UnixNano())
	return unique
}
