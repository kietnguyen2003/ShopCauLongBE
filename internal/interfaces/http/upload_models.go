package http

import appUpload "kafka-order-demo/backend/internal/application/upload"

type uploadResponse struct {
	Bucket      string `json:"bucket"`
	ObjectKey   string `json:"object_key"`
	URL         string `json:"url"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
}

func toUploadHTTPResponse(resp *appUpload.UploadResponse) uploadResponse {
	return uploadResponse{
		Bucket:      resp.Bucket,
		ObjectKey:   resp.ObjectKey,
		URL:         resp.URL,
		ContentType: resp.ContentType,
		Size:        resp.Size,
	}
}
