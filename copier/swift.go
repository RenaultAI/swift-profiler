package copier

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	gophercloud "github.com/gophercloud/gophercloud"
	openstack "github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/objectstorage/v1/containers"
	objects "github.com/gophercloud/gophercloud/openstack/objectstorage/v1/objects"
)

const defaultContainer = "benchmark-test"

type swiftCopier struct {
	objectStorageClient gophercloud.ServiceClient
}

// NewSwiftCopier returns new a Swift client.
func NewSwiftCopier() *swiftCopier {
	return &swiftCopier{}
}

// Setup implements Copier.Setup.
func (c *swiftCopier) Setup() error {
	authOptions, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		return err
	}

	// Pass authentication options to get a ProviderClient.
	provider, err := openstack.AuthenticatedClient(authOptions)
	if err != nil {
		return fmt.Errorf("Swift credentials did not authenticate: %v", err)
	}

	// Get endpoint options.
	endpointOptions := gophercloud.EndpointOpts{Type: "object-store"}

	// Create a Swift client.
	objectStorageClient, err := openstack.NewObjectStorageV1(provider, endpointOptions)
	if err != nil {
		return fmt.Errorf("Could not create a Swift objectStorageClient: %v", err)
	}

	// Pre-create the container if it doesn't exist.
	metadata, err := containers.Get(objectStorageClient, defaultContainer).ExtractMetadata()
	if _, ok := err.(gophercloud.ErrDefault404); ok {
		log.Printf("Creating swift container %s\n", defaultContainer)
		if _, err = containers.Create(objectStorageClient, defaultContainer, nil).Extract(); err != nil {
			return err
		}
	}

	metadata, err = containers.Get(objectStorageClient, defaultContainer).ExtractMetadata()
	if err != nil {
		return err
	}
	log.Printf("Swift container metadata: %+v\n", metadata)

	c.objectStorageClient = *objectStorageClient

	return nil
}

// Write implements Copier.Write.
func (c *swiftCopier) Copy(sourcePath, destinationContainer string, checksum *string) error {
	// Open the file.
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Prepare for the upload.
	objectName := filepath.Base(sourcePath)
	var options objects.CreateOpts
	if checksum == nil {
		options = objects.CreateOpts{ContentType: "application/octet-stream", Content: sourceFile}
	} else {
		options = objects.CreateOpts{ContentType: "application/octet-stream", Content: sourceFile, ETag: *checksum}
	}

	// Execute the upload.
	_, err = objects.Create(&c.objectStorageClient, destinationContainer, objectName, options).Extract()
	return err

	// TODO: perform checksum check later.
	// if err != nil {
	// return err
	// } else if createHeader.ETag != checksums["md5"] {
	// return errors.New("md5 received from Swift API does not match md5 of uploaded file.")
	// }
	// return nil
}

// Size implements Copier.Size.
// func (c *swiftCopier) Size(checksums map[string]string, readOnly bool, quarantine bool) (int64, error) {
// containerName := containerName(checksums["sha1"], quarantine)
// objectName := objectName(c, quarantine, checksums)
// result := objects.Get(&c.objectStorageClient, containerName, objectName, nil)

// var contentLengthHeader struct {
// ContentLength string `json:"Content-Length"`
// }
// err := result.ExtractInto(&contentLengthHeader)

// contentLength, err := strconv.ParseInt(contentLengthHeader.ContentLength, 10, 64)
// return contentLength, err
// }
