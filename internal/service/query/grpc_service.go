package query

import (
	"context"
	"fmt"
	"time"

	"kama_chat_server/internal/dto/respond"
	pb "kama_chat_server/proto/query"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

type grpcService struct {
	msgClient     pb.MessageQueryServiceClient
	sessClient    pb.SessionQueryServiceClient
	userClient    pb.UserQueryServiceClient
	groupClient   pb.GroupQueryServiceClient
	contactClient pb.ContactQueryServiceClient
}

func newGRPCService(targetAddress string) (Service, error) {
	dialCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		dialCtx,
		targetAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	return &grpcService{
		msgClient:     pb.NewMessageQueryServiceClient(conn),
		sessClient:    pb.NewSessionQueryServiceClient(conn),
		userClient:    pb.NewUserQueryServiceClient(conn),
		groupClient:   pb.NewGroupQueryServiceClient(conn),
		contactClient: pb.NewContactQueryServiceClient(conn),
	}, nil
}

func callCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 3*time.Second)
}

// ==================== Message ====================

func (s *grpcService) GetMessageList(userOneId, userTwoId string) (string, []respond.GetMessageListRespond, int) {
	ctx, cancel := callCtx()
	defer cancel()
	resp, err := s.msgClient.GetMessageList(ctx, &pb.GetMessageListRequest{
		UserOneId: userOneId,
		UserTwoId: userTwoId,
	})
	if err != nil {
		return fmt.Sprintf("grpc call failed: %v", err), nil, -1
	}
	list := make([]respond.GetMessageListRespond, 0, len(resp.List))
	for _, item := range resp.List {
		list = append(list, respond.GetMessageListRespond{
			SendId:     item.SendId,
			SendName:   item.SendName,
			SendAvatar: item.SendAvatar,
			ReceiveId:  item.ReceiveId,
			Type:       int8(item.Type),
			Content:    item.Content,
			Url:        item.Url,
			FileType:   item.FileType,
			FileName:   item.FileName,
			FileSize:   item.FileSize,
			CreatedAt:  item.CreatedAt,
		})
	}
	return resp.Message, list, int(resp.Ret)
}

func (s *grpcService) GetGroupMessageList(groupId string) (string, []respond.GetGroupMessageListRespond, int) {
	ctx, cancel := callCtx()
	defer cancel()
	resp, err := s.msgClient.GetGroupMessageList(ctx, &pb.GetGroupMessageListRequest{GroupId: groupId})
	if err != nil {
		return fmt.Sprintf("grpc call failed: %v", err), nil, -1
	}
	list := make([]respond.GetGroupMessageListRespond, 0, len(resp.List))
	for _, item := range resp.List {
		list = append(list, respond.GetGroupMessageListRespond{
			SendId:     item.SendId,
			SendName:   item.SendName,
			SendAvatar: item.SendAvatar,
			ReceiveId:  item.ReceiveId,
			Type:       int8(item.Type),
			Content:    item.Content,
			Url:        item.Url,
			FileType:   item.FileType,
			FileName:   item.FileName,
			FileSize:   item.FileSize,
			CreatedAt:  item.CreatedAt,
		})
	}
	return resp.Message, list, int(resp.Ret)
}

// ==================== Session ====================

func (s *grpcService) GetUserSessionList(ownerId string) (string, []respond.UserSessionListRespond, int) {
	ctx, cancel := callCtx()
	defer cancel()
	resp, err := s.sessClient.GetUserSessionList(ctx, &pb.GetSessionListRequest{OwnerId: ownerId})
	if err != nil {
		return fmt.Sprintf("grpc call failed: %v", err), nil, -1
	}
	list := make([]respond.UserSessionListRespond, 0, len(resp.List))
	for _, item := range resp.List {
		list = append(list, respond.UserSessionListRespond{
			SessionId: item.SessionId,
			Avatar:    item.Avatar,
			UserId:    item.UserId,
			Username:  item.UserName,
		})
	}
	return resp.Message, list, int(resp.Ret)
}

func (s *grpcService) GetGroupSessionList(ownerId string) (string, []respond.GroupSessionListRespond, int) {
	ctx, cancel := callCtx()
	defer cancel()
	resp, err := s.sessClient.GetGroupSessionList(ctx, &pb.GetSessionListRequest{OwnerId: ownerId})
	if err != nil {
		return fmt.Sprintf("grpc call failed: %v", err), nil, -1
	}
	list := make([]respond.GroupSessionListRespond, 0, len(resp.List))
	for _, item := range resp.List {
		list = append(list, respond.GroupSessionListRespond{
			SessionId: item.SessionId,
			GroupName: item.GroupName,
			GroupId:   item.GroupId,
			Avatar:    item.Avatar,
		})
	}
	return resp.Message, list, int(resp.Ret)
}

// ==================== User ====================

func (s *grpcService) GetUserInfoList(ownerId string) (string, []respond.GetUserListRespond, int) {
	ctx, cancel := callCtx()
	defer cancel()
	resp, err := s.userClient.GetUserInfoList(ctx, &pb.GetUserInfoListRequest{OwnerId: ownerId})
	if err != nil {
		return fmt.Sprintf("grpc call failed: %v", err), nil, -1
	}
	list := make([]respond.GetUserListRespond, 0, len(resp.List))
	for _, item := range resp.List {
		list = append(list, respond.GetUserListRespond{
			Uuid:      item.Uuid,
			Nickname:  item.Nickname,
			Telephone: item.Telephone,
			Status:    int8(item.Status),
			IsAdmin:   int8(item.IsAdmin),
			IsDeleted: item.IsDeleted,
		})
	}
	return resp.Message, list, int(resp.Ret)
}

func (s *grpcService) GetUserInfo(uuid string) (string, *respond.GetUserInfoRespond, int) {
	ctx, cancel := callCtx()
	defer cancel()
	resp, err := s.userClient.GetUserInfo(ctx, &pb.GetUserInfoRequest{Uuid: uuid})
	if err != nil {
		return fmt.Sprintf("grpc call failed: %v", err), nil, -1
	}
	var info *respond.GetUserInfoRespond
	if resp.Info != nil {
		info = &respond.GetUserInfoRespond{
			Uuid:      resp.Info.Uuid,
			Nickname:  resp.Info.Nickname,
			Telephone: resp.Info.Telephone,
			Avatar:    resp.Info.Avatar,
			Email:     resp.Info.Email,
			Gender:    int8(resp.Info.Gender),
			Birthday:  resp.Info.Birthday,
			Signature: resp.Info.Signature,
			CreatedAt: resp.Info.CreatedAt,
			IsAdmin:   int8(resp.Info.IsAdmin),
			Status:    int8(resp.Info.Status),
		}
	}
	return resp.Message, info, int(resp.Ret)
}

// ==================== Group ====================

func (s *grpcService) LoadMyGroup(ownerId string) (string, []respond.LoadMyGroupRespond, int) {
	ctx, cancel := callCtx()
	defer cancel()
	resp, err := s.groupClient.LoadMyGroup(ctx, &pb.LoadMyGroupRequest{OwnerId: ownerId})
	if err != nil {
		return fmt.Sprintf("grpc call failed: %v", err), nil, -1
	}
	list := make([]respond.LoadMyGroupRespond, 0, len(resp.List))
	for _, item := range resp.List {
		list = append(list, respond.LoadMyGroupRespond{
			GroupId:   item.GroupId,
			GroupName: item.GroupName,
			Avatar:    item.Avatar,
		})
	}
	return resp.Message, list, int(resp.Ret)
}

func (s *grpcService) CheckGroupAddMode(groupId string) (string, int8, int) {
	ctx, cancel := callCtx()
	defer cancel()
	resp, err := s.groupClient.CheckGroupAddMode(ctx, &pb.CheckGroupAddModeRequest{GroupId: groupId})
	if err != nil {
		return fmt.Sprintf("grpc call failed: %v", err), 0, -1
	}
	return resp.Message, int8(resp.AddMode), int(resp.Ret)
}

func (s *grpcService) GetGroupInfo(groupId string) (string, *respond.GetGroupInfoRespond, int) {
	ctx, cancel := callCtx()
	defer cancel()
	resp, err := s.groupClient.GetGroupInfo(ctx, &pb.GetGroupInfoRequest{GroupId: groupId})
	if err != nil {
		return fmt.Sprintf("grpc call failed: %v", err), nil, -1
	}
	var info *respond.GetGroupInfoRespond
	if resp.Info != nil {
		info = &respond.GetGroupInfoRespond{
			Uuid:      resp.Info.Uuid,
			Name:      resp.Info.Name,
			Notice:    resp.Info.Notice,
			MemberCnt: int(resp.Info.MemberCnt),
			OwnerId:   resp.Info.OwnerId,
			AddMode:   int8(resp.Info.AddMode),
			Status:    int8(resp.Info.Status),
			Avatar:    resp.Info.Avatar,
			IsDeleted: resp.Info.IsDeleted,
		}
	}
	return resp.Message, info, int(resp.Ret)
}

func (s *grpcService) GetGroupInfoList() (string, []respond.GetGroupListRespond, int) {
	ctx, cancel := callCtx()
	defer cancel()
	resp, err := s.groupClient.GetGroupInfoList(ctx, &emptypb.Empty{})
	if err != nil {
		return fmt.Sprintf("grpc call failed: %v", err), nil, -1
	}
	list := make([]respond.GetGroupListRespond, 0, len(resp.List))
	for _, item := range resp.List {
		list = append(list, respond.GetGroupListRespond{
			Uuid:      item.Uuid,
			Name:      item.Name,
			OwnerId:   item.OwnerId,
			Status:    int8(item.Status),
			IsDeleted: item.IsDeleted,
		})
	}
	return resp.Message, list, int(resp.Ret)
}

func (s *grpcService) GetGroupMemberList(groupId string) (string, []respond.GetGroupMemberListRespond, int) {
	ctx, cancel := callCtx()
	defer cancel()
	resp, err := s.groupClient.GetGroupMemberList(ctx, &pb.GetGroupMemberListRequest{GroupId: groupId})
	if err != nil {
		return fmt.Sprintf("grpc call failed: %v", err), nil, -1
	}
	list := make([]respond.GetGroupMemberListRespond, 0, len(resp.List))
	for _, item := range resp.List {
		list = append(list, respond.GetGroupMemberListRespond{
			UserId:   item.UserId,
			Nickname: item.Nickname,
			Avatar:   item.Avatar,
		})
	}
	return resp.Message, list, int(resp.Ret)
}

// ==================== Contact ====================

func (s *grpcService) GetUserList(ownerId string) (string, []respond.MyUserListRespond, int) {
	ctx, cancel := callCtx()
	defer cancel()
	resp, err := s.contactClient.GetUserList(ctx, &pb.GetContactUserListRequest{OwnerId: ownerId})
	if err != nil {
		return fmt.Sprintf("grpc call failed: %v", err), nil, -1
	}
	list := make([]respond.MyUserListRespond, 0, len(resp.List))
	for _, item := range resp.List {
		list = append(list, respond.MyUserListRespond{
			UserId:   item.UserId,
			UserName: item.UserName,
			Avatar:   item.Avatar,
		})
	}
	return resp.Message, list, int(resp.Ret)
}

func (s *grpcService) LoadMyJoinedGroup(ownerId string) (string, []respond.LoadMyJoinedGroupRespond, int) {
	ctx, cancel := callCtx()
	defer cancel()
	resp, err := s.contactClient.LoadMyJoinedGroup(ctx, &pb.LoadMyJoinedGroupRequest{OwnerId: ownerId})
	if err != nil {
		return fmt.Sprintf("grpc call failed: %v", err), nil, -1
	}
	list := make([]respond.LoadMyJoinedGroupRespond, 0, len(resp.List))
	for _, item := range resp.List {
		list = append(list, respond.LoadMyJoinedGroupRespond{
			GroupId:   item.GroupId,
			GroupName: item.GroupName,
			Avatar:    item.Avatar,
		})
	}
	return resp.Message, list, int(resp.Ret)
}

func (s *grpcService) GetContactInfo(contactId string) (string, respond.GetContactInfoRespond, int) {
	ctx, cancel := callCtx()
	defer cancel()
	resp, err := s.contactClient.GetContactInfo(ctx, &pb.GetContactInfoRequest{ContactId: contactId})
	if err != nil {
		return fmt.Sprintf("grpc call failed: %v", err), respond.GetContactInfoRespond{}, -1
	}
	info := respond.GetContactInfoRespond{}
	if resp.Info != nil {
		info = respond.GetContactInfoRespond{
			ContactId:        resp.Info.ContactId,
			ContactName:      resp.Info.ContactName,
			ContactAvatar:    resp.Info.ContactAvatar,
			ContactPhone:     resp.Info.ContactPhone,
			ContactEmail:     resp.Info.ContactEmail,
			ContactGender:    int8(resp.Info.ContactGender),
			ContactSignature: resp.Info.ContactSignature,
			ContactBirthday:  resp.Info.ContactBirthday,
			ContactNotice:    resp.Info.ContactNotice,
			ContactMembers:   resp.Info.ContactMembers,
			ContactMemberCnt: int(resp.Info.ContactMemberCnt),
			ContactOwnerId:   resp.Info.ContactOwnerId,
			ContactAddMode:   int8(resp.Info.ContactAddMode),
		}
	}
	return resp.Message, info, int(resp.Ret)
}

func (s *grpcService) GetNewContactList(ownerId string) (string, []respond.NewContactListRespond, int) {
	ctx, cancel := callCtx()
	defer cancel()
	resp, err := s.contactClient.GetNewContactList(ctx, &pb.GetNewContactListRequest{OwnerId: ownerId})
	if err != nil {
		return fmt.Sprintf("grpc call failed: %v", err), nil, -1
	}
	list := make([]respond.NewContactListRespond, 0, len(resp.List))
	for _, item := range resp.List {
		list = append(list, respond.NewContactListRespond{
			ContactId:     item.ContactId,
			ContactName:   item.ContactName,
			ContactAvatar: item.ContactAvatar,
			Message:       item.Msg,
		})
	}
	return resp.Message, list, int(resp.Ret)
}

func (s *grpcService) GetAddGroupList(groupId string) (string, []respond.AddGroupListRespond, int) {
	ctx, cancel := callCtx()
	defer cancel()
	resp, err := s.contactClient.GetAddGroupList(ctx, &pb.GetAddGroupListRequest{GroupId: groupId})
	if err != nil {
		return fmt.Sprintf("grpc call failed: %v", err), nil, -1
	}
	list := make([]respond.AddGroupListRespond, 0, len(resp.List))
	for _, item := range resp.List {
		list = append(list, respond.AddGroupListRespond{
			ContactId:     item.ContactId,
			ContactName:   item.ContactName,
			ContactAvatar: item.ContactAvatar,
			Message:       item.Msg,
		})
	}
	return resp.Message, list, int(resp.Ret)
}
