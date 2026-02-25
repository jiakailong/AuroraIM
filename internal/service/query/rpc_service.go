package query

import (
	"context"
	"fmt"
	"strings"
	"time"

	"kama_chat_server/internal/dto/respond"
	rpcclient "kama_chat_server/pkg/rpc/client"
)

type rpcService struct {
	client *rpcclient.Client
}

func newRPCService(targetAddress string) (Service, error) {
	client, err := rpcclient.NewClient(
		targetAddress,
		rpcclient.WithTimeout(3*time.Second),
		rpcclient.WithRetryMax(1),
		rpcclient.WithRetryBackoff(20*time.Millisecond, 120*time.Millisecond, 0.1),
		rpcclient.WithIdempotentMethods([]string{
			"rpcMessageQueryService.GetMessageList",
			"rpcMessageQueryService.GetGroupMessageList",
			"rpcSessionQueryService.GetUserSessionList",
			"rpcSessionQueryService.GetGroupSessionList",
			"rpcUserQueryService.GetUserInfoList",
			"rpcUserQueryService.GetUserInfo",
			"rpcGroupQueryService.LoadMyGroup",
			"rpcGroupQueryService.CheckGroupAddMode",
			"rpcGroupQueryService.GetGroupInfo",
			"rpcGroupQueryService.GetGroupInfoList",
			"rpcGroupQueryService.GetGroupMemberList",
			"rpcContactQueryService.GetUserList",
			"rpcContactQueryService.LoadMyJoinedGroup",
			"rpcContactQueryService.GetContactInfo",
			"rpcContactQueryService.GetNewContactList",
			"rpcContactQueryService.GetAddGroupList",
		}),
	)
	if err != nil {
		return nil, err
	}
	return &rpcService{client: client}, nil
}

func (service *rpcService) GetMessageList(userOneId, userTwoId string) (string, []respond.GetMessageListRespond, int) {
	request := &getMessageListRequest{UserOneID: userOneId, UserTwoID: userTwoId}
	result := &messageListResult{}
	if err := service.client.Call(context.Background(), "rpcMessageQueryService.GetMessageList", request, result); err != nil {
		return fmt.Sprintf("rpc call failed: %v", err), nil, -1
	}
	if result.List == nil {
		result.List = []respond.GetMessageListRespond{}
	}
	return result.Message, result.List, result.Ret
}

func (service *rpcService) GetGroupMessageList(groupId string) (string, []respond.GetGroupMessageListRespond, int) {
	request := &getGroupMessageListRequest{GroupID: groupId}
	result := &groupMessageListResult{}
	if err := service.client.Call(context.Background(), "rpcMessageQueryService.GetGroupMessageList", request, result); err != nil {
		return fmt.Sprintf("rpc call failed: %v", err), nil, -1
	}
	if result.List == nil {
		result.List = []respond.GetGroupMessageListRespond{}
	}
	return result.Message, result.List, result.Ret
}

func (service *rpcService) GetUserSessionList(ownerId string) (string, []respond.UserSessionListRespond, int) {
	request := &getSessionListRequest{OwnerID: ownerId}
	result := &userSessionListResult{}
	if err := service.client.Call(context.Background(), "rpcSessionQueryService.GetUserSessionList", request, result); err != nil {
		return fmt.Sprintf("rpc call failed: %v", err), nil, -1
	}
	if result.List == nil {
		result.List = []respond.UserSessionListRespond{}
	}
	return result.Message, result.List, result.Ret
}

func (service *rpcService) GetGroupSessionList(ownerId string) (string, []respond.GroupSessionListRespond, int) {
	request := &getSessionListRequest{OwnerID: ownerId}
	result := &groupSessionListResult{}
	if err := service.client.Call(context.Background(), "rpcSessionQueryService.GetGroupSessionList", request, result); err != nil {
		return fmt.Sprintf("rpc call failed: %v", err), nil, -1
	}
	if result.List == nil {
		result.List = []respond.GroupSessionListRespond{}
	}
	return result.Message, result.List, result.Ret
}

func (service *rpcService) GetUserInfoList(ownerId string) (string, []respond.GetUserListRespond, int) {
	request := &getUserInfoListRequest{OwnerID: ownerId}
	result := &userInfoListResult{}
	if err := service.client.Call(context.Background(), "rpcUserQueryService.GetUserInfoList", request, result); err != nil {
		return fmt.Sprintf("rpc call failed: %v", err), nil, -1
	}
	if result.List == nil {
		result.List = []respond.GetUserListRespond{}
	}
	return result.Message, result.List, result.Ret
}

func (service *rpcService) GetUserInfo(uuid string) (string, *respond.GetUserInfoRespond, int) {
	request := &getUserInfoRequest{UUID: uuid}
	result := &userInfoResult{}
	if err := service.client.Call(context.Background(), "rpcUserQueryService.GetUserInfo", request, result); err != nil {
		return fmt.Sprintf("rpc call failed: %v", err), nil, -1
	}
	return result.Message, result.Info, result.Ret
}

func (service *rpcService) LoadMyGroup(ownerId string) (string, []respond.LoadMyGroupRespond, int) {
	request := &loadMyGroupRequest{OwnerID: ownerId}
	result := &myGroupListResult{}
	if err := service.client.Call(context.Background(), "rpcGroupQueryService.LoadMyGroup", request, result); err != nil {
		return fmt.Sprintf("rpc call failed: %v", err), nil, -1
	}
	if result.List == nil {
		result.List = []respond.LoadMyGroupRespond{}
	}
	return result.Message, result.List, result.Ret
}

func (service *rpcService) CheckGroupAddMode(groupId string) (string, int8, int) {
	request := &checkGroupAddModeRequest{GroupID: groupId}
	result := &addModeResult{}
	if err := service.client.Call(context.Background(), "rpcGroupQueryService.CheckGroupAddMode", request, result); err != nil {
		return fmt.Sprintf("rpc call failed: %v", err), 0, -1
	}
	return result.Message, result.AddMode, result.Ret
}

func (service *rpcService) GetGroupInfo(groupId string) (string, *respond.GetGroupInfoRespond, int) {
	request := &getGroupInfoRequest{GroupID: groupId}
	result := &groupInfoResult{}
	if err := service.client.Call(context.Background(), "rpcGroupQueryService.GetGroupInfo", request, result); err != nil {
		return fmt.Sprintf("rpc call failed: %v", err), nil, -1
	}
	return result.Message, result.Info, result.Ret
}

func (service *rpcService) GetGroupInfoList() (string, []respond.GetGroupListRespond, int) {
	request := &emptyRequest{}
	result := &groupInfoListResult{}
	if err := service.client.Call(context.Background(), "rpcGroupQueryService.GetGroupInfoList", request, result); err != nil {
		return fmt.Sprintf("rpc call failed: %v", err), nil, -1
	}
	if result.List == nil {
		result.List = []respond.GetGroupListRespond{}
	}
	return result.Message, result.List, result.Ret
}

func (service *rpcService) GetGroupMemberList(groupId string) (string, []respond.GetGroupMemberListRespond, int) {
	request := &getGroupMemberListRequest{GroupID: groupId}
	result := &groupMemberListResult{}
	if err := service.client.Call(context.Background(), "rpcGroupQueryService.GetGroupMemberList", request, result); err != nil {
		return fmt.Sprintf("rpc call failed: %v", err), nil, -1
	}
	if result.List == nil {
		result.List = []respond.GetGroupMemberListRespond{}
	}
	return result.Message, result.List, result.Ret
}

func (service *rpcService) GetUserList(ownerId string) (string, []respond.MyUserListRespond, int) {
	request := &getUserListRequest{OwnerID: ownerId}
	result := &userListResult{}
	if err := service.client.Call(context.Background(), "rpcContactQueryService.GetUserList", request, result); err != nil {
		return fmt.Sprintf("rpc call failed: %v", err), nil, -1
	}
	if result.List == nil {
		result.List = []respond.MyUserListRespond{}
	}
	return result.Message, result.List, result.Ret
}

func (service *rpcService) LoadMyJoinedGroup(ownerId string) (string, []respond.LoadMyJoinedGroupRespond, int) {
	request := &getUserListRequest{OwnerID: ownerId}
	result := &joinedGroupListResult{}
	if err := service.client.Call(context.Background(), "rpcContactQueryService.LoadMyJoinedGroup", request, result); err != nil {
		return fmt.Sprintf("rpc call failed: %v", err), nil, -1
	}
	if result.List == nil {
		result.List = []respond.LoadMyJoinedGroupRespond{}
	}
	return result.Message, result.List, result.Ret
}

func (service *rpcService) GetContactInfo(contactId string) (string, respond.GetContactInfoRespond, int) {
	request := &getContactInfoRequest{ContactID: contactId}
	result := &contactInfoResult{}
	if err := service.client.Call(context.Background(), "rpcContactQueryService.GetContactInfo", request, result); err != nil {
		return fmt.Sprintf("rpc call failed: %v", err), respond.GetContactInfoRespond{}, -1
	}
	return result.Message, result.Info, result.Ret
}

func (service *rpcService) GetNewContactList(ownerId string) (string, []respond.NewContactListRespond, int) {
	request := &getNewContactListRequest{OwnerID: ownerId}
	result := &newContactListResult{}
	if err := service.client.Call(context.Background(), "rpcContactQueryService.GetNewContactList", request, result); err != nil {
		return fmt.Sprintf("rpc call failed: %v", err), nil, -1
	}
	if result.List == nil {
		result.List = []respond.NewContactListRespond{}
	}
	return result.Message, result.List, result.Ret
}

func (service *rpcService) GetAddGroupList(groupId string) (string, []respond.AddGroupListRespond, int) {
	request := &getAddGroupListRequest{GroupID: groupId}
	result := &addGroupListResult{}
	if err := service.client.Call(context.Background(), "rpcContactQueryService.GetAddGroupList", request, result); err != nil {
		return fmt.Sprintf("rpc call failed: %v", err), nil, -1
	}
	if result.List == nil {
		result.List = []respond.AddGroupListRespond{}
	}
	return result.Message, result.List, result.Ret
}

func Init(config Config) error {
	mode := strings.TrimSpace(strings.ToLower(config.Mode))
	if mode == "" || mode == "local" {
		QueryService = newLocalService()
		return nil
	}

	switch mode {
	case "rpc":
		listenAddress := strings.TrimSpace(config.RPCListen)
		targetAddress := strings.TrimSpace(config.RPCTarget)

		if listenAddress != "" {
			if err := ensureRPCServer(listenAddress); err != nil {
				return err
			}
			if targetAddress == "" {
				targetAddress = listenAddress
			}
		}

		if targetAddress == "" {
			targetAddress = "127.0.0.1:19110"
			if err := ensureRPCServer(targetAddress); err != nil {
				return err
			}
		}

		service, err := newRPCService(targetAddress)
		if err != nil {
			return err
		}
		QueryService = service
		return nil
	case "grpc":
		listenAddress := strings.TrimSpace(config.GRPCListen)
		targetAddress := strings.TrimSpace(config.GRPCTarget)

		if listenAddress != "" {
			if err := ensureGRPCServer(listenAddress); err != nil {
				return err
			}
			if targetAddress == "" {
				targetAddress = listenAddress
			}
		}

		if targetAddress == "" {
			targetAddress = "127.0.0.1:19120"
			if err := ensureGRPCServer(targetAddress); err != nil {
				return err
			}
		}

		service, err := newGRPCService(targetAddress)
		if err != nil {
			return err
		}
		QueryService = service
		return nil
	default:
		return fmt.Errorf("query: unsupported mode %s", config.Mode)
	}
}
