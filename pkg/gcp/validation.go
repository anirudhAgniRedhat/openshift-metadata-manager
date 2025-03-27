package gcp

import (
	"errors"
	"fmt"
	"regexp"
)

const (
	gcpKeyMaxLength   = 63
	gcpValueMaxLength = 63
	gcpMaxTags        = 64
	gcpKeyPattern     = "^[a-z0-9_-]{1,63}$"
)

// isValidGCPTag validates GCP tags
func IsValidGCPTag(tags map[string]string) error {
	if len(tags) > gcpMaxTags {
		return errors.New("GCP allows a maximum of 64 tags per resource")
	}
	re := regexp.MustCompile(gcpKeyPattern)
	for key, value := range tags {
		if len(key) == 0 || len(key) > gcpKeyMaxLength || !re.MatchString(key) {
			return fmt.Errorf("GCP tag key must match pattern %s", gcpKeyPattern)
		}
		if len(value) > gcpValueMaxLength {
			return fmt.Errorf("GCP tag value length must be 0-%d characters", gcpValueMaxLength)
		}
	}
	return nil
}
