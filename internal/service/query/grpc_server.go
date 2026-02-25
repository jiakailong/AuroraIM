package query

import (
"context"
"fmt"
"log"
"net"
"sync"

"kama_chat_server/internal/dto/respond"
"kama_chat_server/internal/service/gorm"
pb "kama_chat_server/proto/query"

"google.golang.org/grpc"
"google.golang.org/protobuf/types/known/emptypb"
)

// ==================== MessageQueryService ====================

type messageQueryServer struct {
pb.UnimplementedMessageQueryServiceServer
}

func (s *messageQueryServer) GetMessageList(ctx context.Context, req *pb.GetMessageListRequest) (*pb.GetMessageListResponse, error) {
msg, list, ret := gorm.MessageService.GetMessageList(req.UserOneId, req.UserTwoId)
return &pb.GetMessageListResponse{
Message: msg,
List:    convertMessageItems(list),
Ret:     int32(ret),
}, nil
}

func (s *messageQueryServer) GetGroupMessageList(ctx context.Context, req *pb.GetGroupMessageListRequest) (*pb.GetGroupMessageListResponse, error) {
msg, list, ret := gorm.MessageService.GetGroupMessageList(req.GroupId)
pbList := make([]*pb.MessageItem, 0, len(list))
for _, item := range list {
pbList = append(pbList, &pb.MessageItem{
SendId:     item.SendId,
SendName:   item.SendName,
SendAvatar: item.SendAvatar,
ReceiveId:  item.ReceiveId,
Type:       int32(item.Type),
Content:    item.Content,
Url:        item.Url,
FileType:   item.FileType,
FileName:   item.FileName,
FileSize:   item.FileSize,
CreatedAt:  item.CreatedAt,
})
}
return &pb.GetGroupMessageListResponse{
Message: msg,
List:    pbList,
Ret:     int32(ret),
}, nil
}

func convertMessageItems(list []respond.GetMessageListRespond) []*pb.MessageItem {
pbList := make([]*pb.MessageItem, 0, len(list))
for _, item := range list {
pbList = append(pbList, &pb.MessageItem{
SendId:     item.SendId,
SendName:   item.SendName,
SendAvatar: item.SendAvatar,
ReceiveId:  item.ReceiveId,
Type:       int32(item.Type),
Content:    item.Content,
Url:        item.Url,
FileType:   item.FileType,
FileName:   item.FileName,
FileSize:   item.FileSize,
CreatedAt:  item.CreatedAt,
})
}
return pbList
}

// ==================== SessionQueryService ====================

type sessionQueryServer struct {
pb.UnimplementedSessionQueryServiceServer
}

func (s *sessionQueryServer) GetUserSessionList(ctx context.Context, req *pb.GetSessionListRequest) (*pb.UserSessionListResponse, error) {
msg, list, ret := gorm.SessionService.GetUserSessionList(req.OwnerId)
pbList := make([]*pb.UserSessionItem, 0, len(list))
for _, item := range list {
pbList = append(pbList, &pb.UserSessionItem{
SessionId: item.SessionId,
Avatar:    item.Avatar,
UserId:    item.UserId,
UserName:  item.Username,
})
}
return &pb.UserSessionListResponse{
Message: msg,
List:    pbList,
Ret:     int32(ret),
}, nil
}

func (s *sessionQueryServer) GetGroupSessionList(ctx context.Context, req *pb.GetSessionListRequest) (*pb.GroupSessionListResponse, error) {
msg, list, ret := gorm.SessionService.GetGroupSessionList(req.OwnerId)
pbList := make([]*pb.GroupSessionItem, 0, len(list))
for _, item := range list {
pbList = append(pbList, &pb.GroupSessionItem{
SessionId: item.SessionId,
GroupName: item.GroupName,
GroupId:   item.GroupId,
Avatar:    item.Avatar,
})
}
return &pb.GroupSessionListResponse{
Message: msg,
List:    pbList,
Ret:     int32(ret),
}, nil
}

// ==================== UserQueryService ====================

type userQueryServer struct {
pb.UnimplementedUserQueryServiceServer
}

func (s *userQueryServer) GetUserInfoList(ctx context.Context, req *pb.GetUserInfoListRequest) (*pb.GetUserInfoListResponse, error) {
msg, list, ret := gorm.UserInfoService.GetUserInfoList(req.OwnerId)
pbList := make([]*pb.UserListItem, 0, len(list))
for _, item := range list {
pbList = append(pbList, &pb.UserListItem{
Uuid:      item.Uuid,
Nickname:  item.Nickname,
Telephone: item.Telephone,
Status:    int32(item.Status),
IsAdmin:   int32(item.IsAdmin),
IsDeleted: item.IsDeleted,
})
}
return &pb.GetUserInfoListResponse{
Message: msg,
List:    pbList,
Ret:     int32(ret),
}, nil
}

func (s *userQueryServer) GetUserInfo(ctx context.Context, req *pb.GetUserInfoRequest) (*pb.GetUserInfoResponse, error) {
msg, info, ret := gorm.UserInfoService.GetUserInfo(req.Uuid)
var pbInfo *pb.UserInfoDetail
if info != nil {
pbInfo = &pb.UserInfoDetail{
Uuid:      info.Uuid,
Nickname:  info.Nickname,
Telephone: info.Telephone,
Avatar:    info.Avatar,
Email:     info.Email,
Gender:    int32(info.Gender),
Birthday:  info.Birthday,
Signature: info.Signature,
CreatedAt: info.CreatedAt,
IsAdmin:   int32(info.IsAdmin),
Status:    int32(info.Status),
}
}
return &pb.GetUserInfoResponse{
Message: msg,
Info:    pbInfo,
Ret:     int32(ret),
}, nil
}

// ==================== GroupQueryService ====================

type groupQueryServer struct {
pb.UnimplementedGroupQueryServiceServer
}

func (s *groupQueryServer) LoadMyGroup(ctx context.Context, req *pb.LoadMyGroupRequest) (*pb.LoadMyGroupResponse, error) {
msg, list, ret := gorm.GroupInfoService.LoadMyGroup(req.OwnerId)
pbList := make([]*pb.MyGroupItem, 0, len(list))
for _, item := range list {
pbList = append(pbList, &pb.MyGroupItem{
GroupId:   item.GroupId,
GroupName: item.GroupName,
Avatar:    item.Avatar,
})
}
return &pb.LoadMyGroupResponse{
Message: msg,
List:    pbList,
Ret:     int32(ret),
}, nil
}

func (s *groupQueryServer) CheckGroupAddMode(ctx context.Context, req *pb.CheckGroupAddModeRequest) (*pb.CheckGroupAddModeResponse, error) {
msg, addMode, ret := gorm.GroupInfoService.CheckGroupAddMode(req.GroupId)
return &pb.CheckGroupAddModeResponse{
Message: msg,
AddMode: int32(addMode),
Ret:     int32(ret),
}, nil
}

func (s *groupQueryServer) GetGroupInfo(ctx context.Context, req *pb.GetGroupInfoRequest) (*pb.GetGroupInfoResponse, error) {
msg, info, ret := gorm.GroupInfoService.GetGroupInfo(req.GroupId)
var pbInfo *pb.GroupInfoDetail
if info != nil {
pbInfo = &pb.GroupInfoDetail{
Uuid:      info.Uuid,
Name:      info.Name,
Notice:    info.Notice,
MemberCnt: int32(info.MemberCnt),
OwnerId:   info.OwnerId,
AddMode:   int32(info.AddMode),
Status:    int32(info.Status),
Avatar:    info.Avatar,
IsDeleted: info.IsDeleted,
}
}
return &pb.GetGroupInfoResponse{
Message: msg,
Info:    pbInfo,
Ret:     int32(ret),
}, nil
}

func (s *groupQueryServer) GetGroupInfoList(ctx context.Context, _ *emptypb.Empty) (*pb.GetGroupInfoListResponse, error) {
msg, list, ret := gorm.GroupInfoService.GetGroupInfoList()
pbList := make([]*pb.GroupListItem, 0, len(list))
for _, item := range list {
pbList = append(pbList, &pb.GroupListItem{
Uuid:      item.Uuid,
Name:      item.Name,
OwnerId:   item.OwnerId,
Status:    int32(item.Status),
IsDeleted: item.IsDeleted,
})
}
return &pb.GetGroupInfoListResponse{
Message: msg,
List:    pbList,
Ret:     int32(ret),
}, nil
}

func (s *groupQueryServer) GetGroupMemberList(ctx context.Context, req *pb.GetGroupMemberListRequest) (*pb.GetGroupMemberListResponse, error) {
msg, list, ret := gorm.GroupInfoService.GetGroupMemberList(req.GroupId)
pbList := make([]*pb.GroupMemberItem, 0, len(list))
for _, item := range list {
pbList = append(pbList, &pb.GroupMemberItem{
UserId:   item.UserId,
Nickname: item.Nickname,
Avatar:   item.Avatar,
})
}
return &pb.GetGroupMemberListResponse{
Message: msg,
List:    pbList,
Ret:     int32(ret),
}, nil
}

// ==================== ContactQueryService ====================

type contactQueryServer struct {
pb.UnimplementedContactQueryServiceServer
}

func (s *contactQueryServer) GetUserList(ctx context.Context, req *pb.GetContactUserListRequest) (*pb.GetContactUserListResponse, error) {
msg, list, ret := gorm.UserContactService.GetUserList(req.OwnerId)
pbList := make([]*pb.MyUserItem, 0, len(list))
for _, item := range list {
pbList = append(pbList, &pb.MyUserItem{
UserId:   item.UserId,
UserName: item.UserName,
Avatar:   item.Avatar,
})
}
return &pb.GetContactUserListResponse{
Message: msg,
List:    pbList,
Ret:     int32(ret),
}, nil
}

func (s *contactQueryServer) LoadMyJoinedGroup(ctx context.Context, req *pb.LoadMyJoinedGroupRequest) (*pb.LoadMyJoinedGroupResponse, error) {
msg, list, ret := gorm.UserContactService.LoadMyJoinedGroup(req.OwnerId)
pbList := make([]*pb.JoinedGroupItem, 0, len(list))
for _, item := range list {
pbList = append(pbList, &pb.JoinedGroupItem{
GroupId:   item.GroupId,
GroupName: item.GroupName,
Avatar:    item.Avatar,
})
}
return &pb.LoadMyJoinedGroupResponse{
Message: msg,
List:    pbList,
Ret:     int32(ret),
}, nil
}

func (s *contactQueryServer) GetContactInfo(ctx context.Context, req *pb.GetContactInfoRequest) (*pb.GetContactInfoResponse, error) {
msg, info, ret := gorm.UserContactService.GetContactInfo(req.ContactId)
pbInfo := &pb.ContactInfoDetail{
ContactId:        info.ContactId,
ContactName:      info.ContactName,
ContactAvatar:    info.ContactAvatar,
ContactPhone:     info.ContactPhone,
ContactEmail:     info.ContactEmail,
ContactGender:    int32(info.ContactGender),
ContactSignature: info.ContactSignature,
ContactBirthday:  info.ContactBirthday,
ContactNotice:    info.ContactNotice,
ContactMembers:   info.ContactMembers,
ContactMemberCnt: int32(info.ContactMemberCnt),
ContactOwnerId:   info.ContactOwnerId,
ContactAddMode:   int32(info.ContactAddMode),
}
return &pb.GetContactInfoResponse{
Message: msg,
Info:    pbInfo,
Ret:     int32(ret),
}, nil
}

func (s *contactQueryServer) GetNewContactList(ctx context.Context, req *pb.GetNewContactListRequest) (*pb.GetNewContactListResponse, error) {
msg, list, ret := gorm.UserContactService.GetNewContactList(req.OwnerId)
pbList := make([]*pb.NewContactItem, 0, len(list))
for _, item := range list {
pbList = append(pbList, &pb.NewContactItem{
ContactId:     item.ContactId,
ContactName:   item.ContactName,
ContactAvatar: item.ContactAvatar,
Msg:           item.Message,
})
}
return &pb.GetNewContactListResponse{
Message: msg,
List:    pbList,
Ret:     int32(ret),
}, nil
}

func (s *contactQueryServer) GetAddGroupList(ctx context.Context, req *pb.GetAddGroupListRequest) (*pb.GetAddGroupListResponse, error) {
msg, list, ret := gorm.UserContactService.GetAddGroupList(req.GroupId)
pbList := make([]*pb.AddGroupItem, 0, len(list))
for _, item := range list {
pbList = append(pbList, &pb.AddGroupItem{
ContactId:     item.ContactId,
ContactName:   item.ContactName,
ContactAvatar: item.ContactAvatar,
Msg:           item.Message,
})
}
return &pb.GetAddGroupListResponse{
Message: msg,
List:    pbList,
Ret:     int32(ret),
}, nil
}

// ==================== Server Startup ====================

var startGRPCServerOnce sync.Once

func ensureGRPCServer(listenAddress string) error {
if listenAddress == "" {
return fmt.Errorf("query: grpc listen address is empty")
}

var initErr error
startGRPCServerOnce.Do(func() {
listener, err := net.Listen("tcp", listenAddress)
if err != nil {
initErr = err
return
}

grpcServer := grpc.NewServer()
pb.RegisterMessageQueryServiceServer(grpcServer, &messageQueryServer{})
pb.RegisterSessionQueryServiceServer(grpcServer, &sessionQueryServer{})
pb.RegisterUserQueryServiceServer(grpcServer, &userQueryServer{})
pb.RegisterGroupQueryServiceServer(grpcServer, &groupQueryServer{})
pb.RegisterContactQueryServiceServer(grpcServer, &contactQueryServer{})

go func() {
log.Printf("query grpc server listening on %s", listenAddress)
if serveErr := grpcServer.Serve(listener); serveErr != nil {
log.Printf("query grpc server stopped: %v", serveErr)
}
}()
})
return initErr
}
