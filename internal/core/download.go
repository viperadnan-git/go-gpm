package core

import (
	"context"
	"fmt"

	"github.com/viperadnan-git/go-gpm/internal/pb"
)

// DownloadInfo contains download information for a media item
type DownloadInfo struct {
	Filename    string
	FileSize    int64
	IsEdited    bool
	DownloadURL string // Preferred URL (OriginalURL if available, otherwise EditedURL)
	OriginalURL string
	EditedURL   string
}

// GetDownloadInfo gets the download information for a media item
func (a *Api) GetDownloadInfo(ctx context.Context, mediaKey string) (*DownloadInfo, error) {
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
		return nil, err
	}

	info := &DownloadInfo{}

	if response.GetField1() != nil {
		if response.GetField1().GetMetadata() != nil {
			info.Filename = response.GetField1().GetMetadata().GetFilename()
			info.FileSize = response.GetField1().GetMetadata().GetFileSize()
		}

		if response.GetField1().GetUrls() != nil {
			info.IsEdited = response.GetField1().GetUrls().GetIsEdited() > 0

			if response.GetField1().GetUrls().GetDownloadUrls() != nil {
				info.OriginalURL = response.GetField1().GetUrls().GetDownloadUrls().GetOriginalUrl()
				info.EditedURL = response.GetField1().GetUrls().GetDownloadUrls().GetEditedUrl()
			} else if response.GetField1().GetUrls().GetField3() != nil {
				info.OriginalURL = response.GetField1().GetUrls().GetField3().GetDownloadUrl()
			}
		}
	}

	// Set DownloadURL to preferred URL (original if available, otherwise edited)
	if info.OriginalURL != "" {
		info.DownloadURL = info.OriginalURL
	} else {
		info.DownloadURL = info.EditedURL
	}

	if info.DownloadURL == "" {
		return nil, fmt.Errorf("no download URL available")
	}

	return info, nil
}
