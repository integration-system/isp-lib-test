package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/integration-system/isp-lib-test/ctx"
	"github.com/integration-system/isp-lib-test/internal"
	log "github.com/integration-system/isp-log"
)

const backupPrefixFilename = "isp-test-docker-session_"

type (
	imageId     string
	containerId string

	backup struct {
		BasicContainers map[containerId]imageId
		AppContainers   map[containerId]imageId
		NetworkId       string
	}
)

func (te *TestEnvironment) makeBackupFile() {
	te.updateBackup()

	data, err := json.Marshal(*te.backup)
	if err != nil {
		log.Errorf(0, "can't marshal docker settings")
		return
	}

	file, err := os.Create(getFileName())
	if err != nil || file == nil {
		log.Errorf(0, "can't create file %s to save docker settings: %v", getFileName(), err)
		return
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Errorf(0, "can't close file %s with docker settings: %v", getFileName(), err)
		}
	}()

	_, err = file.Write(data)
	if err != nil {
		log.Errorf(0, "can't write docker settings into file %s", getFileName())
		return
	}
}

func (te *TestEnvironment) updateBackup() {
	for _, container := range te.basicContainers {
		if _, ok := te.backup.BasicContainers[containerId(container.containerId)]; !ok {
			te.backup.BasicContainers[containerId(container.containerId)] = imageId(container.imageId)
		}
	}
	for _, container := range te.appContainers {
		if _, ok := te.backup.AppContainers[containerId(container.containerId)]; !ok {
			te.backup.AppContainers[containerId(container.containerId)] = imageId(container.imageId)
		}
	}
	if te.network != nil {
		te.backup.NetworkId = te.network.id
	}
}

func init() {
	internal.CleanupByBackup = CleanupByBackup
}

func CleanupByBackup() error {
	cli, err := NewClient()
	if err != nil {
		return fmt.Errorf("can't open new docker client: %v", err)
	}

	var errors *multierror.Error
	backupFiles, err := getBackupFileNames()
	if err != nil {
		return err
	}
	sort.Strings(backupFiles)
	for _, fileName := range backupFiles {
		b, err := readBackupFile(fileName)
		errors = multierror.Append(errors, err)
		err = b.cleanup(cli)
		errors = multierror.Append(errors, err)
		err = os.Remove(fileName)
		errors = multierror.Append(errors, err)
	}
	return errors.ErrorOrNil()
}

func readBackupFile(filename string) (*backup, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("can't read file %s with docker settings", getFileName())
	}
	b := &backup{}

	err = json.Unmarshal(data, b)
	if err != nil {
		return nil, fmt.Errorf("can't unmarshal docker settings from %s", getFileName())
	}
	return b, nil
}

func (b backup) cleanup(cli *ispDockerClient) error {
	var errors *multierror.Error
	var err error
	for containerId, imageId := range b.AppContainers {
		err := (&ContainerContext{
			imageId:     string(imageId),
			containerId: string(containerId),
			client:      cli,
		}).Close()
		errors = multierror.Append(errors, err)
	}

	for containerId, _ := range b.BasicContainers {
		err := (&ContainerContext{
			containerId: string(containerId),
			client:      cli,
		}).ForceRemoveContainer()
		errors = multierror.Append(errors, err)
	}
	if b.NetworkId != "" {
		err = cli.c.NetworkRemove(context.Background(), b.NetworkId)
	}
	errors = multierror.Append(errors, err)

	err = cli.Close()
	errors = multierror.Append(errors, err)
	return errors.ErrorOrNil()
}

func getBackupFileNames() ([]string, error) {
	files, err := ioutil.ReadDir(".")
	if err != nil {
		return nil, fmt.Errorf("can't read dirr")
	}
	backupFiles := make([]string, 0)
	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), backupPrefixFilename) {
			backupFiles = append(backupFiles, file.Name())
		}
	}
	return backupFiles, nil
}

func getFileName() string {
	return backupPrefixFilename + ctx.CurrentSessionName()
}
