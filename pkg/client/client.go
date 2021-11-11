package client

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type Image struct {
	Id   string
	Name string
}

func Init() (*client.Client, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}

	return cli, nil

}

func GetImages(cli *client.Client) ([]Image, error) {

	var images []Image

	imageslist, err := cli.ImageList(context.Background(), types.ImageListOptions{})

	if err != nil {
		return nil, err
	}

	for _, image := range imageslist {

		var img Image
		img.Id = image.ID

		name := strings.Join(image.RepoTags, " ")
		img.Name = name
		images = append(images, img)

	}

	return images, nil
}

func Save(id string) error {

	fmt.Println("💾 Saving Images as tar File")
	cli, err := client.NewEnvClient()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)

	}

	// convert string to string slice array in golang
	ids := make([]string, 1)
	ids[0] = id

	file_content, err := cli.ImageSave(context.Background(), ids)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	outFile, err := os.Create("output.tar")
	// handle err
	defer outFile.Close()
	_, err = io.Copy(outFile, file_content)
	if err != nil {
		return err

	}
	return nil
}
