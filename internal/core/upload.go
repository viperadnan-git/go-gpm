package core

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/viperadnan-git/go-gpm/internal/pb"

	"google.golang.org/protobuf/proto"
)

// GetUploadToken obtains a file upload token from the Google Photos API
func (a *Api) GetUploadToken(ctx context.Context, sha1HashBase64 string, fileSize int64) (string, error) {
	requestBody := pb.GetUploadToken{
		F1:            2,
		F2:            2,
		F3:            1,
		F4:            3,
		FileSizeBytes: fileSize,
	}

	serializedData, err := proto.Marshal(&requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal protobuf: %w", err)
	}

	_, resp, err := a.DoRequest(
		ctx,
		"https://photos.googleapis.com/data/upload/uploadmedia/interactive",
		bytes.NewReader(serializedData),
		WithAuth(),
		WithCommonHeaders(),
		WithStatusCheck(),
		WithHeaders(map[string]string{
			"X-Goog-Hash":             "sha1=" + sha1HashBase64,
			"X-Upload-Content-Length": strconv.Itoa(int(fileSize)),
		}),
	)
	if err != nil {
		return "", err
	}

	uploadToken := resp.Header.Get("X-GUploader-UploadID")
	if uploadToken == "" {
		return "", errors.New("response missing X-GUploader-UploadID header")
	}

	return uploadToken, nil
}

// FindMediaKeyByHash checks the library for existing files with the given hash
func (a *Api) FindMediaKeyByHash(ctx context.Context, sha1Hash []byte) (string, error) {
	requestBody := pb.FindMediaByHashRequest{
		Field1: &pb.FindMediaByHashRequestField1Type{
			Field1: &pb.FindMediaByHashRequestField1TypeField1Type{
				Sha1Hash: sha1Hash,
			},
			Field2: &pb.FindMediaByHashRequestField1TypeField2Type{},
		},
	}

	var response pb.FindMediaByHashResponse
	if err := a.DoProtoRequest(
		ctx,
		"https://photosdata-pa.googleapis.com/6439526531001121323/5084965799730810217",
		&requestBody,
		&response,
		WithAuth(),
		WithCommonHeaders(),
		WithStatusCheck(),
	); err != nil {
		return "", err
	}

	return response.GetMediaKey(), nil
}

// UploadFile uploads a file to Google Photos using the provided upload token
func (a *Api) UploadFile(ctx context.Context, filePath string, uploadToken string) (*pb.CommitToken, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	uploadURL := "https://photos.googleapis.com/data/upload/uploadmedia/interactive?upload_id=" + uploadToken

	bodyBytes, _, err := a.DoRequest(
		ctx,
		uploadURL,
		file,
		WithMethod("PUT"),
		WithAuth(),
		WithCommonHeaders(),
		WithStatusCheck(),
		WithChunkedTransfer(),
	)
	if err != nil {
		return nil, err
	}

	var commitToken pb.CommitToken
	if err := proto.Unmarshal(bodyBytes, &commitToken); err != nil {
		return nil, fmt.Errorf("failed to unmarshal protobuf: %w", err)
	}

	return &commitToken, nil
}

// CommitUpload commits the upload to Google Photos and returns the media key
// qualityStr: "original" or "storage-saver" (empty string uses Api default)
// useQuota: override Api default if true
func (a *Api) CommitUpload(
	ctx context.Context,
	commitToken *pb.CommitToken,
	fileName string,
	sha1Hash []byte,
	uploadTimestamp int64,
	qualityStr string,
	useQuota bool,
) (string, error) {
	if uploadTimestamp == 0 {
		uploadTimestamp = time.Now().Unix()
	}

	// Use defaults from Api if not overridden
	effectiveQuality := qualityStr
	if effectiveQuality == "" {
		effectiveQuality = a.Quality
	}
	effectiveUseQuota := useQuota || a.UseQuota

	// Determine model based on quality and quota settings
	model := a.Model
	var quality int64 = 3 // original
	if effectiveQuality == "storage-saver" {
		quality = 1
		model = "Pixel 2"
	}
	if effectiveUseQuota {
		model = "Pixel 8"
	}

	unknownConstant := int64(46000000)

	requestBody := pb.CommitUpload{
		Field1: &pb.CommitUploadField1Type{
			Field1: &pb.CommitUploadField1TypeField1Type{
				Field1: commitToken.Field1,
				Field2: commitToken.Field2,
			},
			FileName: fileName,
			Sha1Hash: sha1Hash,
			Field4: &pb.CommitUploadField1TypeField4Type{
				FileLastModifiedTimestamp: uploadTimestamp,
				Field2:                    unknownConstant,
			},
			Quality: quality,
			Field10: 1,
		},
		Field2: &pb.CommitUploadField2Type{
			Model:             model,
			Make:              a.Make,
			AndroidApiVersion: a.AndroidAPIVersion,
		},
		Field3: []byte{1, 3},
	}

	var response pb.CommitUploadResponse
	if err := a.DoProtoRequest(
		ctx,
		"https://photosdata-pa.googleapis.com/6439526531001121323/16538846908252377752",
		&requestBody,
		&response,
		WithAuth(),
		WithCommonHeaders(),
		WithStatusCheck(),
	); err != nil {
		return "", err
	}

	if response.GetField1() == nil || response.GetField1().GetField3() == nil {
		return "", fmt.Errorf("upload rejected by API: invalid response structure")
	}

	mediaKey := response.GetField1().GetField3().GetMediaKey()
	if mediaKey == "" {
		return "", fmt.Errorf("upload rejected by API: no media key returned")
	}

	return mediaKey, nil
}
