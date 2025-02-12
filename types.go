// types.go - libalpm types.
//
// Copyright (c) 2013 The go-alpm Authors
//
// MIT Licensed. See LICENSE for details.

package alpm

// #cgo CFLAGS: -D_FILE_OFFSET_BITS=64
// #include <alpm.h>
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
)

// Depend provides a description of a dependency.
type Depend struct {
	Name        string
	Version     string
	Description string
	NameHash    uint
	Mod         DepMod
}

func convertDepend(dep *C.alpm_depend_t) Depend {
	return Depend{
		Name:        C.GoString(dep.name),
		Version:     C.GoString(dep.version),
		Mod:         DepMod(dep.mod),
		Description: C.GoString(dep.desc),
		NameHash:    uint(dep.name_hash),
	}
}

func convertCDepend(dep Depend) *C.alpm_depend_t {
	cName := C.CString(dep.Name)
	cVersion := C.CString(dep.Version)
	cDesc := C.CString(dep.Description)

	cDep := C.alpm_depend_t{
		name:      cName,
		version:   cVersion,
		desc:      cDesc,
		name_hash: C.ulong(dep.NameHash),
		mod:       C.alpm_depmod_t(dep.Mod),
	}

	return &cDep
}

func freeCDepend(dep *C.alpm_depend_t) {
	C.free(unsafe.Pointer(dep.name))
	C.free(unsafe.Pointer(dep.version))
	C.free(unsafe.Pointer(dep.desc))
}

func (dep Depend) String() string {
	return dep.Name + dep.Mod.String() + dep.Version
}

// File provides a description of package files.
type File struct {
	Name string
	Size int64
	Mode uint32
}

func convertFile(file *C.alpm_file_t) (File, error) {
	if file == nil {
		return File{}, errors.New("no file")
	}

	return File{
		Name: C.GoString(file.name),
		Size: int64(file.size),
		Mode: uint32(file.mode),
	}, nil
}

func convertFilelist(files *C.alpm_filelist_t) []File {
	size := int(files.count)
	items := make([]File, size)

	cFiles := unsafe.Slice(files.files, size)

	for i := 0; i < size; i++ {
		if file, err := convertFile(&cFiles[i]); err == nil {
			items[i] = file
		}
	}

	return items
}

// Internal alpm list structure.
type list struct {
	Data unsafe.Pointer
	Prev *list
	Next *list
}

// Iterates a function on a list and stop on error.
func (l *list) forEach(f func(unsafe.Pointer) error) error {
	for ; l != nil; l = l.Next {
		err := f(l.Data)
		if err != nil {
			return err
		}
	}

	return nil
}

func (l *list) Len() int {
	count := 0
	for ; l != nil; l = l.Next {
		count++
	}

	return count
}

func (l *list) Empty() bool {
	return l == nil
}

type StringList struct {
	*list
}

func (l StringList) ForEach(f func(string) error) error {
	return l.forEach(func(p unsafe.Pointer) error {
		return f(C.GoString((*C.char)(p)))
	})
}

func (l StringList) Slice() []string {
	slice := []string{}
	_ = l.ForEach(func(s string) error {
		slice = append(slice, s)
		return nil
	})

	return slice
}

type BackupFile struct {
	Name string
	Hash string
}

type BackupList struct {
	*list
}

func (l BackupList) ForEach(f func(BackupFile) error) error {
	return l.forEach(func(p unsafe.Pointer) error {
		bf := (*C.alpm_backup_t)(p)
		return f(BackupFile{
			Name: C.GoString(bf.name),
			Hash: C.GoString(bf.hash),
		})
	})
}

func (l BackupList) Slice() (slice []BackupFile) {
	_ = l.ForEach(func(f BackupFile) error {
		slice = append(slice, f)
		return nil
	})

	return
}

type QuestionAny struct {
	ptr *C.alpm_question_any_t
}

func (question QuestionAny) SetAnswer(answer bool) {
	if answer {
		question.ptr.answer = 1
	} else {
		question.ptr.answer = 0
	}
}

type QuestionInstallIgnorepkg struct {
	ptr *C.alpm_question_install_ignorepkg_t
}

func (question QuestionAny) Type() QuestionType {
	return QuestionType(question.ptr._type)
}

func (question QuestionAny) Answer() bool {
	return question.ptr.answer == 1
}

func (question QuestionAny) QuestionInstallIgnorepkg() (QuestionInstallIgnorepkg, error) {
	if question.Type() == QuestionTypeInstallIgnorepkg {
		return *(*QuestionInstallIgnorepkg)(unsafe.Pointer(&question)), nil
	}

	return QuestionInstallIgnorepkg{}, fmt.Errorf("cannot convert to QuestionInstallIgnorepkg")
}

func (question QuestionAny) QuestionSelectProvider() (QuestionSelectProvider, error) {
	if question.Type() == QuestionTypeSelectProvider {
		return *(*QuestionSelectProvider)(unsafe.Pointer(&question)), nil
	}

	return QuestionSelectProvider{}, fmt.Errorf("cannot convert to QuestionInstallIgnorepkg")
}

func (question QuestionAny) QuestionReplace() (QuestionReplace, error) {
	if question.Type() == QuestionTypeReplacePkg {
		return *(*QuestionReplace)(unsafe.Pointer(&question)), nil
	}

	return QuestionReplace{}, fmt.Errorf("cannot convert to QuestionReplace")
}

func (question QuestionInstallIgnorepkg) SetInstall(install bool) {
	if install {
		question.ptr.install = 1
	} else {
		question.ptr.install = 0
	}
}

func (question QuestionInstallIgnorepkg) Type() QuestionType {
	return QuestionType(question.ptr._type)
}

func (question QuestionInstallIgnorepkg) Install() bool {
	return question.ptr.install == 1
}

func (question QuestionInstallIgnorepkg) Pkg(h *Handle) Package {
	return Package{
		question.ptr.pkg,
		*h,
	}
}

type QuestionReplace struct {
	ptr *C.alpm_question_replace_t
}

func (question QuestionReplace) Type() QuestionType {
	return QuestionType(question.ptr._type)
}

func (question QuestionReplace) SetReplace(replace bool) {
	if replace {
		question.ptr.replace = 1
	} else {
		question.ptr.replace = 0
	}
}

func (question QuestionReplace) Replace() bool {
	return question.ptr.replace == 1
}

func (question QuestionReplace) NewPkg(h *Handle) *Package {
	return &Package{
		question.ptr.newpkg,
		*h,
	}
}

func (question QuestionReplace) OldPkg(h *Handle) *Package {
	return &Package{
		question.ptr.oldpkg,
		*h,
	}
}

type QuestionSelectProvider struct {
	ptr *C.alpm_question_select_provider_t
}

func (question QuestionSelectProvider) Type() QuestionType {
	return QuestionType(question.ptr._type)
}

func (question QuestionSelectProvider) SetUseIndex(index int) {
	question.ptr.use_index = C.int(index)
}

func (question QuestionSelectProvider) UseIndex() int {
	return int(question.ptr.use_index)
}

func (question QuestionSelectProvider) Providers(h *Handle) PackageList {
	return PackageList{
		(*list)(unsafe.Pointer(question.ptr.providers)),
		*h,
	}
}

func (question QuestionSelectProvider) Dep() Depend {
	return convertDepend(question.ptr.depend)
}
