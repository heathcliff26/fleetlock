package errors

import "fmt"

type ErrorUnknownGroup struct {
	group string
}

func NewErrorUnknownGroup(group string) error {
	return &ErrorUnknownGroup{group: group}
}

func (e *ErrorUnknownGroup) Error() string {
	return fmt.Sprintf("Unknown group: %s", e.group)
}

type ErrorEmptyID struct{}

func (e ErrorEmptyID) Error() string {
	return "Received empty id, can't reserve a slot without an id"
}

type ErrorUnkownStorageType struct {
	Type string
}

func NewErrorUnkownStorageType(t string) error {
	return &ErrorUnkownStorageType{
		Type: t,
	}
}

func (e *ErrorUnkownStorageType) Error() string {
	return fmt.Sprintf("Unsupported storage type \"%s\" selected", e.Type)
}

type ErrorGroupSlotsOutOfRange struct{}

func NewErrorGroupSlotsOutOfRange() error {
	return ErrorGroupSlotsOutOfRange{}
}

func (e ErrorGroupSlotsOutOfRange) Error() string {
	return "At least one group has not enough slots, need at least 1"
}
