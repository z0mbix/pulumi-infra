#!/usr/bin/env bash

[[ $TRACE ]] && set -x
set -euo pipefail

export DEBIAN_FRONTEND='noninteractive'

# Upgrade package cache and install some nice packages
apt-get -qq update
apt-get -y -q dist-upgrade
apt-get -y -q install \
  bat \
  ca-certificates \
  curl \
  curl \
  entr \
  fd-find \
  gnupg \
  gpg \
  jq \
  ncat \
  net-tools \
  ripgrep \
  tmux \
  unbound \
  unzip

# Install ansible
apt-add-repository -q ppa:ansible/ansible
apt-get -qq update
apt-get -y -qq install ansible
ansible --version
