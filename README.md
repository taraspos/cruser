# [CR]eate [USER] -> cruser

Tool to create users on Linux system. 

Mechanism of users creation are inspired by Google Cloud, when username are automatically taken from SSH-key(and email after).

List of users are need to be provided in the **authorized_keys** file format.

## Requirements 
Tool don't use any external libraries, but relies on some default(in most Linux distros) CLI tools:
* `id` - used to check, if user exist in the system, get user's *gid* and *uid*
* `useradd` - used for user creation
* `visudo` - used for validation of sudoers lines

## Example:
* File *users*:
```
ssh-rsa aaaaaaaaaaaaaaaaaaaaaa test@gmail.com
ssh-rsa bbbbbbbbbbbbbbbbbbbbbb test@gmail.com
ssh-rsa cccccccccccccccccccccc hello@gmail.com
ssh-rsa cccccccccccccccccccccc hello@gmail.com
```

* Result of running command `cruser -file users` will be:
    * Created users **test** and **hello**
    * Sudoers lines are generated and validated with **visudo**
    * Line `test ALL=(ALL) NOPASSWD:ALL` added to the file */etc/sudoers.d/test*
    * Line `hello ALL=(ALL) NOPASSWD:ALL` added to the file */etc/sudoers.d/hello*
    * Provided SSH keys are added to the */home/test/.ssh/authorized_keys* and */home/hello/.ssh/authorized_keys*
    * Duplicated lines are skipped
    * **test@gmail.com** and **hello@gmail.com** are added as comment entry in the */etc/password* file

## Build:
```
make build
```

## Running demo:
```
make demo
```

## Missing features:
* Adding SSH-keys for existing users. Currently only newly created users supported
* Reading SSH-keys list form remote location(S3, github, etc)
* More flexible Sudoers configuration(only NOPASSWD:ALL are supported now)