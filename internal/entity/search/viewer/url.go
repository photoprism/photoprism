package viewer

import (
	"fmt"
)

// DownloadUrl returns a download url based on hash, api uri, and download token.
func DownloadUrl(fileHash, apiUri, downloadToken string) string {
	return fmt.Sprintf("%s/dl/%s?t=%s", apiUri, fileHash, downloadToken)
}
