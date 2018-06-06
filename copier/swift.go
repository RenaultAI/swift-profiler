package copier

import (
	"errors"
	"fmt"
	"strconv"

	gophercloud "github.com/gophercloud/gophercloud"
	openstack "github.com/gophercloud/gophercloud/openstack"
	objects "github.com/gophercloud/gophercloud/openstack/objectstorage/v1/objects"

	"github.robot.car/cruise/gofer/conf"
	"github.robot.car/cruise/gofer/file"
)

type swiftCopier struct {
	objectStorageClient gophercloud.ServiceClient
}

// NewSwiftCopier returns new Copier that copies into Swift.
func NewSwiftCopier() Copier {
	return &swiftCopier{}
}

// Setup implements Copier.Setup.
func (c *swiftCopier) Setup() error {
	spec, err := conf.Initialize()
	if err != nil {
		return fmt.Errorf("Conf initialization failed: %v", err)
	}

	if spec.SwiftUsername == "" || spec.SwiftPassword == "" || spec.SwiftAuthUrl == "" {
		return fmt.Errorf("Username, API key and authentication URL are required")
	}

	authOptions := gophercloud.AuthOptions{
		IdentityEndpoint: spec.SwiftAuthUrl,
		Username:         spec.SwiftUsername,
		Password:         spec.SwiftPassword,
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

	c.objectStorageClient = *objectStorageClient

	return nil
}

// DestinationPath implements Copier.DestinationPath.
func (c *swiftCopier) DestinationPath(checksums map[string]string, readOnly bool) string {
	return checksums["sha1"]
}

// QuarantinePath implements Copier.QuarantinePath.
// This is identical to DestinationPath since quarantine files
// will be placed in a different container.
func (c *swiftCopier) QuarantinePath(checksums map[string]string) string {
	return destinationPath("", checksums["sha1"], false)
}

// Write implements Copier.Write.
func (c *swiftCopier) Write(io file.IO, checksums map[string]string, readOnly bool) error {
	return c.write(io, false, checksums)
}

// WriteQuarantine implements Copier.WriteQuarantine.
func (c *swiftCopier) WriteQuarantine(io file.IO, checksums map[string]string) error {
	return c.write(io, true, checksums)
}

func containerName(checksum string, quarantine bool) string {
	if quarantine {
		return "quarantine"
	}
	return "bagstore-" + checksum[:4]
}

func objectName(copier Copier, quarantine bool, checksums map[string]string) string {
	if quarantine {
		return copier.QuarantinePath(checksums)
	} else {
		return copier.DestinationPath(checksums, false)
	}

}

// Internal copier function called by either Write or WriteQuarantine.
func (c *swiftCopier) write(io file.IO, quarantine bool, checksums map[string]string) error {
	// Open the file.
	fileObj, err := io.Open()
	if err != nil {
		return err
	}
	defer fileObj.Close()

	// Prepare for the upload.
	containerName := containerName(checksums["sha1"], quarantine)
	objectName := objectName(c, quarantine, checksums)
	options := objects.CreateOpts{ContentType: "application/octet-stream", Content: fileObj}

	// Execute the upload.
	createHeader, err := objects.Create(&c.objectStorageClient, containerName, objectName, options).Extract()
	if err != nil {
		return err
	} else if createHeader.ETag != checksums["md5"] {
		return errors.New("md5 received from Swift API does not match md5 of uploaded file.")
	}
	return nil
}

// Size implements Copier.Size.
func (c *swiftCopier) Size(checksums map[string]string, readOnly bool, quarantine bool) (int64, error) {
	containerName := containerName(checksums["sha1"], quarantine)
	objectName := objectName(c, quarantine, checksums)
	result := objects.Get(&c.objectStorageClient, containerName, objectName, nil)

	var contentLengthHeader struct {
		ContentLength string `json:"Content-Length"`
	}
	err := result.ExtractInto(&contentLengthHeader)

	contentLength, err := strconv.ParseInt(contentLengthHeader.ContentLength, 10, 64)
	return contentLength, err
}
