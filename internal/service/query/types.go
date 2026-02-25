package query

import (
	"kama_chat_server/internal/dto/respond"
)

type Service interface {
	GetMessageList(userOneId, userTwoId string) (string, []respond.GetMessageListRespond, int)
	GetGroupMessageList(groupId string) (string, []respond.GetGroupMessageListRespond, int)
	GetUserSessionList(ownerId string) (string, []respond.UserSessionListRespond, int)
	GetGroupSessionList(ownerId string) (string, []respond.GroupSessionListRespond, int)

	GetUserInfoList(ownerId string) (string, []respond.GetUserListRespond, int)
	GetUserInfo(uuid string) (string, *respond.GetUserInfoRespond, int)

	LoadMyGroup(ownerId string) (string, []respond.LoadMyGroupRespond, int)
	CheckGroupAddMode(groupId string) (string, int8, int)
	GetGroupInfo(groupId string) (string, *respond.GetGroupInfoRespond, int)
	GetGroupInfoList() (string, []respond.GetGroupListRespond, int)
	GetGroupMemberList(groupId string) (string, []respond.GetGroupMemberListRespond, int)

	GetUserList(ownerId string) (string, []respond.MyUserListRespond, int)
	LoadMyJoinedGroup(ownerId string) (string, []respond.LoadMyJoinedGroupRespond, int)
	GetContactInfo(contactId string) (string, respond.GetContactInfoRespond, int)
	GetNewContactList(ownerId string) (string, []respond.NewContactListRespond, int)
	GetAddGroupList(groupId string) (string, []respond.AddGroupListRespond, int)
}

type Config struct {
	Mode      string
	RPCListen string
	RPCTarget string
}

var QueryService Service

type messageListResult struct {
	Message string                          `json:"message"`
	List    []respond.GetMessageListRespond `json:"list"`
	Ret     int                             `json:"ret"`
}

type groupMessageListResult struct {
	Message string                               `json:"message"`
	List    []respond.GetGroupMessageListRespond `json:"list"`
	Ret     int                                  `json:"ret"`
}

type userSessionListResult struct {
	Message string                           `json:"message"`
	List    []respond.UserSessionListRespond `json:"list"`
	Ret     int                              `json:"ret"`
}

type groupSessionListResult struct {
	Message string                            `json:"message"`
	List    []respond.GroupSessionListRespond `json:"list"`
	Ret     int                               `json:"ret"`
}

type userInfoListResult struct {
	Message string                       `json:"message"`
	List    []respond.GetUserListRespond `json:"list"`
	Ret     int                          `json:"ret"`
}

type userInfoResult struct {
	Message string                      `json:"message"`
	Info    *respond.GetUserInfoRespond `json:"info"`
	Ret     int                         `json:"ret"`
}

type myGroupListResult struct {
	Message string                       `json:"message"`
	List    []respond.LoadMyGroupRespond `json:"list"`
	Ret     int                          `json:"ret"`
}

type addModeResult struct {
	Message string `json:"message"`
	AddMode int8   `json:"add_mode"`
	Ret     int    `json:"ret"`
}

type groupInfoResult struct {
	Message string                       `json:"message"`
	Info    *respond.GetGroupInfoRespond `json:"info"`
	Ret     int                          `json:"ret"`
}

type groupInfoListResult struct {
	Message string                        `json:"message"`
	List    []respond.GetGroupListRespond `json:"list"`
	Ret     int                           `json:"ret"`
}

type groupMemberListResult struct {
	Message string                              `json:"message"`
	List    []respond.GetGroupMemberListRespond `json:"list"`
	Ret     int                                 `json:"ret"`
}

type userListResult struct {
	Message string                      `json:"message"`
	List    []respond.MyUserListRespond `json:"list"`
	Ret     int                         `json:"ret"`
}

type joinedGroupListResult struct {
	Message string                             `json:"message"`
	List    []respond.LoadMyJoinedGroupRespond `json:"list"`
	Ret     int                                `json:"ret"`
}

type contactInfoResult struct {
	Message string                        `json:"message"`
	Info    respond.GetContactInfoRespond `json:"info"`
	Ret     int                           `json:"ret"`
}

type newContactListResult struct {
	Message string                          `json:"message"`
	List    []respond.NewContactListRespond `json:"list"`
	Ret     int                             `json:"ret"`
}

type addGroupListResult struct {
	Message string                        `json:"message"`
	List    []respond.AddGroupListRespond `json:"list"`
	Ret     int                           `json:"ret"`
}

func init() {
	QueryService = newLocalService()
}
