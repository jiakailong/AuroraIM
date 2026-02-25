package query

import (
	"kama_chat_server/internal/dto/respond"
	"kama_chat_server/internal/service/gorm"
)

type localService struct{}

func newLocalService() Service {
	return &localService{}
}

func (service *localService) GetMessageList(userOneId, userTwoId string) (string, []respond.GetMessageListRespond, int) {
	return gorm.MessageService.GetMessageList(userOneId, userTwoId)
}

func (service *localService) GetGroupMessageList(groupId string) (string, []respond.GetGroupMessageListRespond, int) {
	return gorm.MessageService.GetGroupMessageList(groupId)
}

func (service *localService) GetUserSessionList(ownerId string) (string, []respond.UserSessionListRespond, int) {
	return gorm.SessionService.GetUserSessionList(ownerId)
}

func (service *localService) GetGroupSessionList(ownerId string) (string, []respond.GroupSessionListRespond, int) {
	return gorm.SessionService.GetGroupSessionList(ownerId)
}

func (service *localService) GetUserInfoList(ownerId string) (string, []respond.GetUserListRespond, int) {
	return gorm.UserInfoService.GetUserInfoList(ownerId)
}

func (service *localService) GetUserInfo(uuid string) (string, *respond.GetUserInfoRespond, int) {
	return gorm.UserInfoService.GetUserInfo(uuid)
}

func (service *localService) LoadMyGroup(ownerId string) (string, []respond.LoadMyGroupRespond, int) {
	return gorm.GroupInfoService.LoadMyGroup(ownerId)
}

func (service *localService) CheckGroupAddMode(groupId string) (string, int8, int) {
	return gorm.GroupInfoService.CheckGroupAddMode(groupId)
}

func (service *localService) GetGroupInfo(groupId string) (string, *respond.GetGroupInfoRespond, int) {
	return gorm.GroupInfoService.GetGroupInfo(groupId)
}

func (service *localService) GetGroupInfoList() (string, []respond.GetGroupListRespond, int) {
	return gorm.GroupInfoService.GetGroupInfoList()
}

func (service *localService) GetGroupMemberList(groupId string) (string, []respond.GetGroupMemberListRespond, int) {
	return gorm.GroupInfoService.GetGroupMemberList(groupId)
}

func (service *localService) GetUserList(ownerId string) (string, []respond.MyUserListRespond, int) {
	return gorm.UserContactService.GetUserList(ownerId)
}

func (service *localService) LoadMyJoinedGroup(ownerId string) (string, []respond.LoadMyJoinedGroupRespond, int) {
	return gorm.UserContactService.LoadMyJoinedGroup(ownerId)
}

func (service *localService) GetContactInfo(contactId string) (string, respond.GetContactInfoRespond, int) {
	return gorm.UserContactService.GetContactInfo(contactId)
}

func (service *localService) GetNewContactList(ownerId string) (string, []respond.NewContactListRespond, int) {
	return gorm.UserContactService.GetNewContactList(ownerId)
}

func (service *localService) GetAddGroupList(groupId string) (string, []respond.AddGroupListRespond, int) {
	return gorm.UserContactService.GetAddGroupList(groupId)
}
