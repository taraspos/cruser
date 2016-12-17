package main

import (
	"bufio"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"regexp"

	"fmt"

	"github.com/trane9991/cruser/user"
)

type sshKey struct {
	Line     string
	Type     string
	Key      string
	Username string
	Email    string
}

type sshKeys []sshKey

var re = regexp.MustCompile(`(\S*) (\S*) ((\w*)@.*)`)

type usr struct {
	Profile user.Profile
	// TODO: exists is not used anywhere. Do something with it...
	Exist bool
}
type users []usr

func main() {
	fileWithKeys := flag.String("file", "users", "File with the list of SSH-keys and user emails in format of '~/.ssh/authorized_keys' file")
	dryRun := flag.Bool("dry-run", false, "Do not execute commands, just print them.")
	flag.Parse()
	// Read provided ssh keys
	lines := readLines(*fileWithKeys)
	// Remove dublicated SSH keys
	lines = removeDuplicatesUnordered(lines)
	// Get username from ssh keys emails
	sks := parseKeys(lines)

	// Transform list of SSH keys to the list of users
	us := keys2users(sks, false)
	fmt.Printf("%+v", us)

	// Checking if users exist, then read SSH keys that he already have\
	// And add them to the list
	us.checkUsers()
	// Merge users again with SSH keys from existing users
	us = us.mergeUsers()
	user.DryRun = *dryRun
	if user.DryRun {
		log.SetOutput(ioutil.Discard)
	}
	for _, u := range us {
		if !u.Profile.Exists() {
			u.Profile.Shell = "/bin/bash"
			if err := u.Profile.Create(); err != nil {
				log.Printf("Failed creating user '%s': %v", u.Profile.Name, err)
			} else {
				log.Printf("User '%s' Successfully created", u.Profile.Name)

				if err := u.Profile.AuthorizeSSHKeys(); err != nil {
					log.Printf("Failed to add ssh key for user '%s': %v", u.Profile.Name, err)
				}
				if err := u.Profile.AuthorizeSudo(); err != nil {
					log.Printf("Failed to authorize sudo rights for user '%s': %v", u.Profile.Name, err)
				}
			}
		} else {
			log.Printf("User '%s' already exists. Appending SSH keys forcefully(rewriting completely SSH-keys file with old and new ssh keys)...", u.Profile.Name)
			if err := u.Profile.AuthorizeSSHKeys(); err != nil {
				log.Printf("Failed to add ssh key for user '%s': %v", u.Profile.Name, err)
			}
		}
	}
}

// parse ssh keys, and save next data from it. Example:
// ssh-rsa  aaaaaa...bbbbbbbb test@user.com
// 1. ssh-key       - ssh-rsa  aaaaaa...bbbbbbbb test@user.com
// 2. Username      - test
// 3. Email address - test@user.com
func parseKeys(lines []string) sshKeys {
	var sks sshKeys
	for _, key := range lines {
		var sk sshKey
		match := re.FindStringSubmatch(key)

		if len(match) == 5 {
			sk.Line = match[0]
			sk.Type = match[1]
			sk.Key = match[2]
			sk.Email = match[3]
			sk.Username = match[4]

			if sk.Type != "ssh-rsa" && sk.Type != "ssh-dss" && sk.Type != "ssh-ed25519" && sk.Type != "ecdsa-sha2-nistp256" {
				fmt.Printf("'%s' - not known SSH Key type ", sk.Type)
			} else {
				sks = append(sks, sk)
			}
		}
	}
	return sks
}

func keys2users(sks sshKeys, existing bool) users {
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
					var u usr
					u.Profile.Name = sk.Username
					u.Profile.SSHAuthorizedKeys = append(u.Profile.SSHAuthorizedKeys, sk.Line)
					u.Profile.Comment = sk.Email
					result = append(result, u)
				}
			}
		case value > 1: // if user appears more then once on the list, merge them in one and add to the list of users
			var u usr
			for _, sk := range sks {
				if key == sk.Username {
					u.Profile.Name = sk.Username
					u.Profile.SSHAuthorizedKeys = append(u.Profile.SSHAuthorizedKeys, sk.Line)
					u.Profile.Comment = fmt.Sprintf("%s%s;", u.Profile.Comment, sk.Email)

					// if reading keys of existing users setting Exist=True
					u.Exist = existing
				}
			}
			result = append(result, u)
		}
	}
	return result
}

// TODO: this func and func above pretty similar. Do something
// merge dublicated users in one with multiple ssh-keys
func (us users) mergeUsers() users {
	encountered := map[string]int{}
	// Count how many times user appears in the file
	for i := range us {
		encountered[us[i].Profile.Name]++
	}

	var result users

	for key, value := range encountered {
		switch {
		case value == 1: // if user appears only once in list, add him to the new list of users
			for _, u := range us {
				if key == u.Profile.Name {
					result = append(result, u)
				}
			}
		case value > 1: // if user appears more then once on the list, merge them in one and add to the list of users
			var mergedUser usr
			for _, u := range us {
				if key == u.Profile.Name {
					mergedUser.Profile.Name = key
					mergedUser.Profile.SSHAuthorizedKeys = append(mergedUser.Profile.SSHAuthorizedKeys, u.Profile.SSHAuthorizedKeys...)
					mergedUser.Profile.Comment = fmt.Sprintf("%s%s;", mergedUser.Profile.Comment, u.Profile.Comment)
					if u.Exist || mergedUser.Exist {
						mergedUser.Exist = true
					}
				}
			}
			result = append(result, mergedUser)
		}
	}

	return result
}

// read lines from file
func readLines(path string) []string {
	var lines []string
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return lines
}

// https://www.dotnetperls.com/duplicates-go
func removeDuplicatesUnordered(elements []string) []string {
	encountered := map[string]bool{}

	// Create a map of all unique elements.
	for v := range elements {
		encountered[elements[v]] = true
	}

	// Place all keys from the map into a slice.
	result := []string{}
	for key := range encountered {
		result = append(result, key)
	}
	return result
}

func getSSHKeysFromExistingUser(u usr) users {
	// Read provided ssh keys
	lines := readLines(u.Profile.AuthorizedKeysFile())
	// Remove dublicated SSH keys
	lines = removeDuplicatesUnordered(lines)
	// Get username from ssh keys emails
	sks := parseKeys(lines)
	us := keys2users(sks, true)
	return us
}

func (us *users) checkUsers() {
	for _, u := range *us {
		u.Exist = u.Profile.Exists()
		if u.Exist {
			*us = append(*us, getSSHKeysFromExistingUser(u)...)
		}
	}
}
