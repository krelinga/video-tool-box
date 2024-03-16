#! /bin/bash

# Exit on any error
set -e

# Instructions based on the 'Debian packages' section of https://mkvtoolnix.download/downloads.html
wget -O /usr/share/keyrings/gpg-pub-moritzbunkus.gpg https://mkvtoolnix.download/gpg-pub-moritzbunkus.gpg

echo 'deb [signed-by=/usr/share/keyrings/gpg-pub-moritzbunkus.gpg] https://mkvtoolnix.download/debian/ bookworm main' > /etc/apt/sources.list.d/mkvtoolnix.download.list
echo 'deb-src [signed-by=/usr/share/keyrings/gpg-pub-moritzbunkus.gpg] https://mkvtoolnix.download/debian/ bookworm main' >> /etc/apt/sources.list.d/mkvtoolnix.download.list

apt-get update -y
apt-get install -y mkvtoolnix
