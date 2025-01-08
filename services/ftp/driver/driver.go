package driver

import (
	"errors"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
)

// fsPerm结构体用于处理文件系统权限相关操作，实现server.Perm接口
type fsPerm struct {
	root string
}

// path辅助方法，拼接根路径与给定名称
func (f *fsPerm) path(name string) string {
	return filepath.Join(f.root, name)
}

// GetOwner获取文件所有者，这里简单返回错误，可按需完善
func (f *fsPerm) GetOwner(name string) (string, error) {
	return "", errors.New("获取所有者功能未完整实现")
}

// GetGroup获取文件所属组，这里简单返回错误，可按需完善
func (f *fsPerm) GetGroup(name string) (string, error) {
	return "", errors.New("获取所属组功能未完整实现")
}

// GetMode获取文件模式
func (f *fsPerm) GetMode(name string) (os.FileMode, error) {
	info, err := os.Stat(f.path(name))
	if err != nil {
		return 0, err
	}
	return info.Mode(), nil
}

// ChOwner更改文件所有者
func (f *fsPerm) ChOwner(name string, owner string) error {
	user, err := user.Lookup(owner)
	if err != nil {
		return err
	}
	uid, err := strconv.Atoi(user.Uid)
	if err != nil {
		return err
	}
	return os.Chown(f.path(name), uid, -1)
}

// ChGroup更改文件所属组
func (f *fsPerm) ChGroup(name string, groupName string) error {
	group, err := user.Lookup(groupName)
	if err != nil {
		return err
	}
	gid, err := strconv.Atoi(group.Gid)
	if err != nil {
		return err
	}
	return os.Chown(f.path(name), -1, gid)
}

// ChMode更改文件模式
func (f *fsPerm) ChMode(name string, mode os.FileMode) error {
	return os.Chmod(f.path(name), mode)
}

// newFsPerm 创建一个用于权限管理的fsPerm实例
func NewFsPerm(root string) *fsPerm {
	return &fsPerm{
		root: root,
	}
}
