package userdata

import (
	"net"
)

type ChatUser struct {
	Name *string
	Conn *net.Conn
}

type ChatGroup struct {
	Name  *string
	Users []*ChatUser
}

type ChatUserData struct {
	userMap  map[string]*ChatUser
	groupMap map[string]*ChatGroup
}

func NewChatUserData() *ChatUserData {
	return &ChatUserData{
		userMap:  make(map[string]*ChatUser),
		groupMap: make(map[string]*ChatGroup),
	}
}

func (userdata *ChatUserData) AddUser(name *string, conn *net.Conn) *ChatUser {
	/* do we contain the same user? */
	if userdata.userMap[*name] != nil {
		return nil
	}

	user := ChatUser{
		Name: name,
		Conn: conn,
	}

	userdata.userMap[*name] = &user
	return &user
}

func (userdata *ChatUserData) ReoveUser(name *string) *ChatUser {
	user := userdata.userMap[*name]
	delete(userdata.userMap, *name)

	for groupName, _ := range userdata.groupMap {
		userdata.RemoveUserFromGroup(user, &groupName)
	}

	return user
}

func (userdata *ChatUserData) GetUser(name *string) *ChatUser {
	return userdata.userMap[*name]
}

func (userdata *ChatUserData) GetAllUsers() []*ChatUser {
	users := make([]*ChatUser, len(userdata.userMap))

	index := 0
	for _, user := range userdata.userMap {
		users[index] = user
		index++
	}

	return users
}

func (userdata *ChatUserData) ReNameUser(newName *string, user *ChatUser) *ChatUser {
	if userdata.userMap[*newName] != nil {
		return nil
	}

	delete(userdata.userMap, *user.Name)

	user.Name = newName
	userdata.userMap[*newName] = user

	return user
}

func (userdata *ChatUserData) AddUserToGroup(user *ChatUser, groupName *string) *ChatGroup {
	group := userdata.groupMap[*groupName]

	/* if group  */
	if group == nil {
		group = &ChatGroup{
			Name:  groupName,
			Users: make([]*ChatUser, 0, 5),
		}

		userdata.groupMap[*groupName] = group
	}

	/* user has already in group */
	for _, gUser := range group.Users {
		if gUser == user {
			return group
		}
	}

	group.Users = append(group.Users, user)
	return group
}

func (userdata *ChatUserData) RemoveUserFromGroup(user *ChatUser, groupName *string) *ChatGroup {
	group := userdata.groupMap[*groupName]

	if group != nil {
		for index, groupUser := range group.Users {
			if groupUser == user {
				if index+1 == len(group.Users) {
					group.Users = group.Users[:index]
				} else {
					group.Users = append(group.Users[:index], group.Users[index+1:]...)
				}
				break
			}
		}
	}

	if group != nil && len(group.Users) == 0 {
		delete(userdata.groupMap, *groupName)
	}

	return group
}

func (userdata *ChatUserData) GetUsersFromGroup(groupName *string) []*ChatUser {
	group := userdata.groupMap[*groupName]
	if group != nil {
		return group.Users
	}

	return nil
}

func (userdata *ChatUserData) GetGroup(groupName *string) *ChatGroup {
	return userdata.groupMap[*groupName]
}

func (userdata *ChatUserData) GetAllGroups() []*ChatGroup {
	groups := make([]*ChatGroup, len(userdata.groupMap))

	index := 0
	for _, group := range userdata.groupMap {
		groups[index] = group
		index++
	}

	return groups
}
