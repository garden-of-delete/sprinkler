package common

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/spf13/viper"
)

type AwsCredentials struct {
	ClientRegion   string
	AwsAccessKeyId string
	AwsSecretKey   string
	SessionToken   string
	AssumeRoleArn  string
}

func WithAwsCredentials() AwsCredentials {
	return AwsCredentials{
		ClientRegion:   viper.GetString("aws.clientRegion"),
		AwsAccessKeyId: viper.GetString("aws.staticCredentials.awsAccessKeyId"),
		AwsSecretKey:   viper.GetString("aws.staticCredentials.awsSecretKey"),
		SessionToken:   viper.GetString("aws.staticCredentials.sessionToken"),
		AssumeRoleArn:  viper.GetString("aws.assumeRoleArn"),
	}
}

func (c AwsCredentials) credentialsProvider() (config.LoadOptionsFunc, error) {
	var credProvider = config.WithCredentialsProvider(nil)
	if c.AwsAccessKeyId != "" && c.AwsSecretKey != "" {
		credProvider = config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				c.AwsAccessKeyId,
				c.AwsSecretKey,
				c.SessionToken, // empty string will be ignored
			),
		)
	}
	if c.AssumeRoleArn != "" {
		awsConfig, err := config.LoadDefaultConfig(
			context.TODO(),
			credProvider, // use static credentials if provided
		)
		if err != nil {
			return nil, err
		}
		stsClient := sts.NewFromConfig(awsConfig)
		assumeRoleProvider := config.WithCredentialsProvider(
			stscreds.NewAssumeRoleProvider(
				stsClient,
				c.AssumeRoleArn,
			),
		)
		return assumeRoleProvider, nil
	}
	return credProvider, nil // default credentials chain if no aws configs set
}

func (c AwsCredentials) AwsConfig() (aws.Config, error) {
	credProvider, err := c.credentialsProvider()
	if err != nil {
		return aws.Config{}, err
	}
	return config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(c.ClientRegion), // empty string will be ignored
		credProvider,
	)
}

func (c AwsCredentials) S3Client() (*s3.Client, error) {
	awsConfig, err := c.AwsConfig()
	if err != nil {
		return nil, fmt.Errorf("Couldn't load configuration. Error: %w\n", err)
	}
	s3Client := s3.NewFromConfig(awsConfig)
	return s3Client, nil
}
