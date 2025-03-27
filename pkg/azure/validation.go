package azure

import (
	"errors"
	"fmt"
)

const (
	azureKeyMaxLength   = 512
	azureValueMaxLength = 256
	azureMaxTags        = 50
)

// isValidAzureTag validates Azure tags
func IsValidAzureTag(tags map[string]string) error {
	if len(tags) > azureMaxTags {
		return errors.New("Azure allows a maximum of 50 tags per resource")
	}
	for key, value := range tags {
		if len(key) == 0 || len(key) > azureKeyMaxLength {
			return fmt.Errorf("Azure tag key length must be 1-%d characters", azureKeyMaxLength)
		}
		if len(value) > azureValueMaxLength {
			return fmt.Errorf("Azure tag value length must be 0-%d characters", azureValueMaxLength)
		}
	}
	return nil
}
