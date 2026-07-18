package postgres

import "github.com/pood1e/realtime-me/services/library/internal/domain"

const uploadColumns = `uid, file_name, content_type, total_size_bytes, received_bytes,
	chunk_size_bytes, status, create_time, expire_time, COALESCE(claimed_resource_uid, ''),
	COALESCE(sealed_sha256, ''::bytea), COALESCE(sealed_size_bytes, 0),
	COALESCE(sealed_content_type, ''), COALESCE(sealed_storage_key, ''),
	COALESCE(sealed_content_uid, ''), failure_code`

func scanUpload(row rowScanner) (domain.Upload, error) {
	var upload domain.Upload
	var status string
	var sealed domain.SealedContent
	if err := row.Scan(&upload.UID, &upload.FileName, &upload.ContentType, &upload.TotalSizeBytes,
		&upload.ReceivedBytes, &upload.ChunkSizeBytes, &status, &upload.CreateTime, &upload.ExpireTime,
		&upload.ClaimedUID, &sealed.SHA256, &sealed.SizeBytes, &sealed.ContentType, &sealed.StorageKey,
		&upload.SealedContentUID, &upload.FailureCode); err != nil {
		return domain.Upload{}, err
	}
	upload.Status = domain.UploadStatus(status)
	if len(sealed.SHA256) > 0 {
		upload.Sealed = &sealed
	}
	return upload, nil
}
