package query

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"kama_chat_server/internal/service/gorm"
	rpcserver "kama_chat_server/pkg/rpc/server"
)

type rpcMessageQueryService struct{}
type rpcSessionQueryService struct{}
type rpcUserQueryService struct{}
type rpcGroupQueryService struct{}
type rpcContactQueryService struct{}

type getMessageListRequest struct {
	UserOneID string `json:"user_one_id"`
	UserTwoID string `json:"user_two_id"`
}

type getGroupMessageListRequest struct {
	GroupID string `json:"group_id"`
}

type getSessionListRequest struct {
	OwnerID string `json:"owner_id"`
}

type getUserInfoListRequest struct {
	OwnerID string `json:"owner_id"`
}

type getUserInfoRequest struct {
	UUID string `json:"uuid"`
}

type loadMyGroupRequest struct {
	OwnerID string `json:"owner_id"`
}

type checkGroupAddModeRequest struct {
	GroupID string `json:"group_id"`
}

type getGroupInfoRequest struct {
	GroupID string `json:"group_id"`
}

type getGroupMemberListRequest struct {
	GroupID string `json:"group_id"`
}

type getUserListRequest struct {
	OwnerID string `json:"owner_id"`
}

type getContactInfoRequest struct {
	ContactID string `json:"contact_id"`
}

type getNewContactListRequest struct {
	OwnerID string `json:"owner_id"`
}

type getAddGroupListRequest struct {
	GroupID string `json:"group_id"`
}

type emptyRequest struct{}

func (service *rpcMessageQueryService) GetMessageList(ctx context.Context, request *getMessageListRequest) (*messageListResult, error) {
	message, list, ret := gorm.MessageService.GetMessageList(request.UserOneID, request.UserTwoID)
	return &messageListResult{Message: message, List: list, Ret: ret}, nil
}

func (service *rpcMessageQueryService) GetGroupMessageList(ctx context.Context, request *getGroupMessageListRequest) (*groupMessageListResult, error) {
	message, list, ret := gorm.MessageService.GetGroupMessageList(request.GroupID)
	return &groupMessageListResult{Message: message, List: list, Ret: ret}, nil
}

func (service *rpcSessionQueryService) GetUserSessionList(ctx context.Context, request *getSessionListRequest) (*userSessionListResult, error) {
	message, list, ret := gorm.SessionService.GetUserSessionList(request.OwnerID)
	return &userSessionListResult{Message: message, List: list, Ret: ret}, nil
}

func (service *rpcSessionQueryService) GetGroupSessionList(ctx context.Context, request *getSessionListRequest) (*groupSessionListResult, error) {
	message, list, ret := gorm.SessionService.GetGroupSessionList(request.OwnerID)
	return &groupSessionListResult{Message: message, List: list, Ret: ret}, nil
}

func (service *rpcUserQueryService) GetUserInfoList(ctx context.Context, request *getUserInfoListRequest) (*userInfoListResult, error) {
	message, list, ret := gorm.UserInfoService.GetUserInfoList(request.OwnerID)
	return &userInfoListResult{Message: message, List: list, Ret: ret}, nil
}

func (service *rpcUserQueryService) GetUserInfo(ctx context.Context, request *getUserInfoRequest) (*userInfoResult, error) {
	message, info, ret := gorm.UserInfoService.GetUserInfo(request.UUID)
	return &userInfoResult{Message: message, Info: info, Ret: ret}, nil
}

func (service *rpcGroupQueryService) LoadMyGroup(ctx context.Context, request *loadMyGroupRequest) (*myGroupListResult, error) {
	message, list, ret := gorm.GroupInfoService.LoadMyGroup(request.OwnerID)
	return &myGroupListResult{Message: message, List: list, Ret: ret}, nil
}

func (service *rpcGroupQueryService) CheckGroupAddMode(ctx context.Context, request *checkGroupAddModeRequest) (*addModeResult, error) {
	message, addMode, ret := gorm.GroupInfoService.CheckGroupAddMode(request.GroupID)
	return &addModeResult{Message: message, AddMode: addMode, Ret: ret}, nil
}

func (service *rpcGroupQueryService) GetGroupInfo(ctx context.Context, request *getGroupInfoRequest) (*groupInfoResult, error) {
	message, info, ret := gorm.GroupInfoService.GetGroupInfo(request.GroupID)
	return &groupInfoResult{Message: message, Info: info, Ret: ret}, nil
}

func (service *rpcGroupQueryService) GetGroupInfoList(ctx context.Context, request *emptyRequest) (*groupInfoListResult, error) {
	message, list, ret := gorm.GroupInfoService.GetGroupInfoList()
	return &groupInfoListResult{Message: message, List: list, Ret: ret}, nil
}

func (service *rpcGroupQueryService) GetGroupMemberList(ctx context.Context, request *getGroupMemberListRequest) (*groupMemberListResult, error) {
	message, list, ret := gorm.GroupInfoService.GetGroupMemberList(request.GroupID)
	return &groupMemberListResult{Message: message, List: list, Ret: ret}, nil
}

func (service *rpcContactQueryService) GetUserList(ctx context.Context, request *getUserListRequest) (*userListResult, error) {
	message, list, ret := gorm.UserContactService.GetUserList(request.OwnerID)
	return &userListResult{Message: message, List: list, Ret: ret}, nil
}

func (service *rpcContactQueryService) LoadMyJoinedGroup(ctx context.Context, request *getUserListRequest) (*joinedGroupListResult, error) {
	message, list, ret := gorm.UserContactService.LoadMyJoinedGroup(request.OwnerID)
	return &joinedGroupListResult{Message: message, List: list, Ret: ret}, nil
}

func (service *rpcContactQueryService) GetContactInfo(ctx context.Context, request *getContactInfoRequest) (*contactInfoResult, error) {
	message, info, ret := gorm.UserContactService.GetContactInfo(request.ContactID)
	return &contactInfoResult{Message: message, Info: info, Ret: ret}, nil
}

func (service *rpcContactQueryService) GetNewContactList(ctx context.Context, request *getNewContactListRequest) (*newContactListResult, error) {
	message, list, ret := gorm.UserContactService.GetNewContactList(request.OwnerID)
	return &newContactListResult{Message: message, List: list, Ret: ret}, nil
}

func (service *rpcContactQueryService) GetAddGroupList(ctx context.Context, request *getAddGroupListRequest) (*addGroupListResult, error) {
	message, list, ret := gorm.UserContactService.GetAddGroupList(request.GroupID)
	return &addGroupListResult{Message: message, List: list, Ret: ret}, nil
}

var startRPCServerOnce sync.Once

func ensureRPCServer(listenAddress string) error {
	if listenAddress == "" {
		return fmt.Errorf("query: rpc listen address is empty")
	}

	var initErr error
	startRPCServerOnce.Do(func() {
		srv := rpcserver.NewServer()
		if err := srv.RegisterService(&rpcMessageQueryService{}); err != nil {
			initErr = err
			return
		}
		if err := srv.RegisterService(&rpcSessionQueryService{}); err != nil {
			initErr = err
			return
		}
		if err := srv.RegisterService(&rpcUserQueryService{}); err != nil {
			initErr = err
			return
		}
		if err := srv.RegisterService(&rpcGroupQueryService{}); err != nil {
			initErr = err
			return
		}
		if err := srv.RegisterService(&rpcContactQueryService{}); err != nil {
			initErr = err
			return
		}

		listener, err := net.Listen("tcp", listenAddress)
		if err != nil {
			initErr = err
			return
		}

		go func() {
			if serveErr := srv.Serve(listener); serveErr != nil {
				log.Printf("query rpc server stopped: %v", serveErr)
			}
		}()
	})
	return initErr
}
