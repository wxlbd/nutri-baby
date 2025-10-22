package persistence

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/wxlbd/nutri-baby-server/internal/domain/entity"
	"github.com/wxlbd/nutri-baby-server/internal/domain/repository"
	"github.com/wxlbd/nutri-baby-server/pkg/errors"
)

// familyRepositoryImpl 家庭仓储实现
type familyRepositoryImpl struct {
	db *gorm.DB
}

// NewFamilyRepository 创建家庭仓储
func NewFamilyRepository(db *gorm.DB) repository.FamilyRepository {
	return &familyRepositoryImpl{db: db}
}

func (r *familyRepositoryImpl) Create(ctx context.Context, family *entity.Family) error {
	if err := r.db.WithContext(ctx).Create(family).Error; err != nil {
		return errors.Wrap(errors.DatabaseError, "failed to create family", err)
	}
	return nil
}

func (r *familyRepositoryImpl) FindByID(ctx context.Context, familyID string) (*entity.Family, error) {
	var family entity.Family
	err := r.db.WithContext(ctx).
		Where("family_id = ? AND deleted_at IS NULL", familyID).
		First(&family).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.ErrFamilyNotFound
	}
	if err != nil {
		return nil, errors.Wrap(errors.DatabaseError, "failed to find family", err)
	}

	return &family, nil
}

func (r *familyRepositoryImpl) Update(ctx context.Context, family *entity.Family) error {
	err := r.db.WithContext(ctx).
		Model(&entity.Family{}).
		Where("family_id = ? AND deleted_at IS NULL", family.FamilyID).
		Updates(family).Error

	if err != nil {
		return errors.Wrap(errors.DatabaseError, "failed to update family", err)
	}

	return nil
}

func (r *familyRepositoryImpl) Delete(ctx context.Context, familyID string) error {
	err := r.db.WithContext(ctx).
		Where("family_id = ?", familyID).
		Delete(&entity.Family{}).Error

	if err != nil {
		return errors.Wrap(errors.DatabaseError, "failed to delete family", err)
	}

	return nil
}

func (r *familyRepositoryImpl) FindByMember(ctx context.Context, openID string) ([]*entity.Family, error) {
	var families []*entity.Family

	err := r.db.WithContext(ctx).
		Joins("JOIN family_members ON family_members.family_id = families.family_id").
		Where("family_members.openid = ? AND family_members.deleted_at IS NULL AND families.deleted_at IS NULL", openID).
		Find(&families).Error

	if err != nil {
		return nil, errors.Wrap(errors.DatabaseError, "failed to find families by member", err)
	}

	return families, nil
}

// familyMemberRepositoryImpl 家庭成员仓储实现
type familyMemberRepositoryImpl struct {
	db *gorm.DB
}

// NewFamilyMemberRepository 创建家庭成员仓储
func NewFamilyMemberRepository(db *gorm.DB) repository.FamilyMemberRepository {
	return &familyMemberRepositoryImpl{db: db}
}

func (r *familyMemberRepositoryImpl) Add(ctx context.Context, member *entity.FamilyMember) error {
	if err := r.db.WithContext(ctx).Create(member).Error; err != nil {
		return errors.Wrap(errors.DatabaseError, "failed to add family member", err)
	}
	return nil
}

func (r *familyMemberRepositoryImpl) FindByFamilyID(ctx context.Context, familyID string) ([]*entity.FamilyMember, error) {
	var members []*entity.FamilyMember
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("family_id = ? AND deleted_at IS NULL", familyID).
		Order("join_time ASC").
		Find(&members).Error

	if err != nil {
		return nil, errors.Wrap(errors.DatabaseError, "failed to find family members", err)
	}

	return members, nil
}

func (r *familyMemberRepositoryImpl) FindByFamilyAndUser(ctx context.Context, familyID, openID string) (*entity.FamilyMember, error) {
	var member entity.FamilyMember
	err := r.db.WithContext(ctx).
		Where("family_id = ? AND openid = ? AND deleted_at IS NULL", familyID, openID).
		First(&member).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(errors.DatabaseError, "failed to find family member", err)
	}

	return &member, nil
}

func (r *familyMemberRepositoryImpl) Remove(ctx context.Context, familyID, openID string) error {
	err := r.db.WithContext(ctx).
		Where("family_id = ? AND openid = ?", familyID, openID).
		Delete(&entity.FamilyMember{}).Error

	if err != nil {
		return errors.Wrap(errors.DatabaseError, "failed to remove family member", err)
	}

	return nil
}

func (r *familyMemberRepositoryImpl) IsMember(ctx context.Context, familyID, openID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.FamilyMember{}).
		Where("family_id = ? AND openid = ? AND deleted_at IS NULL", familyID, openID).
		Count(&count).Error

	if err != nil {
		return false, errors.Wrap(errors.DatabaseError, "failed to check member", err)
	}

	return count > 0, nil
}

func (r *familyMemberRepositoryImpl) IsAdmin(ctx context.Context, familyID, openID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.FamilyMember{}).
		Where("family_id = ? AND openid = ? AND role = ? AND deleted_at IS NULL", familyID, openID, "admin").
		Count(&count).Error

	if err != nil {
		return false, errors.Wrap(errors.DatabaseError, "failed to check admin", err)
	}

	return count > 0, nil
}

// invitationRepositoryImpl 邀请码仓储实现
type invitationRepositoryImpl struct {
	db *gorm.DB
}

// NewInvitationRepository 创建邀请码仓储
func NewInvitationRepository(db *gorm.DB) repository.InvitationRepository {
	return &invitationRepositoryImpl{db: db}
}

func (r *invitationRepositoryImpl) Create(ctx context.Context, invitation *entity.Invitation) error {
	if err := r.db.WithContext(ctx).Create(invitation).Error; err != nil {
		return errors.Wrap(errors.DatabaseError, "failed to create invitation", err)
	}
	return nil
}

func (r *invitationRepositoryImpl) FindByCode(ctx context.Context, code string) (*entity.Invitation, error) {
	var invitation entity.Invitation
	err := r.db.WithContext(ctx).
		Preload("Family").
		Where("invitation_code = ? AND deleted_at IS NULL", code).
		First(&invitation).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.ErrInvalidInvitation
	}
	if err != nil {
		return nil, errors.Wrap(errors.DatabaseError, "failed to find invitation", err)
	}

	return &invitation, nil
}

func (r *invitationRepositoryImpl) Delete(ctx context.Context, code string) error {
	err := r.db.WithContext(ctx).
		Where("invitation_code = ?", code).
		Delete(&entity.Invitation{}).Error

	if err != nil {
		return errors.Wrap(errors.DatabaseError, "failed to delete invitation", err)
	}

	return nil
}

func (r *invitationRepositoryImpl) DeleteExpired(ctx context.Context) error {
	now := time.Now().UnixMilli()
	err := r.db.WithContext(ctx).
		Where("expires_at < ?", now).
		Delete(&entity.Invitation{}).Error

	if err != nil {
		return errors.Wrap(errors.DatabaseError, "failed to delete expired invitations", err)
	}

	return nil
}
