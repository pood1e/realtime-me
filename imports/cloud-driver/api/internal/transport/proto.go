package transport

import (
	"fmt"
	"time"

	cloud_drivev1 "example.com/cloud-drive/api/gen/cloud/drive/v1"
	"example.com/cloud-drive/api/internal/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func itemProto(item domain.Item) *cloud_drivev1.DriveItem {
	result := &cloud_drivev1.DriveItem{
		Uid:         item.UID,
		Name:        item.Name,
		Kind:        itemKindProto(item.Kind),
		SizeBytes:   item.SizeBytes,
		ContentType: item.ContentType,
		CreateTime:  timestamppb.New(item.CreateTime),
		UpdateTime:  timestamppb.New(item.UpdateTime),
	}
	if item.ParentUID != nil {
		result.ParentUid = *item.ParentUID
	}
	if item.DeleteTime != nil {
		result.DeleteTime = timestamppb.New(*item.DeleteTime)
	}
	return result
}

func itemKindProto(kind domain.ItemKind) cloud_drivev1.DriveItemKind {
	if kind == domain.ItemKindDirectory {
		return cloud_drivev1.DriveItemKind_DRIVE_ITEM_KIND_DIRECTORY
	}
	return cloud_drivev1.DriveItemKind_DRIVE_ITEM_KIND_FILE
}

func uploadProto(upload domain.Upload) *cloud_drivev1.Upload {
	result := &cloud_drivev1.Upload{
		Uid:            upload.UID,
		ItemUid:        upload.ItemUID,
		FileName:       upload.FileName,
		ContentType:    upload.ContentType,
		TotalSizeBytes: upload.TotalSizeBytes,
		ReceivedBytes:  upload.ReceivedBytes,
		ChunkSizeBytes: upload.ChunkSizeBytes,
		Status:         uploadStatusProto(upload.Status),
		CreateTime:     timestamppb.New(upload.CreateTime),
		ExpireTime:     timestamppb.New(upload.ExpireTime),
	}
	if upload.ParentUID != nil {
		result.ParentUid = *upload.ParentUID
	}
	for _, chunk := range upload.Chunks {
		result.Chunks = append(result.Chunks, &cloud_drivev1.UploadChunk{
			StartOffset: chunk.StartOffset,
			EndOffset:   chunk.EndOffset,
		})
	}
	return result
}

func uploadStatusProto(status domain.UploadStatus) cloud_drivev1.UploadStatus {
	switch status {
	case domain.UploadStatusActive:
		return cloud_drivev1.UploadStatus_UPLOAD_STATUS_ACTIVE
	case domain.UploadStatusCompleted:
		return cloud_drivev1.UploadStatus_UPLOAD_STATUS_COMPLETED
	case domain.UploadStatusExpired:
		return cloud_drivev1.UploadStatus_UPLOAD_STATUS_EXPIRED
	default:
		return cloud_drivev1.UploadStatus_UPLOAD_STATUS_UNSPECIFIED
	}
}

func shareProto(share domain.ShareLink) *cloud_drivev1.ShareLink {
	result := &cloud_drivev1.ShareLink{
		Uid:        share.UID,
		TargetUid:  share.TargetUID,
		CreateTime: timestamppb.New(share.CreateTime),
		ExpireTime: timestamppb.New(share.ExpireTime),
	}
	if share.RevokeTime != nil {
		result.RevokeTime = timestamppb.New(*share.RevokeTime)
	}
	return result
}

func pageProto(page domain.Page) ([]*cloud_drivev1.DriveItem, string) {
	items := make([]*cloud_drivev1.DriveItem, 0, len(page.Items))
	for _, item := range page.Items {
		items = append(items, itemProto(item))
	}
	return items, page.NextPageToken
}

func timestampFromProto(value *timestamppb.Timestamp) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	if err := value.CheckValid(); err != nil {
		return nil, fmt.Errorf("invalid timestamp: %w", err)
	}
	result := value.AsTime().UTC()
	return &result, nil
}

func stringPointer(value string) *string {
	return &value
}
