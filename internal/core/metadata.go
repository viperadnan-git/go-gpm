package core

import (
	"context"
	"time"

	"github.com/viperadnan-git/go-gpm/internal/pb"
)

// SetCaption sets the caption for a media item
// itemKey can be either mediaKey or dedupKey
func (a *Api) SetCaption(ctx context.Context, itemKey, caption string) error {
	requestBody := pb.SetCaption{
		Caption: caption,
		ItemKey: itemKey,
	}

	return a.DoProtoRequest(
		ctx,
		"https://photosdata-pa.googleapis.com/6439526531001121323/1552790390512470739",
		&requestBody,
		nil,
		WithAuth(),
		WithCommonHeaders(),
		WithStatusCheck(),
	)
}

// SetFavourite sets or removes the favourite status for a media item
// itemKey can be either mediaKey or dedupKey
// isFavourite: true = favourite, false = unfavourite
func (a *Api) SetFavourite(ctx context.Context, itemKey string, isFavourite bool) error {
	// Action map: true (favourite) = 1, false (unfavourite) = 2
	var action int64 = 2
	if isFavourite {
		action = 1
	}

	requestBody := pb.SetFavourite{
		Field1: &pb.SetFavourite_Field1{
			ItemKey: itemKey,
		},
		Field2: &pb.SetFavourite_Field2{
			Action: action,
		},
		Field3: &pb.SetFavourite_Field3{
			Field1: &pb.SetFavourite_Field3_Field1Inner{
				Field19: &pb.SetFavourite_Field3_Field1Inner_Field19{},
			},
		},
	}

	return a.DoProtoRequest(
		ctx,
		"https://photosdata-pa.googleapis.com/6439526531001121323/5144645502632292153",
		&requestBody,
		nil,
		WithAuth(),
		WithCommonHeaders(),
		WithStatusCheck(),
	)
}

// SetLocation sets the geographic location for a media item
// itemKey can be either mediaKey or dedupKey
// latitude and longitude are in decimal degrees
func (a *Api) SetLocation(ctx context.Context, itemKey string, latitude, longitude float32) error {
	// Convert coordinates to int32 by multiplying by 10^7
	// Google Photos stores coordinates as sfixed32 integers scaled by 10,000,000
	const scale = 10000000
	latInt := int32(latitude * scale)
	lonInt := int32(longitude * scale)

	// Calculate viewport bounds (roughly ±0.3 lat, ±0.125 lon from center)
	swLat := int32((latitude - 0.3) * scale)
	swLon := int32((longitude - 0.125) * scale)
	neLat := int32((latitude + 0.1) * scale)
	neLon := int32((longitude + 0.125) * scale)

	field2 := &pb.SetLocation_Field4_Field2{
		Action: 2, // Always 2 for setting location
		Coordinates: &pb.SetLocation_Field4_Field2_Coordinates{
			Latitude:  latInt,
			Longitude: lonInt,
		},
		Field3: &pb.SetLocation_Field4_Field2_Field3{
			Field1: &pb.SetLocation_Field4_Field2_Bounds{
				Field1: swLat,
				Field2: swLon,
			},
			Field2: &pb.SetLocation_Field4_Field2_Bounds{
				Field1: neLat,
				Field2: neLon,
			},
		},
	}

	// Use empty place name and generic place ID
	// Google Photos uses coordinates for map position
	field2.PlaceName = &pb.SetLocation_Field4_Field2_PlaceName{
		Name:   "", // Empty - let Google Photos handle it
		Field3: 1,
	}
	field2.PlaceId = "ChIJN1t_tDeuEmsRUsoyG83frY4" // Generic Place ID (required by API)

	requestBody := pb.SetLocation{
		Field4: &pb.SetLocation_Field4{
			Field1: &pb.SetLocation_Field4_Field1{
				MediaKey: itemKey,
			},
			Field2: field2,
		},
	}

	return a.DoProtoRequest(
		ctx,
		"https://photosdata-pa.googleapis.com/6439526531001121323/227609453150053792",
		&requestBody,
		nil,
		WithAuth(),
		WithCommonHeaders(),
		WithStatusCheck(),
	)
}

// SetDateTime sets the date and time for one or more media items
// itemKeys can be mediaKey or dedupKey (one or more)
// timestamp is a time.Time value
func (a *Api) SetDateTime(ctx context.Context, itemKeys []string, timestamp time.Time) error {
	// Convert to Unix seconds (as float64)
	timestampSec := float64(timestamp.Unix())

	// Get timezone offset in seconds
	_, offset := timestamp.Zone()

	requestBody := pb.SetDateTime{
		Field1: &pb.SetDateTime_Field1{
			MediaKey:       itemKeys,
			Timestamp:      timestampSec,
			TimezoneOffset: int32(offset),
		},
	}

	return a.DoProtoRequest(
		ctx,
		"https://photosdata-pa.googleapis.com/6439526531001121323/17462398412150687934",
		&requestBody,
		nil,
		WithAuth(),
		WithCommonHeaders(),
		WithStatusCheck(),
	)
}
