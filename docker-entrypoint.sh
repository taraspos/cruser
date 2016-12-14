#!/bin/bash
echo ">>>>>>>>>>>>>> running cruser tool"
cruser
sleep 1
echo ""
echo ">>>>>>>>>>>>>> checking content of '/etc/passwd'"
cat /etc/passwd
sleep 1
echo ""
echo ">>>>>>>>>>>>>> checking if user's homedirectories exists'"
ls -al /home/
sleep 1
echo ""
echo ">>>>>>>>>>>>>> checking content of authorized_keys files"
for f in /home/*/.ssh/authorized_keys; do echo "Filename: $f"; cat "$f" && echo ""; done
sleep 1
echo ""
echo ">>>>>>>>>>>>>> checking permissions of '.ssh' directories and 'authorized_keys' files"
ls -al /home/*/ | grep ssh
sleep 1
ls -al /home/*/.ssh/authorized_keys
sleep 1
echo ""
echo ">>>>>>>>>>>>>> checking content of '/etc/sudoers.d' directory and files permissions"
ls -al /etc/sudoers.d/
sleep 1
echo ""
echo ">>>>>>>>>>>>>> ehcking content of sudoers files "
for f in /etc/sudoers.d/*; do echo "Filename: $f"; cat "$f" && echo ""; done
sleep 1
echo ""
echo ">>>>>>>>>>>>>> trying if sudo not broken"
sudo echo "!!!!!!!!!!!SUDO WORKS!!!!!!!!!!!"
sleep 1
echo ""