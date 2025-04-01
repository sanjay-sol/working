package models

import "github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"

type AzureClient struct {
	Client        *azblob.Client
	ContainerName string
	Credential    *azblob.SharedKeyCredential
}
