package aws

import (
	"errors"
	"fmt"
)

const (
	awsKeyMaxLength   = 128
	awsValueMaxLength = 256
	awsMaxTags        = 50
)

// isValidAWSTag validates AWS tags
func IsValidAWSTag(tags map[string]string) error {
	if len(tags) > awsMaxTags {
		return errors.New("AWS allows a maximum of 50 tags per resource")
	}
	for key, value := range tags {
		if len(key) == 0 || len(key) > awsKeyMaxLength {
			return fmt.Errorf("AWS tag key length must be 1-%d characters", awsKeyMaxLength)
		}
		if len(value) > awsValueMaxLength {
			return fmt.Errorf("AWS tag value length must be 0-%d characters", awsValueMaxLength)
		}
	}
	return nil
}
