package service

import (
	"errors"
	"template-mall/TemplateOrderServer/internal/repository"
)

var ErrMemberUpdateFailed = errors.New("member status update failed")

// MemberService 会员服务
type MemberService struct {
	userRepo *repository.UserRepo
}

// NewMemberService 创建会员服务实例。
func NewMemberService(userRepo *repository.UserRepo) *MemberService {
	return &MemberService{userRepo: userRepo}
}

// SetMember 设置会员状态（0=非会员 1=会员）。
func (s *MemberService) SetMember(userID uint64, memberStatus int8) error {
	if memberStatus != 0 && memberStatus != 1 {
		return errors.New("member_status must be 0 or 1")
	}
	return s.userRepo.UpdateMemberStatus(userID, memberStatus)
}