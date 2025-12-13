package core

import (
	"strconv"

	"github.com/viperadnan-git/go-gpm/internal/pb"
)

// PerformTrashAction performs a trash operation on items
// itemKeys can be either mediaKeys or dedupKeys (URL-safe base64 encoded SHA1 hashes)
// actionType specifies the operation: MOVE_TO_TRASH, PERMANENT_DELETE, or RESTORE_FROM_TRASH
// This is the main function that can be used directly for any trash operation
func (a *Api) PerformTrashAction(itemKeys []string, actionType pb.TrashActionType) error {
	var field4 int64
	var field8 *pb.TrashAction_Field8
	var field9 *pb.TrashAction_Field9

	switch actionType {
	case pb.TrashActionType_MOVE_TO_TRASH:
		field4 = 1
		field8 = &pb.TrashAction_Field8{
			Field4: &pb.TrashAction_Field8_Field4{
				Field2: &pb.TrashAction_Field8_Field4_Empty{},
				Field3: &pb.TrashAction_Field8_Field4_Field3{
					Field1: &pb.TrashAction_Field8_Field4_Empty{},
				},
				Field4: &pb.TrashAction_Field8_Field4_Empty{},
				Field5: &pb.TrashAction_Field8_Field4_Field5{
					Field1: &pb.TrashAction_Field8_Field4_Empty{},
				},
			},
		}
		field9 = &pb.TrashAction_Field9{
			Field1: 5,
			Field2: &pb.TrashAction_Field9_Field2{
				Field1: a.ClientVersionCode,
				Field2: strconv.FormatInt(a.AndroidAPIVersion, 10),
			},
		}

	case pb.TrashActionType_PERMANENT_DELETE:
		field4 = 1
		field8 = &pb.TrashAction_Field8{
			Field4: &pb.TrashAction_Field8_Field4{
				Field2: &pb.TrashAction_Field8_Field4_Empty{},
				Field3: &pb.TrashAction_Field8_Field4_Field3{
					Field1: &pb.TrashAction_Field8_Field4_Empty{},
				},
				Field4: &pb.TrashAction_Field8_Field4_Empty{},
				Field5: &pb.TrashAction_Field8_Field4_Field5{
					Field1: &pb.TrashAction_Field8_Field4_Empty{},
				},
			},
		}
		// field9 is empty for permanent delete

	case pb.TrashActionType_RESTORE_FROM_TRASH:
		field4 = 2
		field8 = &pb.TrashAction_Field8{
			Field4: &pb.TrashAction_Field8_Field4{
				Field2: &pb.TrashAction_Field8_Field4_Empty{},
				Field3: &pb.TrashAction_Field8_Field4_Field3{
					Field1: &pb.TrashAction_Field8_Field4_Empty{},
				},
			},
		}
		field9 = &pb.TrashAction_Field9{
			Field1: 5,
			Field2: &pb.TrashAction_Field9_Field2{
				Field1: a.ClientVersionCode,
				Field2: strconv.FormatInt(a.AndroidAPIVersion, 10),
			},
		}
	}

	requestBody := pb.TrashAction{
		ActionType: actionType,
		ItemKeys:   itemKeys,
		Field4:     field4,
		Field8:     field8,
		Field9:     field9,
	}

	return a.DoProtoRequest(
		"https://photosdata-pa.googleapis.com/6439526531001121323/17490284929287180316",
		&requestBody,
		nil,
		WithAuth(),
		WithCommonHeaders(),
		WithStatusCheck(),
	)
}

// MoveToTrash moves items to trash (wrapper around PerformTrashAction)
// itemKeys can be either mediaKeys or dedupKeys (URL-safe base64 encoded SHA1 hashes)
func (a *Api) MoveToTrash(itemKeys []string) error {
	return a.PerformTrashAction(itemKeys, pb.TrashActionType_MOVE_TO_TRASH)
}

// RestoreFromTrash restores items from trash (wrapper around PerformTrashAction)
// itemKeys can be either mediaKeys or dedupKeys (URL-safe base64 encoded SHA1 hashes)
func (a *Api) RestoreFromTrash(itemKeys []string) error {
	return a.PerformTrashAction(itemKeys, pb.TrashActionType_RESTORE_FROM_TRASH)
}

// PermanentDelete permanently deletes items (wrapper around PerformTrashAction)
// itemKeys can be either mediaKeys or dedupKeys (URL-safe base64 encoded SHA1 hashes)
// Items will be permanently deleted immediately, bypassing trash
func (a *Api) PermanentDelete(itemKeys []string) error {
	return a.PerformTrashAction(itemKeys, pb.TrashActionType_PERMANENT_DELETE)
}
