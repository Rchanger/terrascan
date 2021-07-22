package utils

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"go.uber.org/zap"
)

const (
	tagDelim         = ":"
	regRepoDelimiter = "/"
	defaultTagValue  = "latest"
)

// ScanImageWithAWSECRAPI calles aws ecr api to get image scan details
func ScanImageWithAWSECRAPI(ctx context.Context, imageName string) (*ecr.DescribeImageScanFindingsOutput, error) {
	mySession := session.Must(session.NewSession())

	// Create a ECR client from just a session.
	client := ecr.New(mySession)

	tag, repositoryName := GetImageTagAndRepository(imageName)

	if tag == "" || repositoryName == "" {
		zap.S().Error("invalid image refernce %s", imageName)
		return nil, fmt.Errorf("invalid image refernce %s", imageName)
	}

	return GetImageScanResultFromECR(ctx, client, imageName, tag, repositoryName)

}

// StartECRImageScan starts the scan of provided image
func StartECRImageScan(ctx context.Context, client *ecr.ECR, imageName, tag, repositoryName string) error {

	input := ecr.StartImageScanInput{
		ImageId: &ecr.ImageIdentifier{
			ImageTag: &tag,
		},
		RepositoryName: &repositoryName,
	}

	_, err := client.StartImageScanWithContext(ctx, &input)
	if err != nil {
		zap.S().Error("error scaninng image %s: %v", imageName, err)
		return err
	}

	describeImageScanFindingsInput := ecr.DescribeImageScanFindingsInput{
		ImageId: &ecr.ImageIdentifier{
			ImageTag: &tag,
		},
		RepositoryName: &repositoryName,
	}

	// wait until scan of image is complete
	err = client.WaitUntilImageScanComplete(&describeImageScanFindingsInput)
	if err != nil {
		zap.S().Error("error scaninng image %s: %v", imageName, err)
		return err
	}
	return nil
}

// GetImageScanResultFromECR get the scan result from ECR
func GetImageScanResultFromECR(ctx context.Context, client *ecr.ECR, imageName, tag, repositoryName string) (*ecr.DescribeImageScanFindingsOutput, error) {
	describeImageScanFindingsInput := ecr.DescribeImageScanFindingsInput{
		ImageId: &ecr.ImageIdentifier{
			ImageTag: &tag,
		},
		RepositoryName: &repositoryName,
	}
	results, err := client.DescribeImageScanFindingsWithContext(ctx, &describeImageScanFindingsInput)
	if err != nil {
		if _, ok := err.(*ecr.ScanNotFoundException); ok {
			err := StartECRImageScan(ctx, client, imageName, tag, repositoryName)
			if err != nil {
				return results, err
			}
			return GetImageScanResultFromECR(ctx, client, imageName, tag, repositoryName)
		}
		zap.S().Error("error scaninng image %s: %v", imageName, err)
		return results, err
	}
	if results != nil && results.ImageScanStatus != nil && results.ImageScanStatus.Status != nil && results.ImageScanFindings != nil && results.ImageScanFindings.ImageScanCompletedAt != nil {
		if strings.EqualFold(*results.ImageScanStatus.Status, ecr.ScanStatusComplete) && (*results.ImageScanFindings.ImageScanCompletedAt).Before(time.Now().Add(24*time.Hour)) {
			return results, nil
		}
		if strings.EqualFold(*results.ImageScanStatus.Status, ecr.ScanStatusFailed) {
			return results, nil
		}
		if strings.EqualFold(*results.ImageScanStatus.Status, ecr.ScanStatusInProgress) {
			return GetImageScanResultFromECR(ctx, client, imageName, tag, repositoryName)
		}
		if strings.EqualFold(*results.ImageScanStatus.Status, ecr.ScanStatusComplete) && !((*results.ImageScanFindings.ImageScanCompletedAt).Before(time.Now().Add(24 * time.Hour))) {
			err := StartECRImageScan(ctx, client, imageName, tag, repositoryName)
			if err != nil {
				return results, err
			}
			return GetImageScanResultFromECR(ctx, client, imageName, tag, repositoryName)
		}

	}
	return results, nil
}

func GetImageTagAndRepository(imageName string) (tag string, repository string) {
	parts := strings.Split(imageName, tagDelim)
	base := imageName
	tag = defaultTagValue
	if len(parts) > 1 && !strings.Contains(parts[len(parts)-1], regRepoDelimiter) {
		base = strings.Join(parts[:len(parts)-1], tagDelim)
		tag = parts[len(parts)-1]
	}

	repository = GetRepository(base)
	return
}

func GetRepository(imageName string) (repository string) {
	parts := strings.SplitN(imageName, regRepoDelimiter, 2)
	if len(parts) == 2 && (strings.ContainsRune(parts[0], '.') || strings.ContainsRune(parts[0], ':')) {
		repository = parts[1]
	}
	return
}
