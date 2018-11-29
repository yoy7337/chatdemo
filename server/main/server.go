package main

import (
	"chatdemo/server/userdata"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

var userData *userdata.ChatUserData

// var userMap map[string]*userdata.ChatUser

func main() {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", ":8081")
	checkError(err)

	ln, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	// userMap = make(map[string]*userdata.ChatUser)
	userData = userdata.NewChatUserData()

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error on create connection", err)
			continue
		}

		/* generate chat user */
		username := fmt.Sprintf("user_%d", time.Now().UnixNano())

		user := userData.AddUser(&username, &conn)

		/* handle user */
		go chatroomServ(user)
	}
}

func test() {
	mySlice := []int{1, 4, 293, 4, 9}
	mySlice = append(mySlice, 1)
	fmt.Println("len=", len(mySlice), "cap=", cap(mySlice), "mySlice=", mySlice)
	mySlice = append(mySlice[:2], mySlice[3:]...)
	fmt.Println("len=", len(mySlice), "cap=", cap(mySlice), "mySlice=", mySlice)
}

func chatroomServ(user *userdata.ChatUser) {
	defer (*user.Conn).Close()

	welcome := fmt.Sprintf("Hello %s, welcome entering chat server...", *user.Name)
	(*user.Conn).Write([]byte(welcome))

	for {
		cmd, err := readCommand(user)
		if err != nil {
			fmt.Println("chatroomServ err: ", err)

			// disconnect and remove user data
			userData.ReoveUser(user.Name)
			break
		}

		err = handleCmd(cmd, user)
		if err != nil {
			// send error message to user
			(*user.Conn).Write([]byte(err.Error()))

			// disconnect and remove user data
			userData.ReoveUser(user.Name)
			break
		}
	}
}

func handleCmd(cmd *string, user *userdata.ChatUser) error {
	cmdArgs := strings.Split(*cmd, " ")
	if len(cmdArgs) == 0 {
		(*user.Conn).Write([]byte("Please input command or use 'help' to find the usage"))
		return nil
	}

	switch cmdArgs[0] {
	case "help":
		handleHelpCmd(user)
	case "exit":
		return errors.New("Client disconnect")
	case "ls-user":
		handleLsUserCmd(user)
	case "ls-group":
		handleLsGroupCmd(cmdArgs, user)
	case "name":
		handleNameUser(cmdArgs, user)
	case "sendmsg":
		handleSendMsg(cmdArgs, user)
	case "sendbmsg":
		handleBroadcastMsg(cmdArgs, user)
	case "group-add":
		handleAddUserToGroup(cmdArgs, user)
	case "group-rm":
		handleRemoveUserFromGroup(cmdArgs, user)
	case "sendgmsg":
		handleSendMsgToGroup(cmdArgs, user)
	default:
		(*user.Conn).Write([]byte("Please input command or use 'help' to find the usage"))
	}

	return nil
}

func readCommand(user *userdata.ChatUser) (*string, error) {
	buf := make([]byte, 512)
	var readBuf [1]byte
	var offset = 0

	/* read data until newline */
	for {
		n, err := (*user.Conn).Read(readBuf[0:1])
		if n != 1 || err != nil {
			return nil, err
		}

		if readBuf[0] == '\n' {
			cmd := string(buf[0:offset])
			return &cmd, nil
		}

		if offset == cap(buf) {
			buf = append(buf, readBuf[0])
		} else {
			buf[offset] = readBuf[0]
		}

		offset++
	}

}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func handleHelpCmd(user *userdata.ChatUser) {
	output := fmt.Sprint(
		"usage: <command> [arguments]\n\n",
		"commands:\n",
		"help\n",
		"exit\n",
		"name              <my name>\n",
		"group-add         <group name>      <user name>\n",
		"group-rm          <group name>      <user name>\n",
		"ls-user\n",
		"ls-group        [group name]\n",
		"sendmsg           <uer name>        <message>\n",
		"sendgmsg          <group name>      <message>\n",
		"sendbmsg          <message>\n")
	(*user.Conn).Write([]byte(output))
}

func handleLsUserCmd(currentUser *userdata.ChatUser) {
	var str strings.Builder
	str.WriteString("list all users...\n")

	users := userData.GetAllUsers()

	for _, user := range users {
		if currentUser == user {
			str.WriteString("*")
		} else {
			str.WriteString(" ")
		}

		str.WriteString(*user.Name)
		str.WriteString("\n")
	}

	(*currentUser.Conn).Write([]byte(str.String()))
}

func handleLsGroupCmd(cmdArgs []string, currentUser *userdata.ChatUser) {
	var str strings.Builder

	if len(cmdArgs) >= 2 {
		groupName := cmdArgs[1]
		str.WriteString("list all user in group(")
		str.WriteString(groupName)
		str.WriteString(")...\n")

		group := userData.GetGroup(&groupName)
		if group == nil {
			str.WriteString("Can not find group...\n")
		} else {
			users := group.Users
			for _, user := range users {
				str.WriteString(*user.Name)
				str.WriteString("\n")
			}
		}
	} else {
		str.WriteString("list all groups...\n")

		groups := userData.GetAllGroups()

		for _, group := range groups {
			str.WriteString(" ")
			str.WriteString(*group.Name)
			str.WriteString("\n")
		}
	}

	(*currentUser.Conn).Write([]byte(str.String()))
}

func handleNameUser(cmdArgs []string, currentUser *userdata.ChatUser) {
	if len(cmdArgs) != 2 {
		(*currentUser.Conn).Write([]byte("Usege: name <your new name>"))
		return
	}

	newName := cmdArgs[1]
	if userData.ReNameUser(&newName, currentUser) == nil {
		(*currentUser.Conn).Write([]byte("This name has been used..."))
		return
	}

	(*currentUser.Conn).Write([]byte("Change name success..."))
}

func handleSendMsg(cmdArgs []string, currentUser *userdata.ChatUser) {
	if len(cmdArgs) < 3 {
		(*currentUser.Conn).Write([]byte("Usege: sendmsg <user name> <message>"))
		return
	}

	targetUserName := cmdArgs[1]
	targetUser := userData.GetUser(&targetUserName)
	msg := strings.Join(cmdArgs[2:], " ")

	if targetUser == nil {
		resp := fmt.Sprintf("Can not send message to %s", targetUserName)
		(*currentUser.Conn).Write([]byte(resp))
		return
	}

	forwardMsg := fmt.Sprintf("[%s]: %s", *currentUser.Name, msg)
	(*targetUser.Conn).Write([]byte(forwardMsg))

	(*currentUser.Conn).Write([]byte("Send message success..."))
}

func handleBroadcastMsg(cmdArgs []string, currentUser *userdata.ChatUser) {
	if len(cmdArgs) < 2 {
		(*currentUser.Conn).Write([]byte("Usege: sendbmsg <message>"))
		return
	}

	msg := strings.Join(cmdArgs[1:], " ")
	forwardMsg := fmt.Sprintf("[Broadcast][%s]: %s", *currentUser.Name, msg)

	users := userData.GetAllUsers()
	for _, user := range users {
		if user == currentUser {
			continue
		}

		(*user.Conn).Write([]byte(forwardMsg))
	}

	(*currentUser.Conn).Write([]byte("Send message success..."))
}

func handleAddUserToGroup(cmdArgs []string, currentUser *userdata.ChatUser) {
	if len(cmdArgs) < 3 {
		(*currentUser.Conn).Write([]byte("Usege: group-add <group name> <user name>"))
		return
	}

	groupName := cmdArgs[1]
	userName := cmdArgs[2]

	user := userData.GetUser(&userName)
	if user == nil {
		(*currentUser.Conn).Write([]byte("Can not find user..."))
		return
	}

	group := userData.AddUserToGroup(user, &groupName)
	if group == nil {
		(*currentUser.Conn).Write([]byte("Can not add user to group..."))
		return
	}

	(*currentUser.Conn).Write([]byte("Add user to group success\n"))
	handleLsGroupCmd([]string{"ls-group", groupName}, currentUser)
}

func handleRemoveUserFromGroup(cmdArgs []string, currentUser *userdata.ChatUser) {
	if len(cmdArgs) < 3 {
		(*currentUser.Conn).Write([]byte("Usege: group-rm <group name> <user name>"))
		return
	}

	groupName := cmdArgs[1]
	userName := cmdArgs[2]
	user := userData.GetUser(&userName)
	if user == nil {
		(*currentUser.Conn).Write([]byte("Can not find user..."))
		return
	}

	group := userData.RemoveUserFromGroup(user, &groupName)
	if group == nil {
		(*currentUser.Conn).Write([]byte("Can not add user to group..."))
		return
	}

	(*currentUser.Conn).Write([]byte("Add user to group success\n"))
	handleLsGroupCmd([]string{"ls-group", groupName}, currentUser)
}

func handleSendMsgToGroup(cmdArgs []string, currentUser *userdata.ChatUser) {
	if len(cmdArgs) < 3 {
		(*currentUser.Conn).Write([]byte("Usege: sendgmsg <group name> <message>"))
		return
	}

	groupName := cmdArgs[1]
	msg := strings.Join(cmdArgs[2:], " ")
	forwardMsg := fmt.Sprintf("[Group %s][%s]: %s", groupName, *currentUser.Name, msg)

	users := userData.GetUsersFromGroup(&groupName)
	if users == nil {
		(*currentUser.Conn).Write([]byte("Can not find group..."))
		return
	}

	for _, user := range users {
		if user == currentUser {
			continue
		}

		(*user.Conn).Write([]byte(forwardMsg))
	}

	(*currentUser.Conn).Write([]byte("Send group message success..."))
}
