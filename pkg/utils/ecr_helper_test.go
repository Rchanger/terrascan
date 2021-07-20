package utils

import "testing"

func TestGetImageTagAndRepository(t *testing.T) {
	type args struct {
		imageName string
	}
	tests := []struct {
		name           string
		args           args
		wantTag        string
		wantRepository string
	}{
		{
			name:           "repository with specific tag",
			args:           args{imageName: "245578940568.dkr.ecr.us-east-2.amazonaws.com/image-scan-poc-suvarna:terrascan"},
			wantTag:        "terrascan",
			wantRepository: "image-scan-poc-suvarna",
		},
		{
			name:           "repository without specific tag",
			args:           args{imageName: "245578940568.dkr.ecr.us-east-2.amazonaws.com/image-scan-poc-suvarna"},
			wantTag:        "latest",
			wantRepository: "image-scan-poc-suvarna",
		},
		{
			name:           "inavlid image repository",
			args:           args{imageName: "245578940568.dkr.ecr.us-east-2.amazonaws.com"},
			wantTag:        "latest",
			wantRepository: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTag, gotRepository := GetImageTagAndRepository(tt.args.imageName)
			if gotTag != tt.wantTag {
				t.Errorf("GetImageTagAndRepository() gotTag = %v, want %v", gotTag, tt.wantTag)
			}
			if gotRepository != tt.wantRepository {
				t.Errorf("GetImageTagAndRepository() gotRepository = %v, want %v", gotRepository, tt.wantRepository)
			}
		})
	}
}
