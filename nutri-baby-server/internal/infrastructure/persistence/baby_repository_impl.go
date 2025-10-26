package persistence

import (
	"context"

	"gorm.io/gorm"

	"github.com/wxlbd/nutri-baby-server/internal/domain/entity"
	"github.com/wxlbd/nutri-baby-server/internal/domain/repository"
	"github.com/wxlbd/nutri-baby-server/pkg/errors"
)

// babyRepositoryImpl 宝宝仓储实现 (去家庭化架构)
type babyRepositoryImpl struct {
	db *gorm.DB
}

// NewBabyRepository 创建宝宝仓储
func NewBabyRepository(db *gorm.DB) repository.BabyRepository {
	return &babyRepositoryImpl{db: db}
}

// Create 创建宝宝
func (r *babyRepositoryImpl) Create(ctx context.Context, baby *entity.Baby) error {
	if err := r.db.WithContext(ctx).Create(baby).Error; err != nil {
		return errors.Wrap(errors.DatabaseError, "failed to create baby", err)
	}
	return nil
}

// FindByID 根据ID查找宝宝
func (r *babyRepositoryImpl) FindByID(ctx context.Context, babyID string) (*entity.Baby, error) {
	var baby entity.Baby
	err := r.db.WithContext(ctx).
		Where("baby_id = ? AND deleted_at IS NULL", babyID).
		First(&baby).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New(errors.NotFound, "baby not found")
	}
	if err != nil {
		return nil, errors.Wrap(errors.DatabaseError, "failed to find baby", err)
	}

	return &baby, nil
}

// FindByUserID 查找用户可访问的宝宝列表(通过协作者关系)
func (r *babyRepositoryImpl) FindByUserID(ctx context.Context, openid string) ([]*entity.Baby, error) {
	var babies []*entity.Baby

	// 通过 baby_collaborators 表关联查询
	err := r.db.WithContext(ctx).
		Joins("JOIN baby_collaborators ON baby_collaborators.baby_id = babies.baby_id").
		Where("baby_collaborators.openid = ? AND baby_collaborators.deleted_at IS NULL AND babies.deleted_at IS NULL", openid).
		Order("baby_collaborators.join_time DESC").
		Find(&babies).Error

	if err != nil {
		return nil, errors.Wrap(errors.DatabaseError, "failed to find babies by user id", err)
	}

	return babies, nil
}

// FindByFamilyGroup 查找家庭分组下的宝宝列表
func (r *babyRepositoryImpl) FindByFamilyGroup(ctx context.Context, familyGroup string) ([]*entity.Baby, error) {
	var babies []*entity.Baby

	err := r.db.WithContext(ctx).
		Where("family_group = ? AND deleted_at IS NULL", familyGroup).
		Order("create_time DESC").
		Find(&babies).Error

	if err != nil {
		return nil, errors.Wrap(errors.DatabaseError, "failed to find babies by family group", err)
	}

	return babies, nil
}

// Update 更新宝宝信息
func (r *babyRepositoryImpl) Update(ctx context.Context, baby *entity.Baby) error {
	err := r.db.WithContext(ctx).
		Model(&entity.Baby{}).
		Where("baby_id = ? AND deleted_at IS NULL", baby.BabyID).
		Updates(baby).Error

	if err != nil {
		return errors.Wrap(errors.DatabaseError, "failed to update baby", err)
	}

	return nil
}

// Delete 删除宝宝(软删除)
func (r *babyRepositoryImpl) Delete(ctx context.Context, babyID string) error {
	err := r.db.WithContext(ctx).
		Model(&entity.Baby{}).
		Where("baby_id = ?", babyID).
		Update("deleted_at", gorm.Expr("CURRENT_TIMESTAMP")).Error

	if err != nil {
		return errors.Wrap(errors.DatabaseError, "failed to delete baby", err)
	}

	return nil
}

// FindByCreator 查找用户创建的宝宝列表
func (r *babyRepositoryImpl) FindByCreator(ctx context.Context, creatorID string) ([]*entity.Baby, error) {
	var babies []*entity.Baby

	err := r.db.WithContext(ctx).
		Where("creator_id = ? AND deleted_at IS NULL", creatorID).
		Order("create_time DESC").
		Find(&babies).Error

	if err != nil {
		return nil, errors.Wrap(errors.DatabaseError, "failed to find babies by creator", err)
	}

	return babies, nil
}

// FindAll 查找所有宝宝
func (r *babyRepositoryImpl) FindAll(ctx context.Context) ([]*entity.Baby, error) {
	var babies []*entity.Baby

	err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Order("create_time DESC").
		Find(&babies).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errors.Wrap(errors.DatabaseError, "failed to find all babies", err)
	}

	return babies, nil
}
