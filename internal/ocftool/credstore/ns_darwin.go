package credstore

import "github.com/docker/docker-credential-helpers/osxkeychain"

var nativeStore = &osxkeychain.Osxkeychain{}
