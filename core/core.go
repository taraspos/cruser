package core

import (
	"io/ioutil"
	"log"

	"fmt"

	"github.com/trane9991/cruser/core/keys"
	"github.com/trane9991/cruser/core/user"
	"github.com/trane9991/cruser/core/utils"
)

type users []user.Profile

// Run - creates users in the system from provided list of SSH keys
func Run(path *string, dryRun *bool) {

	us := getUsersFromFileWithKeys(*path)

	// Checking if users exist, then read existing SSH keys
	// from this user and adds them to the list
	us.checkUsers()

	// Merge users again with SSH keys from existing users
	us = us.mergeUsers()
	user.DryRun = *dryRun
	if user.DryRun {
		log.SetOutput(ioutil.Discard)
	}
	for _, u := range us {
		if !u.Exists() {
			u.Shell = "/bin/bash"
			if err := u.Create(); err != nil {
				log.Printf("Failed creating user '%s': %v", u.Name, err)
			} else {
				log.Printf("User '%s' Successfully created", u.Name)

				if err := u.AuthorizeSSHKeys(); err != nil {
					log.Printf("Failed to add ssh key for user '%s': %v", u.Name, err)
				}
				if err := u.AuthorizeSudo(); err != nil {
					log.Printf("Failed to authorize sudo rights for user '%s': %v", u.Name, err)
				}
			}
		} else {
			log.Printf("User '%s' already exists. Appending SSH keys forcefully (rewriting completely SSH-keys file with old and new ssh keys)...", u.Name)
			if err := u.AuthorizeSSHKeys(); err != nil {
				log.Printf("Failed to add ssh key for user '%s': %v", u.Name, err)
			}
		}
	}
}

// convert list of keys to the list of SSH users
func keys2users(sks keys.SSHKeys, existing bool) users {
	encountered := map[string]int{}
	// Count how many times user appears in the file
	for i := range sks {
		encountered[sks[i].Username]++
	}
	var result users

	for key, value := range encountered {
		switch {
		case value == 1: // if user appears only once in list, add him to the new list of users
			for _, sk := range sks {
				if key == sk.Username {
					var u user.Profile
					u.Name = sk.Username
					u.SSHAuthorizedKeys = append(u.SSHAuthorizedKeys, sk.Line)
					u.Comment = sk.Email
					result = append(result, u)
				}
			}
		case value > 1: // if user appears more then once on the list, merge them in one and add to the list of users
			var u user.Profile
			for _, sk := range sks {
				if key == sk.Username {
					u.Name = sk.Username
					u.SSHAuthorizedKeys = append(u.SSHAuthorizedKeys, sk.Line)
					u.Comment = fmt.Sprintf("%s%s;", u.Comment, sk.Email)
				}
			}
			result = append(result, u)
		}
	}
	return result
}

// merge dublicated users in one with multiple ssh-keys
func (us users) mergeUsers() users {
	encountered := map[string]int{}
	// Count how many times user appears in the file
	for i := range us {
		encountered[us[i].Name]++
	}

	var result users

	for key, value := range encountered {
		switch {
		case value == 1: // if user appears only once in list, add him to the new list of users
			for _, u := range us {
				if key == u.Name {
					result = append(result, u)
				}
			}
		case value > 1: // if user appears more then once on the list, merge them in one and add to the list of users
			var mergedUser user.Profile
			for _, u := range us {
				if key == u.Name {
					mergedUser.Name = key
					mergedUser.SSHAuthorizedKeys = append(mergedUser.SSHAuthorizedKeys, u.SSHAuthorizedKeys...)
					mergedUser.Comment = fmt.Sprintf("%s%s;", mergedUser.Comment, u.Comment)
				}
			}
			result = append(result, mergedUser)
		}
	}

	return result
}

// read and parse ssh keys of existing user
func getUsersFromFileWithKeys(path string) users {
	// Read provided ssh keys
	lines := utils.ReadLines(path)
	// Remove dublicated SSH keys
	lines = utils.RemoveDuplicatesUnordered(lines)
	// Get username from ssh keys emails
	sks := keys.Parse(lines)
	us := keys2users(sks, true)
	return us
}

func (us *users) checkUsers() {
	for _, u := range *us {
		if u.Exists() {
			*us = append(*us, getUsersFromFileWithKeys(u.AuthorizedKeysFile())...)
		}
	}
}
