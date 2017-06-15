package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path"
)

// ErrNoAvatarURL is the error that is returned when the
// Avatar instance is unable to provide an avatar URL.
var ErrNoAvatarURL = errors.New("chat: Unable to get avatar URL.")

// Avatar represents types capable of representing
// user profile pictures.
type Avatar interface {

	// GetAvatarURL gets the avatar URL for the specified client,
	// or returns an error if something goes wrong.
	// ErrNoAvatarURL is returned if the object is unable to get
	// a URL for the specified client.
	GetAvatarURL(c *client) (string, error)
}

type AuthAvatar struct {
}

var UseAuthAvatar AuthAvatar

func (AuthAvatar) GetAvatarURL(c *client) (string, error) {

	url, ok := c.userData["avatar_url"]
	if !ok {
		return "", ErrNoAvatarURL
	}

	urlStr, ok := url.(string)
	if !ok {
		return "", ErrNoAvatarURL
	}

	return urlStr, nil

}

type GravatarAvatar struct {
}

var UseGravatar GravatarAvatar

func (GravatarAvatar) GetAvatarURL(c *client) (string, error) {
	userId, exists := c.userData["userId"]
	if !exists {
		return "", ErrNoAvatarURL
	}

	userIdStr, ok := userId.(string)
	if !ok {
		return "", ErrNoAvatarURL
	}

	return fmt.Sprintf("//www.gravatar.com/avatar/%s", userIdStr), nil
}

type FileSystemAvatar struct{}

var UserFileSystemAvatar FileSystemAvatar

func (FileSystemAvatar) GetAvatarURL(c *client) (string, error) {
	userId, exists := c.userData["userId"]

	if !exists {
		return "", ErrNoAvatarURL
	}

	useridStr, ok := userId.(string)
	if !ok {
		return "", ErrNoAvatarURL
	}

	files, err := ioutil.ReadDir("avatars")
	if err != nil {
		return "", ErrNoAvatarURL
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if match, _ := path.Match(useridStr+"*", file.Name()); match {
			return "/avatars/" + useridStr + ".jpg", nil
		}
	}

	return "", ErrNoAvatarURL

}
