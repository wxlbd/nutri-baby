package repository

import (
	"context"

	"github.com/wxlbd/nutri-baby-server/internal/domain/entity"
)

// UserRepository 用户仓储接口
type UserRepository interface {
	// Create 创建用户
	Create(ctx context.Context, user *entity.User) error
	// FindByOpenID 根据OpenID查找用户
	FindByOpenID(ctx context.Context, openID string) (*entity.User, error)
	// Update 更新用户
	Update(ctx context.Context, user *entity.User) error
	// UpdateLastLoginTime 更新最后登录时间
	UpdateLastLoginTime(ctx context.Context, openID string) error
	// UpdateDefaultBabyID 更新默认宝宝ID
	UpdateDefaultBabyID(ctx context.Context, openID string, babyID string) error
}

// InvitationRepository 邀请码仓储接口
type InvitationRepository interface {
	// Create 创建邀请码
	Create(ctx context.Context, invitation *entity.Invitation) error
	// FindByCode 根据邀请码查找
	FindByCode(ctx context.Context, code string) (*entity.Invitation, error)
	// Delete 删除邀请码
	Delete(ctx context.Context, code string) error
	// DeleteExpired 删除过期邀请码
	DeleteExpired(ctx context.Context) error
}

// BabyRepository 宝宝仓储接口 (去家庭化架构)
type BabyRepository interface {
	// Create 创建宝宝
	Create(ctx context.Context, baby *entity.Baby) error
	// FindByID 根据ID查找宝宝
	FindByID(ctx context.Context, babyID string) (*entity.Baby, error)
	// FindByUserID 查找用户可访问的宝宝列表(通过协作者关系)
	FindByUserID(ctx context.Context, openid string) ([]*entity.Baby, error)
	// FindByFamilyGroup 查找家庭分组下的宝宝列表
	FindByFamilyGroup(ctx context.Context, familyGroup string) ([]*entity.Baby, error)
	// Update 更新宝宝信息
	Update(ctx context.Context, baby *entity.Baby) error
	// Delete 删除宝宝(软删除)
	Delete(ctx context.Context, babyID string) error
	// FindByCreator 查找用户创建的宝宝列表
	FindByCreator(ctx context.Context, creatorID string) ([]*entity.Baby, error)
	// FindAll 查找所有宝宝
	FindAll(ctx context.Context) ([]*entity.Baby, error)
}
