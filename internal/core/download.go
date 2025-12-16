package core

import (
	"context"

	"github.com/viperadnan-git/go-gpm/internal/pb"
)

// GetDownloadUrl gets the download URL for a media item
// Returns downloadURL and isEdited (true if the URL is for an edited version)
func (a *Api) GetDownloadUrl(ctx context.Context, mediaKey string) (downloadURL string, isEdited bool, err error) {
	requestBody := pb.GetDownloadUrl{
		Field1: &pb.GetDownloadUrl_Field1{
			Field1: &pb.GetDownloadUrl_Field1_Field1Inner{
				MediaKey: mediaKey,
			},
		},
		Field2: &pb.GetDownloadUrl_Field2{
			Field1: &pb.GetDownloadUrl_Field2_Field1Type{
				Field7: &pb.GetDownloadUrl_Field2_Field1Type_Field7Type{
					Field2: &pb.GetDownloadUrl_Field2_Field1Type_Field7Type_Field2Type{},
				},
			},
			Field5: &pb.GetDownloadUrl_Field2_Field5Type{
				Field2: &pb.GetDownloadUrl_Field2_Field5Type_Field2Type{},
				Field3: &pb.GetDownloadUrl_Field2_Field5Type_Field3Type{},
				Field5: &pb.GetDownloadUrl_Field2_Field5Type_Field5Inner{
					Field1: &pb.GetDownloadUrl_Field2_Field5Type_Field5Inner_Field1Type{},
					Field3: 0,
				},
			},
		},
	}

	var response pb.GetDownloadUrlResponse
	if err := a.DoProtoRequest(
		ctx,
		"https://photosdata-pa.googleapis.com/$rpc/social.frontend.photos.preparedownloaddata.v1.PhotosPrepareDownloadDataService/PhotosPrepareDownload",
		&requestBody,
		&response,
		WithAuth(),
		WithCommonHeaders(),
		WithStatusCheck(),
	); err != nil {
		return "", false, err
	}

	if response.GetField1() != nil && response.GetField1().GetField5() != nil {
		isEdited = response.GetField1().GetField5().GetField1() > 0
		// Try field2 first (primary location), then field3 (fallback)
		if response.GetField1().GetField5().GetField2() != nil {
			downloadURL = response.GetField1().GetField5().GetField2().GetDownloadUrl()
		} else if response.GetField1().GetField5().GetField3() != nil {
			downloadURL = response.GetField1().GetField5().GetField3().GetDownloadUrl()
		}
	}

	return downloadURL, isEdited, nil
}
