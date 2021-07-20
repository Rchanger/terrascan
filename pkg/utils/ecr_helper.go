package utils

import (
	"context"
	"fmt"
	"strings"

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

	// start scan of the image
	// input := ecr.StartImageScanInput{
	// 	ImageId: &ecr.ImageIdentifier{
	// 		ImageTag: &tag,
	// 	},
	// 	RepositoryName: &repositoryName,
	// }

	// req, _ := client.StartImageScanRequest(&input)

	// err := req.Send()
	// if err != nil {
	// 	zap.S().Error("error scaninng image %s: %v", imageName, err)
	// 	return err
	// }

	describeImageScanFindingsInput := ecr.DescribeImageScanFindingsInput{
		ImageId: &ecr.ImageIdentifier{
			ImageTag: &tag,
		},
		RepositoryName: &repositoryName,
	}

	// wait until scan of image is complete
	// err = client.WaitUntilImageScanComplete(&describeImageScanFindingsInput)
	// if err != nil {
	// 	zap.S().Error("error scaninng image %s: %v", imageName, err)
	// 	return err
	// }

	results, err := client.DescribeImageScanFindingsWithContext(ctx, &describeImageScanFindingsInput)
	if err != nil {
		zap.S().Error("error scaninng image %s: %v", imageName, err)
		return results, err
	}

	fmt.Println("findings<><>><><>", results)

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
