package service

import (
	"github.com/Pivot-Studio/pivot-chat/dao"
	"github.com/sirupsen/logrus"
	"sync"
)

var GroupManager sync.Map

func SetGroup(groupID int64, g *Group) {
	GroupManager.Store(groupID, g)
}

func GetGroup(groupID int64) *Group {
	value, ok := GroupManager.Load(groupID)
	if ok {
		return value.(*Group)
	}
	return nil
}

func DeleteGroup(groupID int64) {
	GroupManager.Delete(groupID)
}

// GetUpdatedGroup 得到最新的group
func GetUpdatedGroup(groupID int64) *Group {
	//总是保持除member之外的数据与数据库中的相同
	groupDb, err := dao.RS.QueryGroup(groupID)
	if err != nil {
		logrus.Fatalf("[GetUpdatedGroup] QueryGroup %+v", err)
		return nil
	}
	groupMap := GetGroup(groupID)
	if groupMap != nil {
		groupDb.Members = groupMap.Members
	}
	group := &Group{}
	group.Group = groupDb

	if group.UserNum != int32(len(*group.Members)) {
		//不相等, 需要更新
		members, err := dao.RS.GetGroupUsers(groupID)
		if err != nil {
			logrus.Fatalf("[UpdateGroup] GetGroupUsers %+v", err)
			return group
		}
		group.Members = members
	}
	//更新map中的group
	SetGroup(groupID, group)
	return group
}
